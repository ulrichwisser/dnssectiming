#!/bin/bash

# URL to download the list of current TLDs from IANA
iana_url=""

# Download the IANA TLD list
curl -s -O https://data.iana.org/TLD/tlds-alpha-by-domain.txt

# Extract the list of ccTLDs (country-code TLDs) from the IANA list
awk 'length($1) == 2' tlds-alpha-by-domain.txt > cctlds.txt

# get gTLDs
grep -vFf cctlds.txt tlds-alpha-by-domain.txt > gtlds.txt