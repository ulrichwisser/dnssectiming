#
# usage: gnuplot -c gnuplot/lifet.plt TLD RRTYPE
#
rr=ARG1

filename = sprintf("data/failed.%s.data", rr)
outputname = sprintf("data/failed.%s.png", rr)

title_str = sprintf("Number of gTLDs / ccTLDs with RRSIG lifetime of %s records shorter than SOA expire", rr)

set style data lines
set title title_str
set xlabel "Date"
set ylabel "RRSIG Lifetime"
set xdata time
set timefmt '%Y-%m-%d'
set format y "%.0f"
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [0:]

set terminal png size 1024,768
set output outputname

plot filename using 1:2 title "ccTLD ok", filename using 1:3 title "ccTLD too short", filename using 1:4 title "gTLD ok", filename using 1:5 title "gTLD too short"

