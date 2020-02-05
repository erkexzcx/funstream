package proxy

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Mutex is locked when working in this function!
func handleLinkUnknown(w http.ResponseWriter, r *http.Request, title *string, link string, c *Channel, l *Link) {
	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.cycleLinkNoMux()
		if !res {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no working channels"))
			return
		}
		l = c.ActiveLink
		newLink := l.Link
		handleLinkUnknown(w, r, title, newLink, c, l)
	}

	// We don't know what to expect, so just load URL and check content type of response
	resp, err := getResponse(link, -1)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	contentType := resp.Header.Get("Content-Type")
	l.LinkType = getLinkType(contentType)

	switch l.LinkType {
	case linkTypeUnsupported:
		resp.Body.Close()
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unsupported channel format"))
	case linkTypeM3U8:
		log.Println("Processing type: M3U8")
		defer resp.Body.Close()

		// Create new M3u8 type channel
		m3u8c := &M3U8Channel{Channel: c}
		c.ActiveLink.M3U8C = m3u8c

		m3u8c.link = resp.Request.URL.String()
		m3u8c.linkRoot = deleteAfterLastSlash(m3u8c.link)

		prefix := "http://" + r.Host + "/iptv/" + url.PathEscape(*title) + "/"
		content := rewriteLinks(bufio.NewScanner(resp.Body), prefix, m3u8c.linkRoot)

		for k, v := range resp.Header {
			w.Header().Set(k, strings.Join(v, "; "))
		}
		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, content)
	case linkTypeMedia:
		log.Println("Processing type: Media")
		handleEstablishedStream(w, r, resp)
	case linkTypeStream:
		log.Println("Processing type: Stream")
		handleEstablishedStream(w, r, resp)
	}
}
