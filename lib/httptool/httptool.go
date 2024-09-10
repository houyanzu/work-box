package httptool

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var Header map[string]string

// PostJSON .
// Deprecated: Use boxhttp.PostJSON instead.
func PostJSON(url string, js []byte) ([]byte, int, error) {
	defer func() {
		Header = nil
	}()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if Header != nil {
		for k, v := range Header {
			req.Header.Set(k, v)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {

		}
	}()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

// PostForm .
// Deprecated: Use boxhttp.PostForm instead.
func PostForm(url string, form url.Values) ([]byte, int, error) {
	defer func() {
		Header = nil
	}()
	req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if Header != nil {
		for k, v := range Header {
			req.Header.Set(k, v)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {

		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}

// Get .
// Deprecated: Use boxhttp.Get instead.
func Get(url string, timeout time.Duration) ([]byte, int, error) {
	defer func() {
		Header = nil
	}()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	if Header != nil {
		for k, v := range Header {
			req.Header.Set(k, v)
		}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: timeout, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {

		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode, nil
}
