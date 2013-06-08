package octokit

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	GitHubApiUrl  string = "https://" + GitHubApiHost
	GitHubApiHost string = "api.github.com"
)

type Client struct {
	httpClient *http.Client
	Login      string
	Password   string
	Token      string
}

func (c *Client) get(path string, extraHeaders map[string]string) ([]byte, error) {
	return c.request("GET", path, extraHeaders, nil)
}

func (c *Client) post(path string, extraHeaders map[string]string, content io.Reader) ([]byte, error) {
	return c.request("POST", path, extraHeaders, content)
}

func (c *Client) jsonGet(path string, extraHeaders map[string]string, v interface{}) error {
	body, err := c.get(path, extraHeaders)
	if err != nil {
		return err
	}

	return jsonUnmarshal(body, v)
}

func (c *Client) jsonPost(path string, extraHeaders map[string]string, params interface{}, v interface{}) error {
	b, err := jsonMarshal(params)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(b)
	body, err := c.post(path, extraHeaders, buffer)
	if err != nil {
		return err
	}

	return jsonUnmarshal(body, v)
}

func (c *Client) request(method, path string, extraHeaders map[string]string, content io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", GitHubApiUrl, path)
	request, err := http.NewRequest(method, url, content)
	if err != nil {
		return nil, err
	}

	c.setDefaultHeaders(request)

	if extraHeaders != nil {
		for h, v := range extraHeaders {
			request.Header.Set(h, v)
		}
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 400 && response.StatusCode < 600 {
		return nil, handleErrors(body)
	}

	return body, nil
}

func (c *Client) setDefaultHeaders(request *http.Request) {
	request.Header.Set("Accept", "application/vnd.github.beta+json")
	if c.Token != "" {
		request.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	}
	if c.Login != "" && c.Password != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Basic %s", hashAuth(c.Login, c.Password)))
	}
}
