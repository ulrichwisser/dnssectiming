#
# usage: gnuplot -c gnuplot/lifet.plt TLD RRTYPE
#
tld=ARG1
rr=ARG2

filename = sprintf("data/lifet.%s.%s.data", tld, rr)
outputname = sprintf("data/lifet.%s.%s.png", tld, rr)
title_str = sprintf("RRSIG Lifetime of %s records for .%s", rr, tld)
title_lifetime = sprintf("RRSIG Lifetime of %s records", rr)

# compute y range
stats filename u 2 nooutput
MAX_LIFETIME=STATS_max
stats filename u 3 nooutput
MAX_EXPIRE=3*STATS_max
MAX_Y=MAX_LIFETIME
if (MAX_Y<MAX_EXPIRE) {
    MAX_Y=MAX_EXPIRE
}
MAX_Y=MAX_Y*1.1
set yrange [0:MAX_Y]

set style data lines
set title title_str
set xlabel "Date"
set ylabel "RRSIG Lifetime"
set xdata time
set timefmt '%Y-%m-%d'
set format y ""
set format x "%Y-%m-%d"
set xtics rotate by 45 right

set ytics()
set ytics add ( "1h" 3600, "1d" 84600, "7d" 604800, "14d" 1209600, "30d" 2592000, "100d" 8460000, "7000d" 604800000 )

# add tic for max value
# MAX_DAYS=MAX_Y/86400
# set ytics add ( sprintf("%dd", MAX_DAYS) MAX_Y)

set terminal png size 1024,768
set output outputname

plot filename using 1:2 title title_lifetime, filename using 1:3 title "SOA Expire", filename using 1:(3*$3) title "RFC 6781 Minimum Liftetime"
