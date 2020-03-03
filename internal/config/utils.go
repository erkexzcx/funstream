package config

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

func download(link string) ([]byte, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(link + " returned HTTP code " + strconv.Itoa(resp.StatusCode))
	}

	return ioutil.ReadAll(resp.Body)
}

func downloadString(link string) (string, error) {
	contents, err := download(link)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Normalizes strings, such as channel titles, category titles etc.
func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	parts := strings.Fields(s)
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) <= 3 {
			parts[i] = strings.ToTitle(parts[i]) // All uppercase
		} else {
			parts[i] = strings.Title(parts[i])
		}
		if strings.HasSuffix(parts[i], "tv") {
			parts[i] = strings.TrimSuffix(parts[i], "tv") + "TV"
		}
	}
	return strings.Join(parts, " ")
}

// Same as downloadAsString, but also reads from local files
func retrieveContents(path string) (string, error) {
	if fileExists(path) {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(contents), nil
	}
	return downloadString(path)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
