package proxy

import (
	"bufio"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

var m3u8channels map[string]*M3U8Channel
var m3u8cache = cache.New(time.Minute, 10*time.Second) // 1 minute expiration, clean every 10 sec

// M3U8CacheElem represents M3U8 any media file cache
type M3U8CacheElem struct {
	content     *[]byte
	contentType *string
}

// M3U8Channel stores information about m3u8 channel
type M3U8Channel struct {
	Channel       *Channel
	link          string
	linkCache     []byte
	linkCreatedAt time.Time
	linkRoot      string
}

func (c *M3U8Channel) newRedirectedLink(s string) {
	c.link = s
	c.linkRoot = deleteAfterLastSlash(s)
}

func (c *M3U8Channel) cacheValid() bool {
	if c.linkCreatedAt.IsZero() || time.Now().Sub(c.linkCreatedAt).Seconds() > 2 {
		return false
	}
	return true
}

func deleteAfterLastSlash(str string) string {
	return str[0 : strings.LastIndex(str, "/")+1]
}

var reURILinkExtract = regexp.MustCompile(`URI="([^"]*)"`)

func rewriteLinks(rbody *io.ReadCloser, prefix, linkRoot string) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(*rbody)
	linkRootURL, _ := url.Parse(linkRoot) // It will act as a base URL for full URLs

	modifyLink := func(link string) string {
		var l string

		switch {
		case strings.HasPrefix(link, "//"):
			tmpURL, _ := url.Parse(link)
			tmp2URL, _ := url.Parse(tmpURL.RequestURI())
			link = (linkRootURL.ResolveReference(tmp2URL)).String()
			l = strings.ReplaceAll(link, linkRoot, "")
		case strings.HasPrefix(link, "/"):
			tmp2URL, _ := url.Parse(link)
			link = (linkRootURL.ResolveReference(tmp2URL)).String()
			l = strings.ReplaceAll(link, linkRoot, "")
		default:
			l = link
		}

		return prefix + l
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			line = modifyLink(line)
		} else if strings.Contains(line, "URI=\"") && !strings.Contains(line, "URI=\"\"") {
			link := reURILinkExtract.FindStringSubmatch(line)[1]
			line = reURILinkExtract.ReplaceAllString(line, `URI="`+modifyLink(link)+`"`)
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	return sb.String()
}
