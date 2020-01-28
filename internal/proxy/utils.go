package proxy

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

func downloadAsBytes(u string) (content []byte, contentType []byte, err error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(u)
	req.Header.Add("User-Agent", userAgent)

	if err := fasthttp.DoTimeout(req, resp, 10*time.Second); err != nil {
		return nil, nil, err
	}

	statusCode := resp.StatusCode()
	if statusCode == 200 {
		// Success
		return resp.Body(), resp.Header.ContentType(), nil
	} else if statusCode >= 300 && statusCode < 400 {
		// Redirection
		myURL, err := url.Parse(u)
		if err != nil {
			return nil, nil, err
		}
		nextURL, err := url.Parse(string(resp.Header.Peek("Location")))
		if err != nil {
			return nil, nil, err
		}
		newURL := myURL.ResolveReference(nextURL)
		return downloadAsBytes(newURL.String())
	}
	return nil, nil, errors.New(u + " returned HTTP code " + strconv.Itoa(statusCode))
}

func downloadAsString(u string) (content string, contentType []byte, err error) {
	contents, contentType, err := downloadAsBytes(u)
	if err != nil {
		return "", nil, err
	}
	return string(contents), contentType, nil
}

func getRequest(link string, timeout time.Duration) (req *fasthttp.Request, resp *fasthttp.Response, err error) {
	resp = fasthttp.AcquireResponse()
	req = fasthttp.AcquireRequest()

	req.SetRequestURI(link)
	req.Header.Add("User-Agent", userAgent)

	if timeout <= 0 {
		err = fasthttp.Do(req, resp)
		return
	}
	err = fasthttp.DoTimeout(req, resp, timeout)
	return
}

func getRequestStandard(link string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{}
	if timeout <= 0 {
		client.Timeout = 10 * time.Second
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}
