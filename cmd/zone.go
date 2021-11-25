/*
Copyright © 2021 Ulrich Wisser <ulrich@wisser.se>

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
	"log"
	"os"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	TIMEOUT    time.Duration = 5 // seconds
	CONCURRENT uint          = 3
	EDNS0SIZE  uint16        = 1232
)

// zoneCmd represents the zone command
var zoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runZone,
}

func init() {
	rootCmd.AddCommand(zoneCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// zoneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// zoneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	zoneCmd.Flags().IntP("maxdomains", "m", 0, "Stop parsing after this number of domains")

	// Here you will define your flags and configuration settings.
	viper.BindPFlags(zoneCmd.Flags())
}

// save list of domain details
type DomainList map[string][]dns.RR

func runZone(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Println("zone name has to be given")
		os.Exit(5)
	}

	zone := args[0]

	if len(viper.GetString(zone)) == 0 {
		log.Printf("Zone %s not found in config\n", zone)
	}
	server := viper.GetString(zone)

	fetchzone := dns.Fqdn(zone)
	if zone == "root" {
		fetchzone = "."
	}

	log.Printf("Zone %s will be fetched as %s with AXFR from %s\n", zone, fetchzone, server)

	rrchan := axfr(fetchzone, server, 53)

	domainlist := make(DomainList, 0)
	var dlaccess sync.Mutex

	// syncing of go routines
	var wg sync.WaitGroup
	var threads = make(chan string, CONCURRENT)

	// rotate between all resolvers
	resolvers := getResolvers()
	resolver := 0

	for rr := range rrchan {
		if rr.Header().Name == fetchzone {
			continue
		}
		if rr.Header().Rrtype != dns.TypeNS {
			continue
		}

		dlaccess.Lock()
		if _, ok := domainlist[rr.Header().Name]; !ok {
			domainlist[rr.Header().Name] = make([]dns.RR, 0)
			dlaccess.Unlock()
			wg.Add(1)
			threads <- "x"
			go getRrlist(&domainlist, &dlaccess, &wg, threads, resolvers[resolver], rr.Header().Name)
			resolver = (resolver + 1) % len(resolvers)
		} else {
			dlaccess.Unlock()
		}

		if 0 < viper.GetInt("maxdomains") && viper.GetInt("maxdomains") <= len(domainlist) {
			break
		}

	}

	wg.Wait()
	close(threads)

	if verbose > 0 {
		log.Printf("%d domains found.\n", len(domainlist))
	}

	for _, d := range domainlist {
		for _, rr := range d {
			fmt.Println(rr.String())
		}
	}
}

func getRrlist(dl *DomainList, dla *sync.Mutex, wg *sync.WaitGroup, threads <-chan string, server string, domain string) {
	if verbose > 0 {
		log.Printf("getRrlist(%s, %s)\n", server, domain)
	}

	// cleanup when we are done
	defer func() { _ = <-threads }()
	defer wg.Done()

	// get data
	if verbose > 0 {
		log.Printf("getRrlist: starting to resolve NS for %s\n", domain)
	}
	soa := resolve(domain, dns.TypeNS, server)
	if verbose > 0 {
		log.Printf("getRrlist: starting to resolve DNSKEY for %s\n", domain)
	}
	dnskey := resolve(domain, dns.TypeDNSKEY, server)
	if verbose > 0 {
		log.Printf("getRrlist: starting to resolve DS for %s\n", "nic."+domain)
	}
	nic := resolve("nic."+domain, dns.TypeDS, server)

	// save data to list
	dla.Lock()
	d := (*dl)[domain]
	for _, rr := range soa.Answer {
		d = append(d, rr)
	}
	for _, rr := range dnskey.Answer {
		d = append(d, rr)
	}
	for _, rr := range nic.Answer {
		d = append(d, rr)
	}
	(*dl)[domain] = d
	dla.Unlock()
}
