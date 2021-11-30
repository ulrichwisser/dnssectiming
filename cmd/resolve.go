// Bulkdns takes a file with one domain name per line as input and will request your configured resolver(s) from /etc/resolv.conf for the nameservers of these domains.
// There are some command line arguments
//
// -v (--verbose) will print out some debug information and query results
//
// -c <int> number of concurrent queries
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

// getResolvers will read the list of resolvers from /etc/resolv.conf
func getResolvers() []string {
	resolvers := make([]string, 0)

	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if conf == nil {
		fmt.Printf("Cannot initialize the local resolver: %s\n", err)
		os.Exit(1)
	}
	for _, server := range conf.Servers {
		resolvers = append(resolvers, getResolverDial(server))
	}
	if len(resolvers) == 0 {
		fmt.Println("No resolvers found.")
		os.Exit(5)
	}
	return resolvers
}

func getResolverDial(resolver string) string {
	if strings.ContainsAny(":", resolver) {
		// IPv6 address
		return "[" + resolver + "]:53"
	}
	// IPv4 address
	return resolver + ":53"
}

// resolv will send a query and return the result
func resolve(qname string, qtype uint16, server string) *dns.Msg {
	if verbose > 0 {
		log.Printf("Resolving %s %s using %s\n", qname, dns.TypeToString[qtype], server)
	}

	// Setting up query
	query := new(dns.Msg)
	query.SetQuestion(qname, qtype)
	query.SetEdns0(EDNS0SIZE, false)
	query.IsEdns0().SetDo()
	query.RecursionDesired = true

	// Setting up resolver
	client := new(dns.Client)
	client.ReadTimeout = TIMEOUT * 1e9

	// make the query and wait for answer
	r, _, err := client.Exchange(query, server)

	// check for errors
	if err != nil {
		log.Printf("%-30s: Error resolving %s (server %s)\n", qname, err, server)
		return resolve(qname, qtype, server)
	}
	if r == nil {
		log.Printf("%-30s: No answer (Server %s)\n", qname, server)
		return resolve(qname, qtype, server)
	}
	if r.Rcode != dns.RcodeSuccess {
		log.Printf("%-30s: %s (Rcode %d, Server %s)\n", qname, dns.RcodeToString[r.Rcode], r.Rcode, server)
	}
	return r
}
