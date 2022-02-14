package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var routeList [][]string

// Check domain list for requested hostname
func checkList(domain string, domainList [][]string) bool {
	for _, item := range domainList {
		if len(item) == 2 {
			if item[1] == "suffix" {
				if strings.HasSuffix(domain, item[0]) {
					return true
				}
			} else if item[1] == "fqdn" {
				if domain == item[0] {
					return true
				}
			} else if item[1] == "prefix" {
				if strings.HasPrefix(domain, item[0]) {
					return true
				}
			}
		}
	}

	return false
}

// Load all given domains
func loadDomains(Filename string) [][]string {
	file, err := os.Open(Filename)
	handleError(err)
	log.Println("Reloading file : ", Filename)
	defer file.Close()
	var lines [][]string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, strings.Split(scanner.Text(), ","))
	}

	return lines
}

// Create an external query
func externalQuery(question dns.Question, server string) *dns.Msg {
	client := new(dns.Client)
	msg := new(dns.Msg)
	msg.RecursionDesired = true
	msg.Id = dns.Id()
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = question
	in, _, _ := client.Exchange(msg, fmt.Sprintf("%s:53", server))
	
	return in
}

// Parse query
func parseQ(msg *dns.Msg, ip string) {
	for _, question := range msg.Question {

		if !checkList(question.Name, routeList) {
			log.Printf("Bypassing %s\n", question.Name)
			in := externalQuery(question, *upstreamDNS)
			msg.Answer = append(msg.Answer, in.Answer...)
		} else {
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", question.Name, ip))

			if err == nil {
				log.Printf("Routing %s\n", question.Name)
				msg.Answer = append(msg.Answer, rr)

				return
			}
		}

	}

}
