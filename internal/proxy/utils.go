package proxy

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func cycleAndRetry(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	if !sr.Channel.cycleLink() {
		http.Error(w, "no working channels", http.StatusInternalServerError)
		return
	}
	streamRequestHandler(w, r, sr)
}

// StreamRequest represents HTTP request that is received from the user
type StreamRequest struct {
	Title   string
	Suffix  string
	Channel *Channel
}

func getStreamRequest(w http.ResponseWriter, r *http.Request, prefix string) (*StreamRequest, error) {
	reqPath := strings.Replace(r.URL.RequestURI(), prefix, "", 1)
	reqPathParts := strings.SplitN(reqPath, "/", 2)
	if len(reqPathParts) == 0 {
		return nil, errors.New("Bad request")
	}

	// Unescape channel title
	var err error
	reqPathParts[0], err = url.PathUnescape(reqPathParts[0])
	if err != nil {
		return nil, errors.New("Bad request")
	}

	// Find channel reference
	channel, ok := playlist.Channels[reqPathParts[0]]
	if !ok {
		return nil, errors.New("Bad request")
	}

	if len(reqPathParts) == 1 {
		return &StreamRequest{reqPathParts[0], "", channel}, nil
	}
	return &StreamRequest{reqPathParts[0], reqPathParts[1], channel}, nil
}

func downloadString(link string) (content string, contentType string, err error) {
	contentBytes, contentType, err := download(link)
	if err != nil {
		return "", "", err
	}
	return string(contentBytes), contentType, nil
}

func download(link string) (content []byte, contentType string, err error) {
	resp, err := getResponse(link)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	content, err = ioutil.ReadAll(resp.Body)
	return content, resp.Header.Get("Content-Type"), err
}

// HTTP client that does not follow redirects
// It automatically adds "Referrerr" header which causes
// 404 errors on some backends.
var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func getResponse(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
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
		return getResponse(newLink.String())
	}

	return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
}

func addHeaders(from, to http.Header, contentLength bool) {
	for k, v := range from {
		switch k {
		case "Connection":
			from.Set("Connection", strings.Join(v, "; "))
		case "Content-Type":
			from.Set("Content-Type", strings.Join(v, "; "))
		case "Transfer-Encoding":
			from.Set("Transfer-Encoding", strings.Join(v, "; "))
		case "Cache-Control":
			from.Set("Cache-Control", strings.Join(v, "; "))
		case "Date":
			from.Set("Date", strings.Join(v, "; "))
		case "Content-Length":
			if contentLength {
				from.Set("Content-Length", strings.Join(v, "; "))
			}
		}
	}
}
