#!/bin/bash

# Run the DNSSEC timing expire command
#dnssectiming expire > data/expire.data

# Run the gnuplot command
#gnuplot -c gnuplot/expire.plt

# Get the last date from the expire.data file (the first column of the last line)
last_date=$(awk 'END {print $1}' data/expire.data)

# Filter the data to use only the rows where the first column matches the last_date
filtered_data=$()

# Count unique values from the third column and sort the results
echo "Results for date: $last_date"
awk -v last_date="$last_date" '$1 == last_date' data/expire.data | awk '{print $3}' | sort | uniq -c | sort -k2,2n
