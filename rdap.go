/*
 * Copyright 2024 François "@Ducksify"
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Go module for domain RDAP information query
 */

package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// defaultRDAPBootstrapURL is the IANA RDAP bootstrap URL
	defaultRDAPBootstrapURL = "https://data.iana.org/rdap/dns.json"
	// defaultTimeout is query default timeout
	defaultTimeout = 30 * time.Second
	// bootstrapCacheDuration is how long to cache the bootstrap data
	bootstrapCacheDuration = 24 * time.Hour
)

// DefaultClient is default RDAP client
var DefaultClient = NewClient()

// HTTPClient defines the interface for HTTP client operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RDAPBootstrap represents the IANA RDAP bootstrap file structure
type RDAPBootstrap struct {
	Description string       `json:"description"`
	Publication string       `json:"publication"`
	Services    [][][]string `json:"services"`
	Version     string       `json:"version"`
}

// Client is RDAP client
type Client struct {
	httpClient         HTTPClient
	bootstrapURL       string
	serverMap          map[string]string
	disableCache       bool
	cacheBootstrapOnly bool
}

// RDAP do the RDAP query and returns RDAP information
func RDAP(domain string) (result []byte, err error) {
	return DefaultClient.RDAP(domain)
}

// NewClient returns new RDAP client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		bootstrapURL:       defaultRDAPBootstrapURL,
		serverMap:          make(map[string]string),
		disableCache:       false,
		cacheBootstrapOnly: false,
	}
}

// SetHTTPClient sets the HTTP client
func (c *Client) SetHTTPClient(client HTTPClient) *Client {
	c.httpClient = client
	return c
}

// SetTimeout sets query timeout
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	if httpClient, ok := c.httpClient.(*http.Client); ok {
		httpClient.Timeout = timeout
	}
	return c
}

// SetBootstrapURL sets the bootstrap URL
func (c *Client) SetBootstrapURL(url string) *Client {
	c.bootstrapURL = url
	return c
}

// SetDisableCache disables caching for Lambda environments
func (c *Client) SetDisableCache(disabled bool) *Client {
	c.disableCache = disabled
	return c
}

// SetCacheBootstrapOnly enables caching only for bootstrap data, not domain queries
func (c *Client) SetCacheBootstrapOnly(enabled bool) *Client {
	c.cacheBootstrapOnly = enabled
	return c
}

// SetBootstrapFile sets the path to a local bootstrap file
func (c *Client) SetBootstrapFile(filepath string) *Client {
	c.bootstrapURL = "file://" + filepath
	return c
}

// RDAPRaw performs RDAP query for the given domain and returns raw JSON
func (c *Client) RDAP(domain string) (result []byte, err error) {
	// Normalize domain
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Get the appropriate RDAP server for this domain
	server, err := c.getRDAPServer(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get RDAP server for %s: %w", domain, err)
	}

	// Perform the RDAP query
	return c.queryRDAP(domain, server)
}

// getRDAPServer determines the appropriate RDAP server for a domain
func (c *Client) getRDAPServer(domain string) (string, error) {
	// Extract TLD from domain
	tld := getTLD(domain)
	if tld == "" {
		return "", fmt.Errorf("invalid domain: %s", domain)
	}

	// Special case for .ch domains
	if tld == "ch" {
		return "https://rdap.nic.ch/domain/" + domain, nil
	}

	// Get bootstrap data
	bootstrap, err := c.getBootstrapData()
	if err != nil {
		return "", fmt.Errorf("failed to get bootstrap data: %w", err)
	}

	// Find the appropriate server for this TLD
	server, err := c.findServerForTLD(tld, bootstrap)
	if err != nil {
		return "", fmt.Errorf("no RDAP server found for TLD %s: %w", tld, err)
	}

	return server, nil
}

// getBootstrapData fetches the IANA RDAP bootstrap data
func (c *Client) getBootstrapData() (*RDAPBootstrap, error) {
	var body []byte
	var err error

	// Check if we're reading from a local file
	if strings.HasPrefix(c.bootstrapURL, "file://") {
		filepath := strings.TrimPrefix(c.bootstrapURL, "file://")
		body, err = os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to read bootstrap file %s: %w", filepath, err)
		}
	} else {
		// Fetch from URL
		req, err := http.NewRequest("GET", c.bootstrapURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bootstrap data: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bootstrap request failed with status: %d", resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read bootstrap response: %w", err)
		}
	}

	var bootstrap RDAPBootstrap
	if err := json.Unmarshal(body, &bootstrap); err != nil {
		return nil, fmt.Errorf("failed to parse bootstrap JSON: %w", err)
	}

	return &bootstrap, nil
}

// findServerForTLD finds the appropriate RDAP server for a given TLD
func (c *Client) findServerForTLD(tld string, bootstrap *RDAPBootstrap) (string, error) {
	for _, service := range bootstrap.Services {
		if len(service) != 2 {
			continue
		}

		tlds := service[0]
		servers := service[1]

		for _, serviceTLD := range tlds {
			if serviceTLD == tld && len(servers) > 0 {
				// Use the first server in the list
				server := servers[0]
				// Ensure the URL ends with a slash
				if !strings.HasSuffix(server, "/") {
					server += "/"
				}
				return server, nil
			}
		}
	}

	return "", fmt.Errorf("no server found for TLD: %s", tld)
}

// queryRDAPBytes performs the actual RDAP query and returns raw bytes
func (c *Client) queryRDAP(domain, server string) ([]byte, error) {
	// For .ch domains, the server URL already includes the full path
	if strings.Contains(server, "rdap.nic.ch") {
		req, err := http.NewRequest("GET", server, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Accept", "application/rdap+json;charset=UTF-8")
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("RDAP query failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read RDAP response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("RDAP query failed with status %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}

	// For other domains, construct the query URL
	queryURL := server + "domain/" + domain

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/rdap+json;charset=UTF-8")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RDAP query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDAP response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RDAP query failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// getTLD extracts the TLD from a domain
func getTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}
