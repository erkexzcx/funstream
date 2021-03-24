package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handleContentM3U8(w http.ResponseWriter, r *http.Request, cr *ContentRequest) {
	var link string
	if cr.Suffix == "" {
		link = cr.Channel.ActiveLink.M3u8Ref.link
	} else {
		link = cr.Channel.ActiveLink.M3u8Ref.linkRoot + cr.Suffix
	}

	resp, err := response(link)
	if err != nil {
		log.Println("Link request failed. Trying next one...", err, cr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, cr)
		return
	}
	defer resp.Body.Close()

	handleEstablishedContentM3U8(w, r, cr, resp, link)
}

func handleEstablishedContentM3U8(w http.ResponseWriter, r *http.Request, cr *ContentRequest, resp *http.Response, link string) {
	contentTypeOrig := resp.Header.Get("Content-Type")
	contentType := strings.ToLower(contentTypeOrig)

	prefix := "http://" + r.Host + "/iptv/" + url.PathEscape(cr.Title) + "/"

	switch {
	case contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl": // M3U8 metadata
		content := rewriteLinks(&resp.Body, prefix, cr.Channel.ActiveLink.M3u8Ref.linkRoot)
		addHeaders(resp.Header, w.Header(), false)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	default: // media (or anything else)
		handleEstablishedContentMedia(w, r, cr, resp)
	}
}
