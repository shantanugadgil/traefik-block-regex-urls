// Package "block regex urls" is a Traefik plugin to block access to certain urls using a list of regex values and return a defined status code.
package traefik_block_regex_urls

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

/**********************************
 *          Define types          *
 **********************************/

type traefik_block_regex_urls struct {
	next          http.Handler
	name          string
	regexps       []*regexp.Regexp
	matchStrings  []string
	silentStartUp bool
	statusCode    int
}

type Config struct {
	Regex         []string `yaml:"regex,omitempty"`
	MatchStrings  []string `yaml:"strings,omitempty"`
	SilentStartUp bool     `yaml:"silentStartUp"`
	StatusCode    int      `yaml:"statusCode"`
}

/**********************************
 * Define traefik related methods *
 **********************************/

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		SilentStartUp: true,
		StatusCode:    403, // https://cs.opensource.google/go/go/+/refs/tags/go1.21.4:src/net/http/status.go
	}
}

// New creates a new plugin.
// Returns the configured BlockUrls plugin object.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// if len(config.MatchStrings) == 0 {
	// 	return nil, fmt.Errorf("the match strings list is empty")
	// }

	// if len(config.Regex) == 0 {
	// 	return nil, fmt.Errorf("the regex list is empty")
	// }

	if !config.SilentStartUp {
		log.Println("Regex list: ", config.Regex)
		log.Println("Match String list: ", config.MatchStrings)
		log.Println("StatusCode: ", config.StatusCode)
	}

	// regular expressions
	matchStrings := make([]string, len(config.MatchStrings))
	copy(matchStrings, config.MatchStrings)

	// regular expressions
	regexps := make([]*regexp.Regexp, len(config.Regex))

	for index, regex := range config.Regex {
		compiledRegex, compileError := regexp.Compile(regex)
		if compileError != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", regex, compileError)
		}

		regexps[index] = compiledRegex
	}

	return &traefik_block_regex_urls{
		next:          next,
		name:          name,
		regexps:       regexps,
		matchStrings:  matchStrings,
		silentStartUp: config.SilentStartUp,
		statusCode:    config.StatusCode,
	}, nil
}

// This method is the middleware called during runtime and handling middleware actions.
func (blockUrls *traefik_block_regex_urls) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	fullUrl := request.Host + request.URL.RequestURI()

	for _, str := range blockUrls.matchStrings {
		if strings.Contains(fullUrl, str) {
			log.Printf("URL is blocked (substring match): (%s): module=%s", fullUrl, blockUrls.name)
			responseWriter.WriteHeader(blockUrls.statusCode)
			return
		}
	}

	for _, regex := range blockUrls.regexps {
		if regex.MatchString(fullUrl) {
			log.Printf("URL is blocked (regex match): (%s) module=%s", fullUrl, blockUrls.name)
			responseWriter.WriteHeader(blockUrls.statusCode)
			return
		}
	}

	blockUrls.next.ServeHTTP(responseWriter, request)
}
