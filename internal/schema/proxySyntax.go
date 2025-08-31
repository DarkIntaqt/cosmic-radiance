package schema

import (
	"strings"
)

/*
Validates the proxy syntax of an incoming request
*/
func NewProxySyntax(host string, path string) (*Syntax, error) {

	platform := strings.SplitN(host, ".", 2)[0]

	// remove leading slash, if set
	method := strings.TrimPrefix(path, "/")

	// check if the platform exist, if so, receive all available methods for that platform
	methods, Ok := AllowedPattern[platform]
	if !Ok {
		return nil, &InvalidPathSyntaxError{path: path}
	}

	// find the closest matching method
	id := ""
	endpoint := ""
	for _, endpointPattern := range methods {
		patternSegments := strings.Split(endpointPattern.Method, "/")
		methodSegments := strings.Split(method, "/")

		if len(patternSegments) != len(methodSegments) {
			continue
		}
		match := true
		for i := range patternSegments {
			if strings.HasPrefix(patternSegments[i], "{") && strings.HasSuffix(patternSegments[i], "}") {
				continue // treat as wildcard, technically not right but good enough
			}
			if patternSegments[i] != methodSegments[i] {
				match = false
				break
			}
		}
		if match {
			id = endpointPattern.Id
			endpoint = endpointPattern.Method
			break
		}
	}

	if id == "" {
		return nil, &InvalidPathSyntaxError{path: path}
	}

	return &Syntax{
		Platform: platform,
		Method:   method,
		Id:       id,
		Endpoint: endpoint,
	}, nil
}
