package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func logoHandler(ctx *fasthttp.RequestCtx) {
	title := strings.Replace(string(ctx.RequestURI()), "/logo/", "", 1)
	unescapedTitle, err := url.QueryUnescape(title)
	if err != nil {
		ctx.Error("Invalid request", http.StatusBadRequest)
		return
	}

	// Find channel reference
	channel, ok := playlist.Channels[unescapedTitle]
	if !ok {
		ctx.Error("Channel not found", http.StatusNotFound)
		return
	}

	// Find real URL of logo
	channel.LogoCacheMux.Lock()
	defer channel.LogoCacheMux.Unlock()
	if len(channel.LogoCache) == 0 {
		img, contentType, err := downloadAsBytes(channel.Logo)
		if err != nil {
			ctx.Error("Unable to serve logo", http.StatusInternalServerError)
			return
		}
		channel.LogoCache = img
		channel.LogoCacheContentType = contentType
		ctx.SetContentTypeBytes(contentType)
		ctx.SetBody(img)
	} else {
		ctx.SetContentTypeBytes(channel.LogoCacheContentType)
		ctx.SetBody(channel.LogoCache)
	}
}
