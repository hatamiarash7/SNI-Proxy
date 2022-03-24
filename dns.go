package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

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
	log.Info("Loading the domains from to a list")
	
	var lines [][]string
	var scanner *bufio.Scanner
	
	if strings.HasPrefix(Filename, "http://") || strings.HasPrefix(Filename, "https://") {
		log.Info("Domain list is a URL")

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		client := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}

		resp, err := client.Get(Filename)

		if err != nil {
			log.Fatal(err)
		}

		log.Info("Fetching URL : ", Filename)

		defer resp.Body.Close()
		scanner = bufio.NewScanner(resp.Body)
	} else {
		file, err := os.Open(Filename)

		if err != nil {
			log.Fatal(err)
		}

		log.Info("Loading File : ", Filename)

		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		lowerCaseLine := strings.ToLower(scanner.Text())
		lines = append(lines, strings.Split(lowerCaseLine, ","))
	}

	log.Infof("%s Loaded With %d Records", Filename, len(lines))

	return lines
}

// Create an external query
func externalQuery(question dns.Question, server string) (*dns.Msg, error) {
	client := new(dns.Client)
	msg := new(dns.Msg)
	
	msg.RecursionDesired = true
	msg.Id = dns.Id()
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = question
	in, _, err := client.Exchange(msg, fmt.Sprintf("%s:53", server))
	
	return in, err
}

// Parse query
func parseQ(msg *dns.Msg, ip string) {
	for _, question := range msg.Question {
		if !checkList(question.Name, routeList) {
			log.Printf("Bypassing %s\n", question.Name)

			in, err := externalQuery(question, *upstreamDNS)

			if err != nil {
				log.Println(err)
			}

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
