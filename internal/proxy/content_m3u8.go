package proxy

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func handleContentM3U8(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	var link string
	if sr.Suffix == "" {
		link = sr.Channel.ActiveLink.M3u8Ref.link

		if sr.Channel.ActiveLink.M3u8Ref.cacheValid() {
			w.Header().Set("Content-Type", "application/x-mpegURL")
			w.WriteHeader(http.StatusOK)
			w.Write(sr.Channel.ActiveLink.M3u8Ref.linkCache)
			return
		}
	} else {
		link = sr.Channel.ActiveLink.M3u8Ref.linkRoot + sr.Suffix

		if cache, found := m3u8cache.Get(r.URL.RequestURI()); found {
			mce := cache.(*M3U8CacheElem)
			w.Header().Set("Content-Type", *mce.contentType)
			w.WriteHeader(http.StatusOK)
			w.Write(*mce.content)
			return
		}
	}

	resp, err := getResponse(link)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...", err, sr.Channel.ActiveLink.Link)
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
	case contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl": // media (or anything else)
		// Update links in case of redirect
		link := resp.Request.URL.String()
		sr.Channel.ActiveLink.M3u8Ref.newRedirectedLink(link)

		content := rewriteLinks(&resp.Body, prefix, sr.Channel.ActiveLink.M3u8Ref.linkRoot)
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)

		sr.Channel.ActiveLink.M3u8Ref.linkCache = []byte(content)
		sr.Channel.ActiveLink.M3u8Ref.linkCreatedAt = time.Now()
	default: // media (or anything else)
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "failed to fetch media file", http.StatusInternalServerError)
		}

		for k, v := range resp.Header {
			w.Header().Set(k, strings.Join(v, "; "))
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(content)

		m3u8cache.SetDefault(r.URL.RequestURI(), &M3U8CacheElem{&content, &contentType})
	}
}
