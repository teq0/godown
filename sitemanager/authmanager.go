package sitemanager

import (
	"net/http"
)

func AddAuthHeaders(req *http.Request) {

	// TODO - look up any auth settings based on the host

	if req.Host == "secure.disney.com" {
		req.SetBasicAuth("chronic", "AllYourB4s3")
	}
}
