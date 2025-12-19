#
# usage: gnuplot -c gnuplot/remaining.plt RRTYPE
#
rr=ARG1

filename = sprintf("data/remaining.%s.data", rr)
outputname = sprintf("data/remaining.%s.png", rr)
title_str = sprintf("Number of TLDs with a lifetime for %s records below", rr)

set style data points

set terminal pngcairo size 1024,768
set output outputname

set logscale y
#set yrange [0:*]
eps = 0

plot filename  using ( $5>0 ? $5 : eps ) w p pt 7 ps 1.0 t "between 7 and 14 days", \
     '' using ( $4>0 ? $4 : eps ) w p pt 7 ps 1.0 t "between 3 and 7 days", \
     '' using ( $3>0 ? $3 : eps ) w p pt 7 ps 1.0 t "between 1 and 3 days", \
     '' using ( $2>0 ? $2 : eps ) w p pt 7 ps 1.0 t "under 24h"     

#pause 30