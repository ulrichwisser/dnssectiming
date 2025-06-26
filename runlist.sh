#!/bin/bash

if [ -z "$1" ]; then
    echo "First argument must be filename of file containing a tld list"
    exit 1
fi

input_file="$1"

if [ ! -f "$input_file" ]; then
    echo "File $input_file not found!"
    exit 1
fi

mkdir -p data

while IFS= read -r tld; do
    # ignore comments and empty lines
    if [ -z "$tld" ] || [[ "$tld" == \#* ]]; then
        continue
    fi

    echo "Executing for $tld"

    # DNSSEC Timing für NS und DNSKEY
    ./dnssectiming lifetime -r NS -t "$tld" -v > "data/lifet.$tld.ns.data"
    ./dnssectiming lifetime -r DNSKEY -t "$tld" -v > "data/lifet.$tld.dnskey.data"

    # if no data is available remove .data file, otherwise run gnuplot
    if [ ! -s "data/lifet.$tld.ns.data" ]; then
        echo "No NS data for $tld"
        rm "data/lifet.$tld.ns.data"
    else
        gnuplot -c gnuplot/lifet.plt "$tld" ns
    fi

    if [ ! -s "data/lifet.$tld.dnskey.data" ]; then
        echo "No DNSKEY data for $tld"
        rm "data/lifet.$tld.dnskey.data"
    else
        gnuplot -c gnuplot/lifet.plt "$tld" dnskey
    fi

done < "$input_file"
