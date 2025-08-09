package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ducksify/gordap"
)

func main() {
	// Default to example.com if no domain provided
	domain := "example.com"
	if len(os.Args) >= 2 {
		domain = os.Args[1]
	}

	fmt.Printf("🔍 RDAP Client Example\n")
	fmt.Printf("Querying: %s\n", domain)
	fmt.Println(strings.Repeat("─", 50))

	// Create a client with 10 second timeout
	client := rdap.NewClient().SetTimeout(10 * time.Second)

	// Query RDAP information
	result, err := client.RDAP(domain)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	// Parse and pretty print the JSON response
	var rdapData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &rdapData); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		fmt.Printf("Raw response:\n%s\n", result)
		return
	}

	// Pretty print the JSON
	prettyJSON, _ := json.MarshalIndent(rdapData, "", "  ")
	fmt.Printf("✅ Success! RDAP data for %s:\n\n", domain)
	fmt.Println(string(prettyJSON))

	// Extract and display key information
	fmt.Println("\n📋 Key Information:")
	if ldhName, ok := rdapData["ldhName"].(string); ok {
		fmt.Printf("  Domain: %s\n", ldhName)
	}
	if status, ok := rdapData["status"].([]interface{}); ok && len(status) > 0 {
		fmt.Printf("  Status: %v\n", status)
	}
	if events, ok := rdapData["events"].([]interface{}); ok && len(events) > 0 {
		fmt.Printf("  Events: %d events found\n", len(events))
	}
	if entities, ok := rdapData["entities"].([]interface{}); ok && len(entities) > 0 {
		fmt.Printf("  Entities: %d entities found\n", len(entities))
	}
}
