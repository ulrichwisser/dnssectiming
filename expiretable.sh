#!/bin/bash

# Run the DNSSEC timing expire command
./dnssectiming expire > data/expire.data

# Run the gnuplot command
gnuplot -c gnuplot/expire.plt

# Get the last date from the expire.data file (the first column of the last line)
last_date=$(awk 'END {print $1}' data/expire.data)

# Filter the data to use only the rows where the first column matches the last_date
filtered_data=$()

# Count unique values from the third column and sort the results
echo "Results for date: $last_date"
#awk -v last_date="$last_date" '$1 == last_date' data/expire.data | awk '{print $3}' | sort | uniq -c | sort -k2,2n
awk -v last_date="$last_date" '$1 == last_date {print $3}' data/expire.data | \
    sort | \
    uniq -c | \
    sort -k2,2n | \
    while read count seconds; do
        # Convert seconds into days, hours, minutes, and seconds
        days=$((seconds / 86400))
        hours=$(((seconds % 86400) / 3600))
        minutes=$(((seconds % 3600) / 60))
        remaining_seconds=$((seconds % 60))

        # Print the result with the converted time format
        printf "%5s %11d %5dd %2dh %2dm %2ds\n" "$count" "$seconds" "$days" "$hours" "$minutes" "$remaining_seconds"
    done
