package httptool

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var Header map[string]string

// PostJSON .
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
	client := &http.Client{Timeout: timeout}
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
