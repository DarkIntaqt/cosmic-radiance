package ratelimiter

import (
	"net/http"
	"net/url"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
)

func (rl *RateLimiter) riotApiRequest(region string, method string, queryParams url.Values, keyId int) (*http.Response, error) {

	// prepare the request
	// append the api key as a header

	// build uri with region, method, and query parameters
	uri := "https://" + region + ".api.riotgames.com/" + method + "?" + queryParams.Encode()

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Riot-Token", configs.ApiKeys[keyId])
	req.Header.Set("Accept-Encoding", "gzip") // accept gzip

	resp, err := rl.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
