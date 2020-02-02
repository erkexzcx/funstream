package proxy

import (
	"io"
	"log"
	"net/http"
)

func handleStream(w http.ResponseWriter, r *http.Request, escapedTitle, unescapedTitle *string, link string, c *Channel, l *Link) {
	cycleAndRetry := func() {
		// Link is not working - try to switch to the next one and reload
		res := c.cycleLink()
		if !res {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no working channels"))
			return
		}
		l = c.ActiveLink
		newLink := l.Link
		handleStream(w, r, escapedTitle, unescapedTitle, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...")
		cycleAndRetry()
		return
	}

	handleEstablishedStream(w, r, resp)
}

func handleEstablishedStream(w http.ResponseWriter, r *http.Request, resp *http.Response) {
	defer resp.Body.Close()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(w, resp.Body)
}
