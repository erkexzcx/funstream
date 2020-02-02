package proxy

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
)

func handleM3U8Channel(w http.ResponseWriter, r *http.Request, escapedTitle, unescapedTitle *string, link string, c *M3U8Channel, l *Link) {
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
		handleM3U8Channel(w, r, escapedTitle, unescapedTitle, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
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
	prefix := "http://" + r.Host + "/iptv/" + *escapedTitle + "/"
	content := rewriteLinks(bufio.NewScanner(resp.Body), prefix, linkRoot)

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Write([]byte(content))
}

func handleM3U8ChannelData(w http.ResponseWriter, r *http.Request, escapedTitle, unescapedTitle *string, link string, c *M3U8Channel, l *Link) {
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
		handleM3U8ChannelData(w, r, escapedTitle, unescapedTitle, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	defer resp.Body.Close()
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	// Find content type
	contentTypeOrig := resp.Header.Get("Content-Type")
	contentType := strings.ToLower(contentTypeOrig)

	if (contentType == "application/vnd.apple.mpegurl" || contentType == "application/x-mpegurl") && strings.Contains(link, ".m3u8") {
		// If we reach this code block - it means we got redirect without HTTP code 3**
		c.newRedirectedLink(link)

		linkRoot := c.LinkRoot()
		prefix := "http://" + r.Host + "/iptv/" + *escapedTitle + "/"
		content := rewriteLinks(bufio.NewScanner(resp.Body), prefix, linkRoot)

		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
	} else if strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		// TS files
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", contentTypeOrig)
		io.Copy(w, resp.Body)
	}

}
