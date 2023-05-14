// Package odrvcookie can fetch authentication cookies for a sharepoint webdav endpoint
package odrvcookie

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"golang.org/x/net/publicsuffix"
)

// CookieAuth hold the authentication information
// These are username and password as well as the authentication endpoint
type CookieAuth struct {
	user     string
	pass     string
	endpoint string
}

// CookieResponse contains the requested cookies
type CookieResponse struct {
	RtFa    http.Cookie
	FedAuth http.Cookie
}

// SuccessResponse hold a response from the sharepoint webdav
type SuccessResponse struct {
	XMLName xml.Name            `xml:"Envelope"`
	Succ    SuccessResponseBody `xml:"Body"`
}

// SuccessResponseBody is the body of a success response, it holds the token
type SuccessResponseBody struct {
	XMLName xml.Name
	Type    string    `xml:"RequestSecurityTokenResponse>TokenType"`
	Created time.Time `xml:"RequestSecurityTokenResponse>Lifetime>Created"`
	Expires time.Time `xml:"RequestSecurityTokenResponse>Lifetime>Expires"`
	Token   string    `xml:"RequestSecurityTokenResponse>RequestedSecurityToken>BinarySecurityToken"`
}

// reqString is a template that gets populated with the user data in order to retrieve a "BinarySecurityToken"
const reqString = `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"
xmlns:a="http://www.w3.org/2005/08/addressing"
xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
<s:Header>
<a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/02/trust/RST/Issue</a:Action>
<a:ReplyTo>
<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
</a:ReplyTo>
<a:To s:mustUnderstand="1">{{ .LoginUrl }}</a:To>
<o:Security s:mustUnderstand="1"
 xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
<o:UsernameToken>
  <o:Username>{{ .Username }}</o:Username>
  <o:Password>{{ .Password }}</o:Password>
</o:UsernameToken>
</o:Security>
</s:Header>
<s:Body>
<t:RequestSecurityToken xmlns:t="http://schemas.xmlsoap.org/ws/2005/02/trust">
<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
  <a:EndpointReference>
    <a:Address>{{ .Address }}</a:Address>
  </a:EndpointReference>
</wsp:AppliesTo>
<t:KeyType>http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey</t:KeyType>
<t:RequestType>http://schemas.xmlsoap.org/ws/2005/02/trust/Issue</t:RequestType>
<t:TokenType>urn:oasis:names:tc:SAML:1.0:assertion</t:TokenType>
</t:RequestSecurityToken>
</s:Body>
</s:Envelope>`

// New creates a new CookieAuth struct
func New(pUser, pPass, pEndpoint string) CookieAuth {
	retStruct := CookieAuth{
		user:     pUser,
		pass:     pPass,
		endpoint: pEndpoint,
	}

	return retStruct
}

// Cookies creates a CookieResponse. It fetches the auth token and then
// retrieves the Cookies
func (ca *CookieAuth) Cookies() (CookieResponse, error) {
	spToken, err := ca.getSPToken()
	if err != nil {
		return CookieResponse{}, err
	}
	return ca.getSPCookie(spToken)
}

func (ca *CookieAuth) getSPCookie(conf *SuccessResponse) (CookieResponse, error) {
	spRoot, err := url.Parse(ca.endpoint)
	if err != nil {
		return CookieResponse{}, err
	}

	u, err := url.Parse("https://" + spRoot.Host + "/_forms/default.aspx?wa=wsignin1.0")
	if err != nil {
		return CookieResponse{}, err
	}

	// To authenticate with davfs or anything else we need two cookies (rtFa and FedAuth)
	// In order to get them we use the token we got earlier and a cookieJar
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return CookieResponse{}, err
	}

	client := &http.Client{
		Jar: jar,
	}

	// Send the previously acquired Token as a Post parameter
	if _, err = client.Post(u.String(), "text/xml", strings.NewReader(conf.Succ.Token)); err != nil {
		return CookieResponse{}, err
	}

	cookieResponse := CookieResponse{}
	for _, cookie := range jar.Cookies(u) {
		if (cookie.Name == "rtFa") || (cookie.Name == "FedAuth") {
			switch cookie.Name {
			case "rtFa":
				cookieResponse.RtFa = *cookie
			case "FedAuth":
				cookieResponse.FedAuth = *cookie
			}
		}
	}
	return cookieResponse, err
}

var loginUrlsMap = map[string]string{
	"com": "https://login.microsoftonline.com",
	"cn":  "https://login.chinacloudapi.cn",
	"us":  "https://login.microsoftonline.us",
	"de":  "https://login.microsoftonline.de",
}

func getLoginUrl(endpoint string) (string, error) {
	spRoot, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	domains := strings.Split(spRoot.Host, ".")
	tld := domains[len(domains)-1]
	loginUrl, ok := loginUrlsMap[tld]
	if !ok {
		return "", fmt.Errorf("tld %s is not supported", tld)
	}
	return loginUrl + "/extSTS.srf", nil
}

func (ca *CookieAuth) getSPToken() (*SuccessResponse, error) {
	loginUrl, err := getLoginUrl(ca.endpoint)
	if err != nil {
		return nil, err
	}
	reqData := map[string]string{
		"Username": ca.user,
		"Password": ca.pass,
		"Address":  ca.endpoint,
		"LoginUrl": loginUrl,
	}

	t := template.Must(template.New("authXML").Parse(reqString))

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, reqData); err != nil {
		return nil, err
	}

	// Execute the first request which gives us an auth token for the sharepoint service
	// With this token we can authenticate on the login page and save the returned cookies
	req, err := http.NewRequest("POST", loginUrl, buf)
	if err != nil {
		return nil, err
	}

	client := base.HttpClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBuf := bytes.Buffer{}
	respBuf.ReadFrom(resp.Body)
	s := respBuf.Bytes()

	var conf SuccessResponse
	err = xml.Unmarshal(s, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, err
}
