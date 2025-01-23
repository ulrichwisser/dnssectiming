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
	"time"

	"github.com/miekg/dns"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/apex/log"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// rootCmd represents the base command when called without any subcommands
var lifetimeCmd = &cobra.Command{
	Use:     "lifetime --tld <name> --rr <NS|DNSKEY>",
	Version: "0.0.1a",
	Short:   "get DNSSEC timing data for specific TLD and rr type",
	Long:    "get DNSSEC timing data for specific TLD and rr type",
	Run:     func(cmd *cobra.Command, args []string) { lifetimeRun(args) },
}

func init() {
	// add the command to cobra
	rootCmd.AddCommand(lifetimeCmd)

	// define command line arguments
	lifetimeCmd.Flags().StringP(TLD, TLD_SHORT, TLD_DEFAULT, TLD_DESCRIPTION)
	lifetimeCmd.Flags().StringP(RR, RR_SHORT, RR_DEFAULT, RR_DESCRIPTION)

	// Use flags for viper values
	viper.BindPFlags(lifetimeCmd.Flags())
}

func lifetimeRun(args []string) {

	// check TLD command line arguments
	var tld = viper.GetString(TLD)
	if len(tld) < 2 {
		log.Fatal("No valid TLD value was given")
	}
	tld = dns.Fqdn(tld)

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
	soaData, err := db.Query("SELECT RESOLVED,RRDATA FROM RRSIG JOIN RRDATA ON(RRSIG.SHA256=RRDATA.SHA256) WHERE TLD=? AND RRTYPE=? ORDER BY RESOLVED ", tld, dns.TypeSOA)
	if err != nil {
		log.Fatalf("Could not query for SOA data %s", err)
	}
	defer soaData.Close() // Prepared statements take up server resources and should be closed after use.

	var soaByDate map[time.Time]uint32 = make(map[time.Time]uint32, 0)
	for soaData.Next() {
		var resolved time.Time
		var rrdata string
		err := soaData.Scan(&resolved, &rrdata)
		if err != nil {
			log.Fatalf("Error scanning SOA data %s", err)
		}
		rr, err := dns.NewRR(rrdata)
		if err != nil {
			log.Fatalf("Could not parse SOA record >%<\n%s", rrdata, err)
		}
		soaByDate[resolved] = rr.(*dns.SOA).Expire
	}

	//
	// Get lifetime
	//
	rrData, err := db.Query("SELECT RESOLVED,EXPIRATION FROM RRSIG WHERE TLD=? AND RRTYPE=? ORDER BY RESOLVED ", tld, rrtype)
	if err != nil {
		log.Fatalf("Could not query for %s data %s", rr_str, err)
	}
	defer rrData.Close() // Prepared statements take up server resources and should be closed after use.

	for rrData.Next() {
		var resolved time.Time
		var expiration time.Time
		err := rrData.Scan(&resolved, &expiration)
		if err != nil {
			log.Fatalf("Error scanning RR data. %s", err)
		}
		expire, ok := soaByDate[resolved]
		if !ok {
			continue
		}
		lifetime := expiration.UTC().Unix() - resolved.UTC().Unix()
		fmt.Printf("%s %d %d\n", resolved.Format(time.DateOnly), lifetime, expire)
	}

}
