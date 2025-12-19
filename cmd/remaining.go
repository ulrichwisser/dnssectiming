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
var remainingCmd = &cobra.Command{
	Use:     "remaining --rr <NS|DNSKEY>",
	Version: "0.0.1a",
	Short:   "get DNSSEC timing data for rr type",
	Long:    "get DNSSEC timing data for rr type",
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
		remainingRun(args) 
	},
}

func init() {
	// add the command to cobra
	rootCmd.AddCommand(remainingCmd)
}

func remainingRun(args []string) {

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
	// Get lifetime
	//
	rrData, err := db.Query("SELECT RESOLVED,TLD,EXPIRATION FROM RRSIG WHERE RRTYPE=? ORDER BY RESOLVED,TLD ", rrtype)
	if err != nil {
		log.Fatalf("Could not query for %s data %s", rr_str, err)
	}
	defer rrData.Close() // Prepared statements take up server resources and should be closed after use.

	//
	const under1d int64 = 86400
	const under3d int64 = 259200
	const under7d int64 = 604800
	const under14d int64 = 1209600
	const under35d int64 = 3024000

	var remaining map[time.Time]map[int64]uint = make(map[time.Time]map[int64]uint, 0)
		
	for rrData.Next() {
		var resolved time.Time
		var tld string
		var expiration time.Time
		err := rrData.Scan(&resolved, &tld, &expiration)
		if err != nil {
			log.Fatalf("Error scanning RR data. %s", err)
		}
		lifetime := expiration.UTC().Unix() - resolved.UTC().Unix()
		if lifetime < 0{
			// Signature too old, no meaningful statistic can be extracted
			continue
		}

		// prepare data structure
		if _, ok := remaining[resolved]; !ok {
			remaining[resolved] = make(map[int64]uint, 0)
		}

		// save data
		switch {
		case lifetime <under1d: remaining[resolved][under1d] += 1
								 log.Infof("TLD %s Lifetime %d (expiration %s (%d), resolved %s (%d))\n",tld,lifetime,expiration.Format(time.DateTime),expiration.UTC().Unix(), resolved.Format(time.DateTime),resolved.UTC().Unix())	
		case lifetime <under3d: remaining[resolved][under3d] += 1
		case lifetime <under7d: remaining[resolved][under7d] += 1
		case lifetime <under14d: remaining[resolved][under14d] += 1
		case lifetime <under35d: remaining[resolved][under35d] += 1
		}
	}

	// get sorted lists of resolved
	var resolvedList []time.Time
	for resolved := range remaining {
		resolvedList = append(resolvedList, resolved)
	}
	sort.Slice(resolvedList, func(i, j int) bool { return resolvedList[i].Before(resolvedList[j]) })

	// output final result
	for _, resolved := range resolvedList {
		fmt.Printf("%s %d %d %d %d %d\n", resolved.Format(time.DateOnly), remaining[resolved][under1d], remaining[resolved][under3d], remaining[resolved][under7d], remaining[resolved][under14d], remaining[resolved][under35d])
	}

}
