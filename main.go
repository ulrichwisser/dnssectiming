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
package main

import (
	"os"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
	"bufio"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/miekg/dns"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"
    _ "github.com/go-sql-driver/mysql"

	"crypto/sha256"
)

const VERBOSE = "verbose"
const VERBOSE_QUIET int = 0
const VERBOSE_ERROR int = 1
const VERBOSE_WARNING int = 2
const VERBOSE_INFO int = 3
const VERBOSE_DEBUG int = 4
const VERBOSE_TRACE int = 5

const CONCURRENT = "concurrent"
const CONCURRENT_DEFAULT uint = 100

const DBCREDENTIALS = "dbcredentials"

const RESOLVERS = "resolvers"

const TIMEOUT time.Duration = 5 // seconds

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                   "dnssectiming [-v] [-c <number of concurrent threads>]  -r <resolver ip> <domain list file>",
	Version:               "0.0.1a",
	Short:                 "get dnssec timing information",
	Long:                  `get dnssec timing information`,
	Run:                   func(cmd *cobra.Command, args []string) { run(args) },
	DisableFlagsInUseLine: true,
	Args:                  cobra.MaximumNArgs(1),
}

func main() {
	// Use flags for viper values
	viper.BindPFlags(pflag.CommandLine)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	// Set default log handler
	log.SetHandler(text.New(os.Stderr))

	// Set defaults
	//
	// default log loglevel
	viper.SetDefault(VERBOSE, VERBOSE_QUIET)
	viper.SetDefault(CONCURRENT, CONCURRENT_DEFAULT)

	// define command line arguments
	rootCmd.Flags().CountP(VERBOSE, "v", "repeat for more verbose printouts")
	rootCmd.Flags().UintP(CONCURRENT, "c", CONCURRENT_DEFAULT, "config file (default is $HOME/.umpy)")
	rootCmd.Flags().StringSliceP(RESOLVERS, "r", []string{}, "resolver ip address")

	// Use flags for viper values
	viper.BindPFlags(rootCmd.Flags())

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Search config in home directory with name ".umpy" (without extension).
	viper.SetConfigName(".dnssect")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	// read in environment variables that match
	viper.SetEnvPrefix("DNSSECT")
	viper.AutomaticEnv()

}

func run(args []string) {

	// init log level
	switch viper.GetInt(VERBOSE) {
		case VERBOSE_QUIET:   	log.SetLevel(log.FatalLevel)
		case VERBOSE_ERROR:   	log.SetLevel(log.ErrorLevel)
		case VERBOSE_WARNING: 	log.SetLevel(log.WarnLevel)
		case VERBOSE_INFO:    	log.SetLevel(log.InfoLevel)
		case VERBOSE_DEBUG:   	log.SetLevel(log.DebugLevel)
		default: 				
			if viper.GetInt(VERBOSE) > VERBOSE_DEBUG {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.ErrorLevel)
			}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if len(viper.GetString("config")) > 0 {
			log.Fatalf("Error reading config file: '%s' %s",viper.GetString("config"), err.Error())
		}
		log.Info(err.Error())
	} else {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
	
	// check resolver list
	resolvers := getResolvers()
	log.Debugf("Using resolvers %v", resolvers)

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
        log.Fatal(err.Error())
    }
	log.Debug("DB OPEN")

	//
	// DOMAIN LIST
	//
	log.Debug("Start reading zone file")
	var domainlistfh *os.File
	if len(args) > 0 {
		var err error
		domainlistfh, err = os.Open(args[0])
		if err != nil {
			log.Errorf("Could not open domain list %s", args[0])
			log.Error(err.Error())
			os.Exit(5)
		}
	} else {
		domainlistfh = os.Stdin
	}

    scanner := bufio.NewScanner(domainlistfh)
    scanner.Split(bufio.ScanLines)


	// start concurrent resolving
	var wg sync.WaitGroup
	var threads = make(chan string, viper.GetInt(CONCURRENT))
	var answers = make(chan *dns.Msg, 1000)
	defer close(threads)

	// start listening for answers
	go saveAnswers(answers, &wg, db)

	resolver := 0 // loop over all resolvers

	for scanner.Scan() {
		domain := scanner.Text()
		log.Debugf("Read: %s", domain)
		if err := scanner.Err(); err != nil {
			log.Fatalf("Failed to read domain list: %s", err.Error())
		}
		wg.Add(1)
		threads <- "x"
		go resolve(dns.Fqdn(domain), resolvers[resolver], &wg, threads, answers)
		resolver = (resolver + 1) % len(resolvers)
	}
	wg.Wait()

	wg.Add(1) // this is for save answers
	close(answers)
	wg.Wait()

	// close the file we opened
	if len(args) > 0 {
		domainlistfh.Close()
	}
	log.Debug("Done reading domain list.")
}

// getResolvers will read the list of resolvers from /etc/resolv.conf
func getResolvers() []string {
	resolvers := make([]string, 0)

	rslice := viper.GetStringSlice(RESOLVERS)
	if len(rslice) == 0 {
		log.Fatal("No resolvers are given")
	}

	for _,r := range rslice {

		ip := net.ParseIP(r)
		if ip == nil {
			log.Fatalf("Could not parse resolver ip: %s", r)
		}

		ipstr := ip.String()
		if strings.ContainsAny(":", ipstr) {
			// IPv6 address
			ipstr = "[" + ipstr + "]:53"
		} else {
			// IPv4 address
			ipstr = ipstr + ":53"
		}
		resolvers = append(resolvers, ipstr)
	}
	if len(resolvers) == 0 {
		log.Fatal("No resolvers found.")
	}
	return resolvers
}

// resolv will send a query and return the result
func resolve(domain string, server string, wg *sync.WaitGroup, threads <-chan string, answers chan *dns.Msg) {
	defer log.Trace(fmt.Sprintf("Resolving %s using %s", domain, server)).Stop(nil)

	defer func() { _ = <-threads }()
	defer wg.Done()

	// Setting up query
	query := new(dns.Msg)
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.SetEdns0(1232, false)
	query.IsEdns0().SetDo()

	// Setting up resolver
	client := new(dns.Client)
	client.ReadTimeout = TIMEOUT * 1e9

	for _,rrtype := range []uint16{dns.TypeSOA, dns.TypeNS, dns.TypeDNSKEY, dns.TypeDS} {
		query.SetQuestion(domain, rrtype)

		// query until we get an answer
		for {
			// make the query and wait for answer
			r, _, err := client.Exchange(query, server)

			// check for errors
			if err != nil {
				log.Errorf("%-30s: Error resolving %s (server %s)", domain, err, server)
				continue
			}
			if r == nil {
				log.Errorf("%-30s: No answer (Server %s)", domain, server)
				continue
			}
			if r.Rcode != dns.RcodeSuccess {
				log.Errorf("%-30s: %s (Rcode %d, Server %s)", domain, dns.RcodeToString[r.Rcode], r.Rcode, server)
				break
			}	

			// we got an answer 
			answers <- r
			break
		}
	}
}

func saveAnswers(answers chan *dns.Msg, wg *sync.WaitGroup, db *sql.DB) {
	defer log.Trace("saving answers").Stop(nil)

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Could not start DB transaction %s", err)
	}
	defer tx.Rollback()

	stmt_rrdata, err := tx.Prepare("INSERT IGNORE INTO RRDATA(SHA256,RRDATA) VALUES(?,?)")
	if err != nil {
		log.Fatalf("Could not prepare insert into rrdata %s", err)
	}
	defer stmt_rrdata.Close() // Prepared statements take up server resources and should be closed after use.

	stmt_rrsig, err := tx.Prepare("INSERT INTO RRSIG(TLD,RRTYPE,SHA256,INCEPTION,EXPIRATION,SIG) VALUES(?,?,?,from_unixtime(?),from_unixtime(?),?)")
	if err != nil {
		log.Fatalf("Could not prepare insert into rrsig %s", err)
	}
	defer stmt_rrsig.Close() // Prepared statements take up server resources and should be closed after use.

	for msg := range answers {

		var rrsig *dns.RRSIG
		var rrdata []string
		for _,rr:= range msg.Answer {
			rrdata = append(rrdata, rr.String());
			if rr.Header().Rrtype == dns.TypeRRSIG && (rrsig == nil || rr.(*dns.RRSIG).Algorithm > rrsig.Algorithm) {
				rrsig = rr.(*dns.RRSIG)
			}
		}
		if rrsig == nil {
			log.Infof("%s is not signed. ", msg.Question[0].Name)
			continue
		}
		sort.Strings(rrdata) // sort is need to normalize strings, dns answers with round robin data
		rrdata_str := strings.Join(rrdata,"\n")
		sha256 := fmt.Sprintf("%x", sha256.Sum256([]byte(rrdata_str)))

		_,err = stmt_rrdata.Exec(sha256, rrdata_str)
		if err != nil {
			log.Fatalf("Writing to RRDATA failed %s", err)
		}

		_,err = stmt_rrsig.Exec(msg.Question[0].Name, msg.Question[0].Qtype, sha256, rrsig.Inception, rrsig.Expiration,"")
		if err != nil {
			log.Fatalf("Writing to RRSIG failed %s", err)
		}
		
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Could not prepare commit to DB %s", err)
	}
	wg.Done()
}