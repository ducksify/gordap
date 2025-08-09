package rdap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)


func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient should not return nil")
	}
	if client.httpClient == nil {
		t.Error("HTTP client should not be nil")
	}
	if client.bootstrapURL != defaultRDAPBootstrapURL {
		t.Errorf("Expected bootstrap URL %s, got %s", defaultRDAPBootstrapURL, client.bootstrapURL)
	}
}

func TestSetTimeout(t *testing.T) {
	client := NewClient()
	timeout := 10 * time.Second
	client.SetTimeout(timeout)
	if httpClient, ok := client.httpClient.(*http.Client); ok {
		if httpClient.Timeout != timeout {
			t.Errorf("Expected timeout %v, got %v", timeout, httpClient.Timeout)
		}
	} else {
		t.Error("Expected httpClient to be *http.Client")
	}
}

func TestSetBootstrapURL(t *testing.T) {
	client := NewClient()
	customURL := "https://example.com/rdap.json"
	client.SetBootstrapURL(customURL)
	if client.bootstrapURL != customURL {
		t.Errorf("Expected bootstrap URL %s, got %s", customURL, client.bootstrapURL)
	}
}

func TestSetBootstrapFile(t *testing.T) {
	client := NewClient()
	filepath := "/path/to/bootstrap.json"
	client.SetBootstrapFile(filepath)
	expected := "file://" + filepath
	if client.bootstrapURL != expected {
		t.Errorf("Expected bootstrap URL %s, got %s", expected, client.bootstrapURL)
	}
}

func TestSetDisableCache(t *testing.T) {
	client := NewClient()

	// Test enabling cache (default)
	if client.disableCache {
		t.Error("Cache should be enabled by default")
	}

	// Test disabling cache
	client.SetDisableCache(true)
	if !client.disableCache {
		t.Error("Cache should be disabled after SetDisableCache(true)")
	}

	// Test re-enabling cache
	client.SetDisableCache(false)
	if client.disableCache {
		t.Error("Cache should be enabled after SetDisableCache(false)")
	}
}

func TestSetCacheBootstrapOnly(t *testing.T) {
	client := NewClient()

	// Test default state
	if client.cacheBootstrapOnly {
		t.Error("Bootstrap-only cache should be disabled by default")
	}

	// Test enabling bootstrap-only cache
	client.SetCacheBootstrapOnly(true)
	if !client.cacheBootstrapOnly {
		t.Error("Bootstrap-only cache should be enabled after SetCacheBootstrapOnly(true)")
	}

	// Test disabling bootstrap-only cache
	client.SetCacheBootstrapOnly(false)
	if client.cacheBootstrapOnly {
		t.Error("Bootstrap-only cache should be disabled after SetCacheBootstrapOnly(false)")
	}
}

func TestGetTLD(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"example.com", "com"},
		{"test.org", "org"},
		{"sub.example.net", "net"},
		{"example.ch", "ch"},
		{"invalid", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := getTLD(test.domain)
		if result != test.expected {
			t.Errorf("getTLD(%s) = %s, expected %s", test.domain, result, test.expected)
		}
	}
}

func TestGetBootstrapDataFromURL(t *testing.T) {
	// Create a mock bootstrap server
	bootstrapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock bootstrap response
		bootstrap := RDAPBootstrap{
			Description: "RDAP bootstrap file for Domain Name System registrations",
			Publication: "2025-01-01T00:00:00Z",
			Services: [][][]string{
				{
					{"com"},
					{"https://rdap.verisign.com/com/v1/"},
				},
				{
					{"org"},
					{"https://rdap.pir.org/org/v1/"},
				},
			},
			Version: "1.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bootstrap)
	}))
	defer bootstrapServer.Close()

	client := NewClient()
	client.SetBootstrapURL(bootstrapServer.URL)

	bootstrap, err := client.getBootstrapData()
	if err != nil {
		t.Fatalf("Failed to get bootstrap data: %v", err)
	}

	if bootstrap == nil {
		t.Fatal("Bootstrap data should not be nil")
	}

	if bootstrap.Description != "RDAP bootstrap file for Domain Name System registrations" {
		t.Errorf("Expected description 'RDAP bootstrap file for Domain Name System registrations', got: %s", bootstrap.Description)
	}

	if bootstrap.Version != "1.0" {
		t.Errorf("Expected version '1.0', got: %s", bootstrap.Version)
	}

	if len(bootstrap.Services) != 2 {
		t.Errorf("Expected 2 services, got: %d", len(bootstrap.Services))
	}
}

func TestGetBootstrapDataFromFile(t *testing.T) {
	// Create a temporary bootstrap file
	tempFile, err := os.CreateTemp("", "bootstrap-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write bootstrap data to temp file
	bootstrapData := `{
		"description": "Test bootstrap file",
		"publication": "2025-01-01T00:00:00Z",
		"services": [
			[
				["com"],
				["https://rdap.verisign.com/com/v1/"]
			],
			[
				["org"],
				["https://rdap.pir.org/org/v1/"]
			]
		],
		"version": "1.0"
	}`

	if _, err := tempFile.WriteString(bootstrapData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	client := NewClient()
	client.SetBootstrapFile(tempFile.Name())

	bootstrap, err := client.getBootstrapData()
	if err != nil {
		t.Fatalf("Failed to get bootstrap data from file: %v", err)
	}

	if bootstrap == nil {
		t.Fatal("Bootstrap data should not be nil")
	}

	if bootstrap.Description != "Test bootstrap file" {
		t.Errorf("Expected description 'Test bootstrap file', got: %s", bootstrap.Description)
	}

	if bootstrap.Version != "1.0" {
		t.Errorf("Expected version '1.0', got: %s", bootstrap.Version)
	}

	if len(bootstrap.Services) != 2 {
		t.Errorf("Expected 2 services, got: %d", len(bootstrap.Services))
	}
}

func TestGetBootstrapDataFileNotFound(t *testing.T) {
	client := NewClient()
	client.SetBootstrapFile("/nonexistent/file.json")

	_, err := client.getBootstrapData()
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "failed to read bootstrap file") {
		t.Errorf("Expected error about file read failure, got: %v", err)
	}
}

func TestGetBootstrapDataInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempFile, err := os.CreateTemp("", "invalid-bootstrap-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write invalid JSON
	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	client := NewClient()
	client.SetBootstrapFile(tempFile.Name())

	_, err = client.getBootstrapData()
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse bootstrap JSON") {
		t.Errorf("Expected error about JSON parsing failure, got: %v", err)
	}
}

func TestGetBootstrapDataHTTPError(t *testing.T) {
	// Create a mock server that returns an error
	bootstrapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer bootstrapServer.Close()

	client := NewClient()
	client.SetBootstrapURL(bootstrapServer.URL)

	_, err := client.getBootstrapData()
	if err == nil {
		t.Fatal("Expected error for HTTP error response")
	}

	if !strings.Contains(err.Error(), "bootstrap request failed with status: 500") {
		t.Errorf("Expected error about HTTP status failure, got: %v", err)
	}
}

func TestGetBootstrapDataNetworkError(t *testing.T) {
	client := NewClient()
	client.SetBootstrapURL("http://nonexistent-server.local/bootstrap.json")

	_, err := client.getBootstrapData()
	if err == nil {
		t.Fatal("Expected error for network failure")
	}

	if !strings.Contains(err.Error(), "failed to fetch bootstrap data") {
		t.Errorf("Expected error about network failure, got: %v", err)
	}
}

func TestRDAPWithMockServer(t *testing.T) {
	// Create a mock RDAP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock RDAP response
		response := map[string]interface{}{
			"objectClassName": "domain",
			"ldhName":         "example.com",
			"status":          []string{"active"},
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create a mock bootstrap server
	bootstrapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock bootstrap response
		bootstrap := RDAPBootstrap{
			Description: "RDAP bootstrap file for Domain Name System registrations",
			Publication: "2025-01-01T00:00:00Z",
			Services: [][][]string{
				{
					{"com"},
					{mockServer.URL + "/"},
				},
			},
			Version: "1.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bootstrap)
	}))
	defer bootstrapServer.Close()

	client := NewClient()
	client.SetBootstrapURL(bootstrapServer.URL)

	result, err := client.RDAP("example.com")
	if err != nil {
		t.Fatalf("RDAP query failed: %v", err)
	}

	if !strings.Contains(string(result), "example.com") {
		t.Errorf("Expected response to contain 'example.com', got: %s", result)
	}
}

func TestRDAPEmptyDomain(t *testing.T) {
	client := NewClient()
	_, err := client.RDAP("")
	if err == nil {
		t.Error("Expected error for empty domain")
	}
	if !strings.Contains(err.Error(), "domain cannot be empty") {
		t.Errorf("Expected error about empty domain, got: %v", err)
	}
}

func TestRDAPInvalidDomain(t *testing.T) {
	client := NewClient()
	_, err := client.RDAP("invalid")
	if err == nil {
		t.Error("Expected error for invalid domain")
	}
}

func TestFindServerForTLD(t *testing.T) {
	client := NewClient()

	bootstrap := &RDAPBootstrap{
		Services: [][][]string{
			{
				{"com", "org"},
				{"https://rdap.verisign.com/com/v1/", "https://rdap.pir.org/org/v1/"},
			},
			{
				{"net"},
				{"https://rdap.verisign.com/net/v1/"},
			},
		},
	}

	// Test finding server for .com
	server, err := client.findServerForTLD("com", bootstrap)
	if err != nil {
		t.Fatalf("Failed to find server for .com: %v", err)
	}
	expected := "https://rdap.verisign.com/com/v1/"
	if server != expected {
		t.Errorf("Expected server %s, got %s", expected, server)
	}

	// Test finding server for .net
	server, err = client.findServerForTLD("net", bootstrap)
	if err != nil {
		t.Fatalf("Failed to find server for .net: %v", err)
	}
	expected = "https://rdap.verisign.com/net/v1/"
	if server != expected {
		t.Errorf("Expected server %s, got %s", expected, server)
	}

	// Test non-existent TLD
	_, err = client.findServerForTLD("nonexistent", bootstrap)
	if err == nil {
		t.Error("Expected error for non-existent TLD")
	}
}

func TestGetRDAPServerForCH(t *testing.T) {
	client := NewClient()

	server, err := client.getRDAPServer("example.ch")
	if err != nil {
		t.Fatalf("Failed to get RDAP server for .ch domain: %v", err)
	}

	expected := "https://rdap.nic.ch/domain/example.ch"
	if server != expected {
		t.Errorf("Expected server %s, got %s", expected, server)
	}
}

func TestDefaultClient(t *testing.T) {
	if DefaultClient == nil {
		t.Error("DefaultClient should not be nil")
	}
}

func TestQueryRDAPSuccess(t *testing.T) {
	// Create a mock RDAP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path
		expectedPath := "/domain/example.com"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Mock RDAP response
		response := map[string]interface{}{
			"objectClassName": "domain",
			"ldhName":         "example.com",
			"status":          []string{"active"},
			"events": []map[string]interface{}{
				{
					"eventAction": "registration",
					"eventDate":   "1995-08-14T04:00:00Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewClient()
	result, err := client.queryRDAP("example.com", mockServer.URL+"/")
	if err != nil {
		t.Fatalf("queryRDAP failed: %v", err)
	}

	// Verify the response contains expected data
	if !strings.Contains(string(result), "example.com") {
		t.Errorf("Expected response to contain 'example.com', got: %s", result)
	}

	if !strings.Contains(string(result), "active") {
		t.Errorf("Expected response to contain 'active', got: %s", result)
	}

	// Verify it's valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}
}

func TestQueryRDAPCHDomain(t *testing.T) {
	// Create a mock server for .ch domains
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For .ch domains, the full URL is passed directly
		expectedPath := "/domain/example.ch"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Mock RDAP response for .ch domain
		response := map[string]interface{}{
			"objectClassName": "domain",
			"ldhName":         "example.ch",
			"status":          []string{"active"},
			"handle":          "example.ch",
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewClient()
	// Test the regular domain path construction (not the special .ch case)
	result, err := client.queryRDAP("example.ch", mockServer.URL+"/")
	if err != nil {
		t.Fatalf("queryRDAP failed for .ch domain: %v", err)
	}

	// Verify the response contains expected data
	if !strings.Contains(string(result), "example.ch") {
		t.Errorf("Expected response to contain 'example.ch', got: %s", result)
	}

	if !strings.Contains(string(result), "active") {
		t.Errorf("Expected response to contain 'active', got: %s", result)
	}
}

func TestQueryRDAPHTTPError(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Domain not found"}`))
	}))
	defer mockServer.Close()

	client := NewClient()
	_, err := client.queryRDAP("nonexistent.com", mockServer.URL+"/")
	if err == nil {
		t.Fatal("Expected error for HTTP 404 response")
	}

	if !strings.Contains(err.Error(), "RDAP query failed with status 404") {
		t.Errorf("Expected error about HTTP 404, got: %v", err)
	}
}

func TestQueryRDAPNetworkError(t *testing.T) {
	client := NewClient()
	_, err := client.queryRDAP("example.com", "http://nonexistent-server.local/domain/example.com")
	if err == nil {
		t.Fatal("Expected error for network failure")
	}

	if !strings.Contains(err.Error(), "RDAP query failed") {
		t.Errorf("Expected error about network failure, got: %v", err)
	}
}

func TestQueryRDAPInvalidResponse(t *testing.T) {
	// Create a mock server that returns invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		w.Write([]byte("invalid json response"))
	}))
	defer mockServer.Close()

	client := NewClient()
	result, err := client.queryRDAP("example.com", mockServer.URL+"/")
	if err != nil {
		t.Fatalf("queryRDAP should not fail for invalid JSON, got: %v", err)
	}

	// Should return the raw response even if it's invalid JSON
	if string(result) != "invalid json response" {
		t.Errorf("Expected raw response, got: %s", result)
	}
}

func TestQueryRDAPEmptyResponse(t *testing.T) {
	// Create a mock server that returns empty response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		// Return empty response
	}))
	defer mockServer.Close()

	client := NewClient()
	result, err := client.queryRDAP("example.com", mockServer.URL+"/")
	if err != nil {
		t.Fatalf("queryRDAP failed for empty response: %v", err)
	}

	if string(result) != "" {
		t.Errorf("Expected empty response, got: %s", result)
	}
}

func TestQueryRDAPLargeResponse(t *testing.T) {
	// Create a mock server that returns a large response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a large response (1MB)
		largeData := make([]byte, 1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}
		w.Header().Set("Content-Type", "application/rdap+json")
		w.Write(largeData)
	}))
	defer mockServer.Close()

	client := NewClient()
	result, err := client.queryRDAP("example.com", mockServer.URL+"/")
	if err != nil {
		t.Fatalf("queryRDAP failed for large response: %v", err)
	}

	if len(result) != 1024*1024 {
		t.Errorf("Expected response size 1MB, got %d bytes", len(result))
	}
}

func TestQueryRDAPTimeout(t *testing.T) {
	// Create a mock server that delays response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/rdap+json")
		w.Write([]byte(`{"status": "timeout"}`))
	}))
	defer mockServer.Close()

	client := NewClient().SetTimeout(1 * time.Second)
	_, err := client.queryRDAP("example.com", mockServer.URL+"/")
	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "RDAP query failed") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestRDAPFunction(t *testing.T) {
	// This test verifies the function exists and can be called
	// It may succeed or fail depending on network connectivity and server availability
	_, err := RDAP("example.com")
	// We don't assert on the result since it depends on external factors
	if err != nil {
		// Log the error for debugging but don't fail the test
		t.Logf("RDAP function call resulted in error (expected in some environments): %v", err)
	} else {
		t.Log("RDAP function call succeeded")
	}
}
