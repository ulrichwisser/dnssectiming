#
# usage: gnuplot -c gnuplot/lifet.plt TLS RRTYPE
#
tld=ARG1
rr=ARG2

filename = sprintf("data/lifet.%s.%s.data", tld, rr)
outputname = sprintf("data/lifet.%s.%s.png", tld, rr)
title_str = sprintf("RRSIG Lifetime for %s records of %s", rr, tld)
title_lifetime = sprintf("Lifetime of RRSIG for %s records", rr)

set style data lines
set title title_str
set xlabel "Date"
set ylabel "RRSIG Lifetime"
set xdata time
set timefmt '%Y-%m-%d'
set format y ""
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [0:]

set ytics()
set ytics add ( "1h" 3600, "1d" 84600, "7d" 604800, "14d" 1209600, "30d" 2592000, "100d" 8460000, "7000d" 604800000 )

set terminal png size 1024,768
set output outputname

plot filename using 1:2 title title_lifetime, filename using 1:3 title "SOA Expire"
