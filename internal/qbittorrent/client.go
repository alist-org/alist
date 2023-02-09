package qbittorrent

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client interface {
	AddFromLink(link string, savePath string, id string) error
}

type client struct {
	url    *url.URL
	client http.Client
	Client
}

func New(webuiUrl string) (*client, error) {
	u, err := url.Parse(webuiUrl)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	var c = &client{
		url:    u,
		client: http.Client{Jar: jar},
	}
	return c, nil
}

func (c *client) authorized() (bool, error) {
	resp, err := c.post("/api/v2/app/version", nil)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == 200, nil // the status code will be 403 if not authorized
}

func (c *client) login() error {
	// prepare HTTP request
	v := url.Values{}
	v.Set("username", c.url.User.Username())
	passwd, _ := c.url.User.Password()
	v.Set("password", passwd)
	resp, err := c.post("/api/v2/auth/login", v)
	if err != nil {
		return err
	}

	// check result
	body := make([]byte, 2)
	_, err = resp.Body.Read(body)
	if err != nil {
		return err
	}
	if string(body) != "Ok" {
		return errors.New("failed to login into qBittorrent webui with url: " + c.url.String())
	}
	return nil
}

func (c *client) post(path string, data url.Values) (*http.Response, error) {
	u := c.url.JoinPath(path)
	u.User = nil // remove userinfo for requests

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Cookies() != nil {
		c.client.Jar.SetCookies(u, resp.Cookies())
	}
	return resp, nil
}

func (c *client) AddFromLink(link string, savePath string, id string) error {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	var err error
	addField := func(name string, value string) {
		if err != nil {
			return
		}
		err = writer.WriteField(name, value)
	}
	addField("urls", link)
	addField("savepath", savePath)
	addField("tags", "alist-"+id)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	u := c.url.JoinPath("/api/v2/torrents/add")
	u.User = nil // remove userinfo for requests
	req, err := http.NewRequest("POST", u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	// check result
	body := make([]byte, 2)
	_, err = resp.Body.Read(body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 || string(body) != "Ok" {
		return errors.New("failed to add qBittorrent task: " + link)
	}
	return nil
}
