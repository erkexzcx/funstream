package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handleContentM3U8(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	var link string
	if sr.Suffix == "" {
		link = sr.Channel.ActiveLink.M3u8Ref.link
	} else {
		link = sr.Channel.ActiveLink.M3u8Ref.linkRoot + sr.Suffix
	}

	resp, err := getResponse(link)
	if err != nil {
		log.Println("Link rquest failed. Trying next one...", err, sr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, sr)
		return
	}
	defer resp.Body.Close()

	handleEstablishedContentM3U8(w, r, sr, resp, link)
}

func handleEstablishedContentM3U8(w http.ResponseWriter, r *http.Request, sr *StreamRequest, resp *http.Response, link string) {
	contentTypeOrig := resp.Header.Get("Content-Type")
	contentType := strings.ToLower(contentTypeOrig)

	prefix := "http://" + r.Host + "/iptv/" + url.PathEscape(sr.Title) + "/"

	switch {
	case contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl": // M3U8 metadata
		// Update links in case of redirect
		link := resp.Request.URL.String()
		sr.Channel.ActiveLink.M3u8Ref.newRedirectedLink(link)

		content := rewriteLinks(&resp.Body, prefix, sr.Channel.ActiveLink.M3u8Ref.linkRoot)
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	default: // media (or anything else)
		handleEstablishedContentMedia(w, r, sr, resp)
	}
}
