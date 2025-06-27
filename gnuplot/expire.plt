#
# usage: gnuplot -c gnuplot/expire.plt
#
filename = "data/expire.data"
outputname = "data/expire.png"
title_str = "SOA Expire data for allTLDs"

set style data dots
set title title_str
set xlabel "Date"
set ylabel "Seconds"
set xdata time
set timefmt '%Y-%m-%d'
set format y ""
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [0:]
set logscale y

set ytics add ( "1h" 3600, "1d" 84600, "7d" 604800, "14d" 1209600, "30d" 2592000, "100d" 8460000, "7000d" 604800000 )

set terminal png size 1024,768
set output outputname

plot filename using 1:3 notitle
