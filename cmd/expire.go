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
var expireCmd = &cobra.Command{
	Use:     "expire",
	Version: "0.0.1a",
	Short:   "get DNSSEC timing data for specific TLD and rr type",
	Long:    "get DNSSEC timing data for specific TLD and rr type",
	Run:     func(cmd *cobra.Command, args []string) { expireRun(args) },
}

func init() {
	// add the command to cobra
	rootCmd.AddCommand(expireCmd)

	// Use flags for viper values
	viper.BindPFlags(expireCmd.Flags())
}

func expireRun(args []string) {

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
	soaData, err := db.Query("SELECT RESOLVED,TLD,RRDATA FROM RRSIG JOIN RRDATA ON(RRSIG.SHA256=RRDATA.SHA256) WHERE RRTYPE=? ORDER BY RESOLVED,TLD", dns.TypeSOA)
	if err != nil {
		log.Fatalf("Could not query for SOA data %s", err)
	}
	defer soaData.Close() // Prepared statements take up server resources and should be closed after use.

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
		fmt.Printf("%s %s %d\n", resolved.Format(time.DateOnly), tld, rr.(*dns.SOA).Expire)
	}
}
