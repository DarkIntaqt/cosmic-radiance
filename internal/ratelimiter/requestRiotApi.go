package ratelimiter

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

func (rl *RateLimiter) riotApiRequest(region string, method string, queryParams url.Values, keyId int) (*http.Response, error) {
	// prepare the request
	// append the api key as a header

	// build uri with region, method, and query parameters
	start := time.Now()
	uri := "https://" + region + ".api.riotgames.com/" + method
	if params := queryParams.Encode(); params != "" {
		uri += "?" + params
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", rl.opts.UserAgent)
	req.Header.Set("X-Riot-Token", rl.opts.ApiKeys[keyId].ApiKey)
	req.Header.Set("Accept-Encoding", "gzip") // accept gzip

	resp, err := rl.client.Do(req)
	if err != nil {
		return nil, err
	}

	slog.Debug("Request sent to Riot Games API", "uri", uri, "keyId", keyId, "statusCode", resp.StatusCode, "duration", time.Since(start))

	return resp, nil
}
