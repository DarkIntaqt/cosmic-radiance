package schema

import (
	"strings"
)

/*
Validates the path syntax of an incoming request
*/
func NewPathSyntax(path string) (*Syntax, error) {

	// remove leading slash, if set
	path = strings.TrimPrefix(path, "/")

	// split string in between the platform and the method
	split := strings.SplitN(path, "/", 2)
	if len(split) != 2 {
		return nil, &InvalidPathSyntaxError{path: path}
	}

	platform := split[0]
	method := split[1]

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
