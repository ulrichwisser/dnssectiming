#
# usage: gnuplot -c gnuplot/rfc6781.plt RRTYPE
#
rr=ARG1

filename = sprintf("data/rfc6781.%s.data", rr)
outputname = sprintf("data/rfc6781.%s.png", rr)

title_str = sprintf("Percentage of gTLDs / ccTLDs with RRSIG lifetime following RFC 6781", rr)

set style data lines
set title title_str
set xlabel "Date"
set ylabel "Number of TLDs"
set xdata time
set timefmt '%Y-%m-%d'
set format y "%.1f"
set format x "%Y-%m-%d"
set xtics rotate by 45 right
set yrange [0:1]
#set logscale y
eps = 0

set terminal pngcairo size 1024,768
set output outputname

# ccTLD = red (solid, dotted, dashed)
set style line 1 lc rgb "red"   lt 1 lw 2  dt 1  # solid
set style line 2 lc rgb "red"   lt 2 lw 2  dt 2  # dotted
set style line 3 lc rgb "red"   lt 3 lw 2  dt 3  # dashed

# gTLD = blue (solid, dotted, dashed)
set style line 4 lc rgb "blue"  lt 1 lw 2  dt 1  # solid
set style line 5 lc rgb "blue"  lt 2 lw 2  dt 2  # dotted
set style line 6 lc rgb "blue"  lt 3 lw 2  dt 3  # dashed

plot filename using 1:($3/$2+eps) title "ccTLD too short"  w l ls 1, \
     filename using 1:($4/$2+eps) title "ccTLD ok"         w l ls 2, \
     filename using 1:($5/$2+eps) title "ccTLD too long"   w l ls 3, \
     filename using 1:($7/$6+eps) title "gTLD too short"   w l ls 4, \
     filename using 1:($8/$6+eps) title "gTLD ok"          w l ls 5, \
     filename using 1:($9/$6+eps) title "gTLD too long"    w l ls 6

#pause -1
