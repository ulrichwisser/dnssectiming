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
	"strings"
	"time"

	"github.com/miekg/dns"
)

//
// AXFR a zone
func axfr(zone string, server string, port uint) <-chan dns.RR {

	// Setting up transfer
	transfer := &dns.Transfer{DialTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}

	// Setting up query
	query := new(dns.Msg)
	query.RecursionDesired = false
	query.Question = make([]dns.Question, 1)
	query.SetQuestion(dns.Fqdn(zone), dns.TypeAXFR)

	// add port
	if strings.ContainsAny(":", server) {
		// IPv6 address
		server = fmt.Sprintf("[%s]:%d", server, port)
	} else {
		// IPv4 address
		server = fmt.Sprintf("%s:%d", server, port)
	}

	// start transfer
	channel, err := transfer.In(query, server)
	if err != nil {
		log.Fatal("error starting AXFR", err)
	}

	// prepare output channel
	c := make(chan dns.RR, 100)

	// translate transfer Envelope to dns.RR
	go func() {
		for env := range channel {
			if env.Error != nil {
				log.Fatal("FATAL ERROR reading data from AXFR: ", env.Error)
			}
			for _, rr := range env.RR {
				c <- rr
			}
		}
		close(c)
	}()

	// return
	return c
}
