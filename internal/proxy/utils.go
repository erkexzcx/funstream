package proxy

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
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

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		contentType = resp.Header.Get("Content-Type")
		content, err = ioutil.ReadAll(resp.Body)
		return
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		linkURL, err := url.Parse(link)
		if err != nil {
			return nil, "", errors.New("Unknown error occurred")
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			return nil, "", errors.New("Unknown error occurred")
		}
		newLink := linkURL.ResolveReference(redirectURL)
		return download(newLink.String())
	}

	return nil, "", errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
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
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
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

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		linkURL, err := url.Parse(link)
		if err != nil {
			return nil, errors.New("Unknown error occurred")
		}
		redirectURL, err := url.Parse(resp.Header.Get("Location"))
		if err != nil {
			return nil, errors.New("Unknown error occurred")
		}
		newLink := linkURL.ResolveReference(redirectURL)
		return getResponse(newLink.String(), timeout)
	}

	return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))

}
