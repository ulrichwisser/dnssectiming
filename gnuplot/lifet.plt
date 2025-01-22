#
# usage: gnuplot -c gnuplot/lifet.plt TLS RRTYPE
#
tld=ARG1
rr=ARG2

filename = sprintf("data/lifet.%s.%s.data", tld, rr)
outputname = sprintf("data/lifet.%s.%s.png", tld, rr)
title_str = sprintf("Lifetime data for .%s %s records", tld, rr)

set style data lines
set title title_str
set xlabel "Date"
set ylabel "Seconds"
set xdata time
set timefmt '%Y-%m-%d'
set format y "%.0f"
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [0:]

set terminal png size 1024,768
set output outputname

plot filename using 1:2 title "Lifetime", filename using 1:3 title "SOA Expire"
