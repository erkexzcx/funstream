package proxy

import (
	"io"
	"log"
	"net/http"
)

func handleContentMedia(w http.ResponseWriter, r *http.Request, cr *ContentRequest) {
	resp, err := response(cr.Channel.ActiveLink.Link)
	if err != nil {
		log.Println("Link request failed. Trying next one...", err, cr.Channel.ActiveLink.Link)
		cycleAndRetry(w, r, cr)
		return
	}
	defer resp.Body.Close()

	handleEstablishedContentMedia(w, r, cr, resp)
}

func handleEstablishedContentMedia(w http.ResponseWriter, r *http.Request, cr *ContentRequest, resp *http.Response) {
	cr.Channel.LinksMux.Unlock() // So other clients can watch it too
	addHeaders(resp.Header, w.Header(), true)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	cr.Channel.LinksMux.Lock() // To prevent runtime error because we use 'defer' to unlock mux
}
