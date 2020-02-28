package proxy

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func handleStream(w http.ResponseWriter, r *http.Request, link string, c *Channel, l *Link) {
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
		handleStream(w, r, newLink, c, l)
	}

	resp, err := getResponse(link, m3U8Timeout)
	if err != nil {
		log.Println("Failed to request link. Cycling and retrying...", err, link)
		cycleAndRetry()
		return
	}

	handleEstablishedStream(w, r, resp)
}

func handleEstablishedStream(w http.ResponseWriter, r *http.Request, resp *http.Response) {
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, "; "))
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
