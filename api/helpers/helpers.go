package helpers

import (
	"os"
	"strings"
)

// RemoveDomainError returns true only if the url does not include the same domain as the server
func RemoveDomainError(url string) bool {
	// check if a user sumitted the same domain name as the server providing url shortening service
	if url == os.Getenv("DOMAIN") {
		return false
	}

	// clean up the url - get rid of any of "http://", "https://", "www." from the url
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)

	// check if a user summited the same domain name as the server, but with any '/' after that
	newURL = strings.Split(newURL, "/")[0]
	if newURL == os.Getenv("DOMAIN") {
		return false
	}

	return true
}

// EnforceHTTP attaches "http://" to the url submitted without "http"
func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}
