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
	return content, resp.Header.Get("User-Agent"), err
}

func getResponse(link string) (*http.Response, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		defer resp.Body.Close()
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
