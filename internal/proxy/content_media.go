package proxy

import (
	"io"
	"log"
	"net/http"
)

func handleContentMedia(w http.ResponseWriter, r *http.Request, sr *StreamRequest) {
	resp, err := getResponse(sr.Channel.ActiveLink.Link)
	if err != nil {
		log.Println("Link request failed. Trying next one...", err, sr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, sr)
		return
	}
	defer resp.Body.Close()

	handleEstablishedContentMedia(w, r, sr, resp)
}

func handleEstablishedContentMedia(w http.ResponseWriter, r *http.Request, sr *StreamRequest, resp *http.Response) {
	sr.Channel.LinksMux.Unlock() // So other clients can watch it too
	addHeaders(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	sr.Channel.LinksMux.Lock() // To prevent runtime error because we use 'defer' to unlock mux
}
