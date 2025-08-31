package schema

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenAPISummary struct {
	Paths map[string]struct {
		Platforms []string `json:"x-platforms-available,omitempty"`
	} `json:"paths"`
}

type method struct {
	Method string
	Id     string
}

type allowedPattern map[string][]method

var AllowedPattern = getAllowedPattern()

/*
Parses internal patterns into allowedPattern which represents all available methods
*/
func getAllowedPattern() allowedPattern {

	req, err := http.Get("https://www.mingweisamuel.com/riotapi-schema/openapi-3.0.0.min.json")
	if err != nil {
		panic("Failed to fetch OpenAPI spec: " + err.Error())
	}

	defer req.Body.Close()

	if req.StatusCode != http.StatusOK {
		panic("Failed to fetch OpenAPI spec: " + req.Status)
	}

	var openAPISummary OpenAPISummary
	if err := json.NewDecoder(req.Body).Decode(&openAPISummary); err != nil {
		panic("Failed to decode OpenAPI spec: " + err.Error())
	}

	allowedPatterns := allowedPattern{}
	for path, details := range openAPISummary.Paths {
		for _, platform := range details.Platforms {

			hash := md5.Sum([]byte(path + platform))
			id := fmt.Sprintf("%x", hash)

			allowedPatterns[platform] = append(allowedPatterns[platform], method{
				Method: strings.TrimPrefix(path, "/"),
				Id:     id,
			})

		}
	}

	return allowedPatterns

}
