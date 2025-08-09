#!/bin/bash

# Script to download the IANA RDAP bootstrap file for Docker deployment
# This file contains the mapping of TLDs to RDAP servers

echo "Downloading IANA RDAP bootstrap file..."

# Download the bootstrap file
curl -s https://data.iana.org/rdap/dns.json > bootstrap.json

# Check if download was successful
if [ $? -eq 0 ]; then
    echo "✅ Successfully downloaded bootstrap.json"
    echo "📊 File size: $(wc -c < bootstrap.json) bytes"
    echo "📅 Last modified: $(date -r bootstrap.json)"
    echo ""
    echo "To use this in your Docker deployment:"
    echo "1. Copy this file to your Docker image:"
    echo "   COPY bootstrap.json /opt/bootstrap.json"
    echo ""
    echo "2. Use it in your Go code:"
    echo "   client := rdap.NewClient().SetBootstrapFile(\"/opt/bootstrap.json\")"
    echo ""
    echo "3. Update this file periodically (recommended: weekly)"
else
    echo "❌ Failed to download bootstrap file"
    exit 1
fi
