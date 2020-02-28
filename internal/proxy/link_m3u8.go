package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handleM3U8Channel(w http.ResponseWriter, r *http.Request, title *string, link string, c *M3U8Channel, l *Link) {
	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.Channel.cycleLink()
		if !res {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no working channels"))
			return
		}
		l = c.Channel.ActiveLink
		newLink := l.Link
		handleM3U8Channel(w, r, title, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...", err, link)
		cycleAndRetry()
		return
	}
	defer resp.Body.Close()

	// In case we got redirect - update channel's links
	if link != resp.Request.URL.String() {
		link = resp.Request.URL.String()
		c.newRedirectedLink(link)
	}

	linkRoot := c.LinkRoot()
	prefix := "http://" + r.Host + "/iptv/" + url.PathEscape(*title) + "/"
	content := rewriteLinks(bufio.NewScanner(resp.Body), prefix, linkRoot)

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, "; "))
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, content)
}

func handleM3U8ChannelData(w http.ResponseWriter, r *http.Request, title *string, link string, c *M3U8Channel, l *Link) {
	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.Channel.cycleLink()
		if !res {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no working channels"))
			return
		}
		l = c.Channel.ActiveLink
		newLink := l.Link
		handleM3U8ChannelData(w, r, title, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...", err, link)
		cycleAndRetry()
		return
	}
	defer resp.Body.Close()

	// Find content type
	contentTypeOrig := resp.Header.Get("Content-Type")
	contentType := strings.ToLower(contentTypeOrig)

	if contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl" {
		// If we reach this code block - it means we got redirect without HTTP code 3**
		c.newRedirectedLink(link)

		linkRoot := c.LinkRoot()
		prefix := "http://" + r.Host + "/iptv/" + url.PathEscape(*title) + "/"
		content := rewriteLinks(bufio.NewScanner(resp.Body), prefix, linkRoot)

		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	} else {
		// video/audio/text(subtitles) or anything else...
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header().Set(k, strings.Join(v, "; "))
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}

}
