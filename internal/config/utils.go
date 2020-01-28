package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

func downloadAsBytes(u string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(u)
	req.Header.Add("User-Agent", userAgent)

	if err := fasthttp.DoTimeout(req, resp, 10*time.Second); err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode()
	if statusCode == 200 {
		// Success
		return resp.Body(), nil
	} else if statusCode >= 300 && statusCode < 400 {
		// Redirection
		myURL, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		nextURL, err := url.Parse(string(resp.Header.Peek("Location")))
		if err != nil {
			return nil, err
		}
		newURL := myURL.ResolveReference(nextURL)
		return downloadAsBytes(newURL.String())
	}
	return nil, errors.New(u + " returned HTTP code " + strconv.Itoa(statusCode))
}

func downloadAsString(u string) (string, error) {
	contents, err := downloadAsBytes(u)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Normalizes strings, such as channel titles, category titles etc.
func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	parts := strings.Fields(s)
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) <= 3 {
			parts[i] = strings.ToUpper(parts[i])
		} else if parts[i] == "hd" {
			parts[i] = strings.ToUpper(parts[i])
		} else if parts[i][:2] == "tv" {
			parts[i] = strings.ToUpper(parts[i][:2]) + parts[i][2:]
		} else {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, " ")
}
