// Package "block regex urls" is a Traefik plugin to block access to certain urls using a list of regex values and return a defined status code.
package traefik_block_regex_urls

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

/**********************************
 *          Define types          *
 **********************************/

type traefik_block_regex_urls struct {
	next               http.Handler
	name               string
	allowLocalRequests bool // this field is unused
	privateIPRanges    []*net.IPNet
	regexps            []*regexp.Regexp
	matchStrings       []string
	silentStartUp      bool
	statusCode         int
}

type Config struct {
	AllowLocalRequests bool     `yaml:"allowLocalRequests"` // this field is unused
	Regex              []string `yaml:"regex,omitempty"`
	MatchStrings       []string `yaml:"strings,omitempty"`
	SilentStartUp      bool     `yaml:"silentStartUp"`
	StatusCode         int      `yaml:"statusCode"`
}

/**********************************
 * Define traefik related methods *
 **********************************/

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		AllowLocalRequests: true, // this field is unused
		SilentStartUp:      true,
		StatusCode:         403, // https://cs.opensource.google/go/go/+/refs/tags/go1.21.4:src/net/http/status.go
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
	regexps := make([]*regexp.Regexp, len(config.Regex))

	for index, regex := range config.Regex {
		compiledRegex, compileError := regexp.Compile(regex)
		if compileError != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", regex, compileError)
		}

		regexps[index] = compiledRegex
	}

	return &traefik_block_regex_urls{
		next:               next,
		name:               name,
		allowLocalRequests: config.AllowLocalRequests, // this field is unused
		privateIPRanges:    InitializePrivateIPBlocks(),
		regexps:            regexps,
		matchStrings:       config.MatchStrings,
		silentStartUp:      config.SilentStartUp,
		statusCode:         config.StatusCode,
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

/**********************************
 *         Private methods        *
 **********************************/

// This method collects the remote IP address.
// It tries to parse the IP from the HTTP request.
// Returns the parsed IP and no error on success, otherwise the so far generated list and an error.
func (blockUrls *traefik_block_regex_urls) CollectRemoteIP(request *http.Request) ([]*net.IP, error) {
	var ipList []*net.IP

	// Helper method to split a string at char ','
	splitFn := func(c rune) bool {
		return c == ','
	}

	// Try to parse from header "X-Forwarded-For"
	xForwardedForValue := request.Header.Get("X-Forwarded-For")
	xForwardedForIPs := strings.FieldsFunc(xForwardedForValue, splitFn)
	for _, value := range xForwardedForIPs {
		ipAddress, err := ParseIP(value)
		if err != nil {
			return ipList, fmt.Errorf("parsing failed: %s", err)
		}

		ipList = append(ipList, &ipAddress)
	}

	// Try to parse from header "X-Real-IP"
	xRealIpValue := request.Header.Get("X-Real-IP")
	xRealIpIPs := strings.FieldsFunc(xRealIpValue, splitFn)
	for _, value := range xRealIpIPs {
		ipAddress, err := ParseIP(value)
		if err != nil {
			return ipList, fmt.Errorf("parsing failed: %s", err)
		}

		ipList = append(ipList, &ipAddress)
	}

	return ipList, nil
}

// This method initializes a list of private IP addresses.
// It uses a predefined range of CIDR addresses.
// Returns a list of private IP blocks.
// https://stackoverflow.com/questions/41240761/check-if-ip-address-is-in-private-network-space
func InitializePrivateIPBlocks() []*net.IPNet {
	var privateIPBlocks []*net.IPNet

	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}

	return privateIPBlocks
}

// This method checks whether a provided IP is a private IP.
// If this is the case it returns true, otherwise false.
// https://stackoverflow.com/questions/41240761/check-if-ip-address-is-in-private-network-space
func IsPrivateIP(ip net.IP, privateIPBlocks []*net.IPNet) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return IsIpInList(ip, privateIPBlocks)
}

// Checks whether a string is in a list of strings.
// Returns true if this is the case, otherwise returns false.
func IsIpInList(ip net.IP, list []*net.IPNet) bool {
	for _, block := range list {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

// Tries to parse the IP from a provided address.
// Returns the ip and no error on success, otherwise returns nil and the occured error.
func ParseIP(address string) (net.IP, error) {
	ipAddress := net.ParseIP(address)

	if ipAddress == nil {
		return nil, fmt.Errorf("unable to parse IP from address [%s]", address)
	}

	return ipAddress, nil
}
