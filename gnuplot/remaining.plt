#
# usage: gnuplot -c gnuplot/remaining.plt RRTYPE
#
rr=ARG1

filename = sprintf("data/remaining.%s.data", rr)
outputname = sprintf("data/remaining.%s.png", rr)
title_str = sprintf("Remaining lifetime for %s records", rr)

set style data points
set title title_str
set xlabel "Date"
set ylabel "Number of TLDs"
set xdata time
set timefmt '%Y-%m-%d'
set format y "%g"
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [1:2000]
set logscale y
eps = 0 #set to 0.1 to show 0 values

set terminal pngcairo size 1024,768
set output outputname


plot filename  using 1:( $5>0 ? $5 : eps ) w p pt 7 ps 1.0 t "between 7 and 14 days", \
     '' using 1:( $4>0 ? $4 : eps ) w p pt 7 ps 1.0 t "between 3 and 7 days", \
     '' using 1:( $3>0 ? $3 : eps ) w p pt 7 ps 1.0 t "between 1 and 3 days", \
     '' using 1:( $2>0 ? $2 : eps ) w p pt 7 ps 1.0 t "under 24h"     

#pause -1