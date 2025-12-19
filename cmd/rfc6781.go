/*
Copyright © 2023 Ulrich Wisser <ulrich@wisser.se>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/miekg/dns"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/apex/log"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// rootCmd represents the base command when called without any subcommands
var rfc6781Cmd = &cobra.Command{
	Use:     "rfc6781",
	Version: "0.0.1a",
	Short:   "get DNSSEC timing data all TLD for RFC 6781 recommendations",
	Long:   "get DNSSEC timing data all TLD for RFC 6781 recommendations",
	Run:     func(cmd *cobra.Command, args []string) { 
		// debug command line arguments
		log.Debug("Flags:")
		cmd.Flags().VisitAll(func(f *pflag.Flag) { log.Debugf("  %s = %s (changed=%v)\n", f.Name, f.Value, f.Changed) })

		rrFromFlag, _ := cmd.Flags().GetString(RR)
		tldFromFlag, _ := cmd.Flags().GetString(TLD)

		log.Debugf("rr  from flag: %s", rrFromFlag)
		log.Debugf("tld from flag: %s", tldFromFlag)

		log.Debugf("rr  from viper: %s", viper.GetString(RR))
		log.Debugf("tld from viper: %s", viper.GetString(TLD))

		// now run the command
		rfc6781Run(args) 
	},
}

func init() {
	// add the command to cobra
	rootCmd.AddCommand(rfc6781Cmd)

	// define command line arguments
	rfc6781Cmd.Flags().StringP(RR, RR_SHORT, RR_DEFAULT, RR_DESCRIPTION)

	// Use flags for viper values
	viper.BindPFlags(rfc6781Cmd.Flags())
}

func rfc6781Run(args []string) {

	// check RR command line arguments
	var rrtype uint16 = 0
	var rr_str = viper.GetString(RR)
	if rr_str == "NS" {
		rrtype = dns.TypeNS
	}
	if rr_str == "DNSKEY" {
		rrtype = dns.TypeDNSKEY
	}
	if rrtype == 0 {
		log.Fatal("No valid RR type was given. Must be one of NS or DNSKEY")
	}

	// open database
	if viper.GetString(DBCREDENTIALS) == "" {
		log.Fatal("No DB credentials given.")
	}
	db, err := sql.Open("mysql", viper.GetString(DBCREDENTIALS))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping DB %s", err.Error())
	}
	log.Debug("DB OPEN")

	//
	// Get SOA Expire
	//
	log.Debug("Start SQL Expire")
	soaData, err := db.Query("SELECT RESOLVED,TLD,RRDATA FROM RRSIG JOIN RRDATA ON(RRSIG.SHA256=RRDATA.SHA256) WHERE RRTYPE=? ORDER BY RESOLVED,TLD ", dns.TypeSOA)
	if err != nil {
		log.Fatalf("Could not query for SOA data %s", err)
	}
	defer soaData.Close() // Prepared statements take up server resources and should be closed after use.

	var soaByDateTLD map[time.Time]map[string]uint32 = make(map[time.Time]map[string]uint32, 0)
	for soaData.Next() {
		var resolved time.Time
		var tld string
		var rrdata string
		err := soaData.Scan(&resolved, &tld, &rrdata)
		if err != nil {
			log.Fatalf("Error scanning SOA data %s", err)
		}
		rr, err := dns.NewRR(rrdata)
		if err != nil {
			log.Fatalf("Could not parse SOA record >%<\n%s", rrdata, err)
		}

		// prepare data structure
		if _, ok := soaByDateTLD[resolved]; !ok {
			soaByDateTLD[resolved] = make(map[string]uint32, 0)
		}

		// save data
		soaByDateTLD[resolved][tld] = rr.(*dns.SOA).Expire
	}

	//
	// Get lifetime
	//
	log.Debug("Start SQL")
	rrData, err := db.Query("SELECT RESOLVED,TLD,EXPIRATION FROM RRSIG WHERE RRTYPE=? ORDER BY RESOLVED,TLD", rrtype)
	if err != nil {
		log.Fatalf("Could not query for %s data %s", rr_str, err)
	}
	defer rrData.Close() // Prepared statements take up server resources and should be closed after use.

	var failedByDateTLD map[time.Time]map[string]int = make(map[time.Time]map[string]int, 0)
	for rrData.Next() {
		var resolved time.Time
		var tld string
		var expiration time.Time
		err := rrData.Scan(&resolved, &tld, &expiration)
		if err != nil {
			log.Fatalf("Error scanning RR data. %s", err)
		}
		expire, ok := soaByDateTLD[resolved][tld]
		if !ok {
			continue
		}
		lifetime := expiration.UTC().Unix() - resolved.UTC().Unix()

		// prepare data structure
		if _, ok := failedByDateTLD[resolved]; !ok {
			failedByDateTLD[resolved] = make(map[string]int, 0)
		}

		// save data
		switch {
        case int64(expire) <  3 * lifetime:	failedByDateTLD[resolved][tld] = -1 // too short
        case int64(expire) <= 4 * lifetime: failedByDateTLD[resolved][tld] =  0 // correct
        case 4 * lifetime <  int64(expire):	failedByDateTLD[resolved][tld] =  1 // too long
		}
	}

	//
	// compute daily summary
	//
	type dateStats struct {
		ccShort   int
		ccOK      int
		ccLong    int
		gtldShort int
		gtldOK    int
		gtldLong  int
	}
	var statsByDate map[time.Time]*dateStats = make(map[time.Time]*dateStats, 0)
	for resolved := range failedByDateTLD {
		for tld := range failedByDateTLD[resolved] {
			// prepare data structure
			if _, ok := statsByDate[resolved]; !ok {
				statsByDate[resolved] = &dateStats{}
			}

			if len(tld) == 3 {
				switch failedByDateTLD[resolved][tld] {
				case -1: statsByDate[resolved].ccShort++
				case  0: statsByDate[resolved].ccOK++
				case  1: statsByDate[resolved].ccLong++
				}
			} else {
				switch failedByDateTLD[resolved][tld] {
				case -1: statsByDate[resolved].gtldShort++
				case  0: statsByDate[resolved].gtldOK++
				case  1: statsByDate[resolved].gtldLong++
				}
			}
		}
	}

	// get sorted lists of resolved
	var resolvedList []time.Time
	for resolved := range soaByDateTLD {
		resolvedList = append(resolvedList, resolved)
	}
	sort.Slice(resolvedList, func(i, j int) bool { return resolvedList[i].Before(resolvedList[j]) })

	// output final result
	for _, resolved := range resolvedList {
		fmt.Printf("%s %d %d %d %d %d %d\n", resolved.Format(time.DateOnly), statsByDate[resolved].ccShort, statsByDate[resolved].ccOK, statsByDate[resolved].ccLong, statsByDate[resolved].gtldShort, statsByDate[resolved].gtldOK, statsByDate[resolved].gtldLong)
	}

}
