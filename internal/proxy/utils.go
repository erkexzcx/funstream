package proxy

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

func download(link string) (content []byte, contentType string, err error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
	}

	contentType = resp.Header.Get("Content-Type")
	content, err = ioutil.ReadAll(resp.Body)
	return
}

func downloadString(link string) (content string, contentType string, err error) {
	var contentBytes []byte
	contentBytes, contentType, err = download(link)
	if err != nil {
		return "", "", err
	}
	return string(contentBytes), contentType, nil
}

func getResponse(link string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{}
	if timeout <= 0 {
		client.Timeout = 10 * time.Second
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
	}

	return resp, nil
}
