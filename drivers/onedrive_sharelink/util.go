package onedrive_sharelink

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

// NewNoRedirectClient creates an HTTP client that doesn't follow redirects
func NewNoRedirectCLient() *http.Client {
	return &http.Client{
		Timeout: time.Hour * 48,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.Conf.TlsInsecureSkipVerify},
		},
		// Prevent following redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// getCookiesWithPassword fetches cookies required for authenticated access using the provided password
func getCookiesWithPassword(link, password string) (string, error) {
	// Send GET request
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	// Initialize variables to store form data
	var viewstate, eventvalidation, postAction string

	// Recursive function to find input fields by their IDs
	var findInputFields func(*html.Node)
	findInputFields = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					switch attr.Val {
					case "__VIEWSTATE":
						viewstate = getAttrValue(n, "value")
					case "__EVENTVALIDATION":
						eventvalidation = getAttrValue(n, "value")
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "inputForm" {
					postAction = getAttrValue(n, "action")
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findInputFields(c)
		}
	}
	findInputFields(doc)

	// Prepare the new URL for the POST request
	linkParts, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	newURL := fmt.Sprintf("%s://%s%s", linkParts.Scheme, linkParts.Host, postAction)

	// Prepare the request body
	data := url.Values{
		"txtPassword":          []string{password},
		"__EVENTVALIDATION":    []string{eventvalidation},
		"__VIEWSTATE":          []string{viewstate},
		"__VIEWSTATEENCRYPTED": []string{""},
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// Send the POST request, preventing redirects
	resp, err = client.PostForm(newURL, data)
	if err != nil {
		return "", err
	}

	// Extract the desired cookie value
	cookie := resp.Cookies()
	var fedAuthCookie string
	for _, c := range cookie {
		if c.Name == "FedAuth" {
			fedAuthCookie = c.Value
			break
		}
	}
	if fedAuthCookie == "" {
		return "", fmt.Errorf("wrong password")
	}
	return fmt.Sprintf("FedAuth=%s;", fedAuthCookie), nil
}

// getAttrValue retrieves the value of the specified attribute from an HTML node
func getAttrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// getHeaders constructs and returns the necessary HTTP headers for accessing the OneDrive share link
func (d *OnedriveSharelink) getHeaders() (http.Header, error) {
	header := http.Header{}
	header.Set("User-Agent", base.UserAgent)
	header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")

	// Save current timestamp to d.HeaderTime
	d.HeaderTime = time.Now().Unix()

	if d.ShareLinkPassword == "" {
		// Create a no-redirect client
		clientNoDirect := NewNoRedirectCLient()
		req, err := http.NewRequest("GET", d.ShareLinkURL, nil)
		if err != nil {
			return nil, err
		}
		// Set headers for the request
		req.Header = header
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		redirectUrl := answerNoRedirect.Header.Get("Location")
		log.Debugln("redirectUrl:", redirectUrl)
		if redirectUrl == "" {
			return nil, fmt.Errorf("password protected link. Please provide password")
		}
		header.Set("Cookie", answerNoRedirect.Header.Get("Set-Cookie"))
		header.Set("Referer", redirectUrl)

		// Extract the host part of the redirect URL and set it as the authority
		u, err := url.Parse(redirectUrl)
		if err != nil {
			return nil, err
		}
		header.Set("authority", u.Host)
		return header, nil
	} else {
		cookie, err := getCookiesWithPassword(d.ShareLinkURL, d.ShareLinkPassword)
		if err != nil {
			return nil, err
		}
		header.Set("Cookie", cookie)
		header.Set("Referer", d.ShareLinkURL)
		header.Set("authority", strings.Split(strings.Split(d.ShareLinkURL, "//")[1], "/")[0])
		return header, nil
	}
}

// getFiles retrieves the files from the OneDrive share link at the specified path
func (d *OnedriveSharelink) getFiles(path string) ([]Item, error) {
	clientNoDirect := NewNoRedirectCLient()
	req, err := http.NewRequest("GET", d.ShareLinkURL, nil)
	if err != nil {
		return nil, err
	}
	header := req.Header
	redirectUrl := ""
	if d.ShareLinkPassword == "" {
		header.Set("User-Agent", base.UserAgent)
		header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
		req.Header = header
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		redirectUrl = answerNoRedirect.Header.Get("Location")
	} else {
		header = d.Headers
		req.Header = header
		answerNoRedirect, err := clientNoDirect.Do(req)
		if err != nil {
			return nil, err
		}
		redirectUrl = answerNoRedirect.Header.Get("Location")
	}
	redirectSplitURL := strings.Split(redirectUrl, "/")
	req.Header = d.Headers
	downloadLinkPrefix := ""
	rootFolderPre := ""

	// Determine the appropriate URL and root folder based on whether the link is SharePoint
	if d.IsSharepoint {
		// update req url
		req.URL, err = url.Parse(redirectUrl)
		if err != nil {
			return nil, err
		}
		// Get redirectUrl
		answer, err := clientNoDirect.Do(req)
		if err != nil {
			d.Headers, err = d.getHeaders()
			if err != nil {
				return nil, err
			}
			return d.getFiles(path)
		}
		defer answer.Body.Close()
		re := regexp.MustCompile(`templateUrl":"(.*?)"`)
		body, err := io.ReadAll(answer.Body)
		if err != nil {
			return nil, err
		}
		template := re.FindString(string(body))
		template = template[strings.Index(template, "templateUrl\":\"")+len("templateUrl\":\""):]
		template = template[:strings.Index(template, "?id=")]
		template = template[:strings.LastIndex(template, "/")]
		downloadLinkPrefix = template + "/download.aspx?UniqueId="
		params, err := url.ParseQuery(redirectUrl[strings.Index(redirectUrl, "?")+1:])
		if err != nil {
			return nil, err
		}
		rootFolderPre = params.Get("id")
	} else {
		redirectUrlCut := redirectUrl[:strings.LastIndex(redirectUrl, "/")]
		downloadLinkPrefix = redirectUrlCut + "/download.aspx?UniqueId="
		params, err := url.ParseQuery(redirectUrl[strings.Index(redirectUrl, "?")+1:])
		if err != nil {
			return nil, err
		}
		rootFolderPre = params.Get("id")
	}
	d.downloadLinkPrefix = downloadLinkPrefix
	rootFolder, err := url.QueryUnescape(rootFolderPre)
	if err != nil {
		return nil, err
	}
	log.Debugln("rootFolder:", rootFolder)
	// Extract the relative path up to and including "Documents"
	relativePath := strings.Split(rootFolder, "Documents")[0] + "Documents"

	// URL encode the relative path
	relativeUrl := url.QueryEscape(relativePath)
	// Replace underscores and hyphens in the encoded relative path
	relativeUrl = strings.Replace(relativeUrl, "_", "%5F", -1)
	relativeUrl = strings.Replace(relativeUrl, "-", "%2D", -1)

	// If the path is not the root, append the path to the root folder
	if path != "/" {
		rootFolder = rootFolder + path
	}

	// URL encode the full root folder path
	rootFolderUrl := url.QueryEscape(rootFolder)
	// Replace underscores and hyphens in the encoded root folder URL
	rootFolderUrl = strings.Replace(rootFolderUrl, "_", "%5F", -1)
	rootFolderUrl = strings.Replace(rootFolderUrl, "-", "%2D", -1)

	log.Debugln("relativePath:", relativePath, "relativeUrl:", relativeUrl, "rootFolder:", rootFolder, "rootFolderUrl:", rootFolderUrl)

	// Construct the GraphQL query with the encoded paths
	graphqlVar := fmt.Sprintf(`{"query":"query (\n        $listServerRelativeUrl: String!,$renderListDataAsStreamParameters: RenderListDataAsStreamParameters!,$renderListDataAsStreamQueryString: String!\n        )\n      {\n      \n      legacy {\n      \n      renderListDataAsStream(\n      listServerRelativeUrl: $listServerRelativeUrl,\n      parameters: $renderListDataAsStreamParameters,\n      queryString: $renderListDataAsStreamQueryString\n      )\n    }\n      \n      \n  perf {\n    executionTime\n    overheadTime\n    parsingTime\n    queryCount\n    validationTime\n    resolvers {\n      name\n      queryCount\n      resolveTime\n      waitTime\n    }\n  }\n    }","variables":{"listServerRelativeUrl":"%s","renderListDataAsStreamParameters":{"renderOptions":5707527,"allowMultipleValueFilterForTaxonomyFields":true,"addRequiredFields":true,"folderServerRelativeUrl":"%s"},"renderListDataAsStreamQueryString":"@a1=\'%s\'&RootFolder=%s&TryNewExperienceSingle=TRUE"}}`, relativePath, rootFolder, relativeUrl, rootFolderUrl)
	tempHeader := make(http.Header)
	for k, v := range d.Headers {
		tempHeader[k] = v
	}
	tempHeader["Content-Type"] = []string{"application/json;odata=verbose"}

	client := &http.Client{}
	postUrl := strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/v2.1/graphql"
	req, err = http.NewRequest("POST", postUrl, strings.NewReader(graphqlVar))
	if err != nil {
		return nil, err
	}
	req.Header = tempHeader

	resp, err := client.Do(req)
	if err != nil {
		d.Headers, err = d.getHeaders()
		if err != nil {
			return nil, err
		}
		return d.getFiles(path)
	}
	defer resp.Body.Close()
	var graphqlReq GraphQLRequest
	json.NewDecoder(resp.Body).Decode(&graphqlReq)
	log.Debugln("graphqlReq:", graphqlReq)
	filesData := graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row
	if graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.NextHref != "" {
		nextHref := graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.NextHref + "&@a1=REPLACEME&TryNewExperienceSingle=TRUE"
		nextHref = strings.Replace(nextHref, "REPLACEME", "%27"+relativeUrl+"%27", -1)
		log.Debugln("nextHref:", nextHref)
		filesData = append(filesData, graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row...)

		listViewXml := graphqlReq.Data.Legacy.RenderListDataAsStream.ViewMetadata.ListViewXml
		log.Debugln("listViewXml:", listViewXml)
		renderListDataAsStreamVar := `{"parameters":{"__metadata":{"type":"SP.RenderListDataParameters"},"RenderOptions":1216519,"ViewXml":"REPLACEME","AllowMultipleValueFilterForTaxonomyFields":true,"AddRequiredFields":true}}`
		listViewXml = strings.Replace(listViewXml, `"`, `\"`, -1)
		renderListDataAsStreamVar = strings.Replace(renderListDataAsStreamVar, "REPLACEME", listViewXml, -1)

		graphqlReqNEW := GraphQLNEWRequest{}
		postUrl = strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream" + nextHref
		req, _ = http.NewRequest("POST", postUrl, strings.NewReader(renderListDataAsStreamVar))
		req.Header = tempHeader

		resp, err := client.Do(req)
		if err != nil {
			d.Headers, err = d.getHeaders()
			if err != nil {
				return nil, err
			}
			return d.getFiles(path)
		}
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&graphqlReqNEW)
		for graphqlReqNEW.ListData.NextHref != "" {
			graphqlReqNEW = GraphQLNEWRequest{}
			postUrl = strings.Join(redirectSplitURL[:len(redirectSplitURL)-3], "/") + "/_api/web/GetListUsingPath(DecodedUrl=@a1)/RenderListDataAsStream" + nextHref
			req, _ = http.NewRequest("POST", postUrl, strings.NewReader(renderListDataAsStreamVar))
			req.Header = tempHeader
			resp, err := client.Do(req)
			if err != nil {
				d.Headers, err = d.getHeaders()
				if err != nil {
					return nil, err
				}
				return d.getFiles(path)
			}
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&graphqlReqNEW)
			nextHref = graphqlReqNEW.ListData.NextHref + "&@a1=REPLACEME&TryNewExperienceSingle=TRUE"
			nextHref = strings.Replace(nextHref, "REPLACEME", "%27"+relativeUrl+"%27", -1)
			filesData = append(filesData, graphqlReqNEW.ListData.Row...)
		}
		filesData = append(filesData, graphqlReqNEW.ListData.Row...)
	} else {
		filesData = append(filesData, graphqlReq.Data.Legacy.RenderListDataAsStream.ListData.Row...)
	}
	return filesData, nil
}
