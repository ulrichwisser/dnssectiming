#!/bin/bash

# Run the DNSSEC timing expire command
./dnssectiming failed -r NS -v > data/failed.ns.data
./dnssectiming failed -r DNSKEY -v > data/failed.dnskey.data

# Run the gnuplot command
gnuplot -c gnuplot/failed.plt ns
gnuplot -c gnuplot/failed.plt dnskey

