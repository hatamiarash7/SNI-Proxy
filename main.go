package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/miekg/dns"
)

var bindIP = flag.String("BIP", "0.0.0.0", "Bind to an IP Address")
var publicIP = flag.String("PIP", "", "Public IP of server")
var domainList = flag.String("list", "", "Domain list")
var refreshInterval = flag.Duration("domainListRefreshInterval", 60 * time.Second, "Interval to reload domains list")
var allDomains = flag.Bool("all", false, "Do for All Domains")
var upstreamDNS = flag.String("upstream", "1.1.1.1", "Upstream DNS")

// Redirect all HTTP traffic to HTTPS
func get80(writer http.ResponseWriter, req *http.Request) {
	http.Redirect(writer, req, "https://" + req.Host + req.RequestURI, 302)
}

func pipe(connection1 net.Conn, connection2 net.Conn) {
	channel1 := getChannel(connection1)
	channel2 := getChannel(connection2)

	for {
		select {
		case b1 := <-channel1:
			if b1 == nil {
				return
			}

			connection2.Write(b1)
		case b2 := <-channel2:
			if b2 == nil {
				return
			}

			connection1.Write(b2)
		}
	}
}

func getChannel(connection net.Conn) chan []byte {
	channel := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := connection.Read(b)

			if n > 0 {
				result := make([]byte, n)
				copy(result, b[:n])
				channel <- result
			}

			if err != nil {
				channel <- nil
				break
			}
		}
	}()

	return channel
}

func lookupDomain4(domain string) (net.IP, error) {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	rAddrDns, err := externalQuery(dns.Question{Name: domain, Qtype: dns.TypeA, Qclass: dns.ClassINET}, *upstreamDNS)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if rAddrDns.Answer[0].Header().Rrtype == dns.TypeCNAME {
		return lookupDomain4(rAddrDns.Answer[0].(*dns.CNAME).Target)
	}

	if rAddrDns.Answer[0].Header().Rrtype == dns.TypeA {
		return rAddrDns.Answer[0].(*dns.A).A, nil
	}

	return nil, fmt.Errorf("Unknown type")
}

// Handle HTTPS traffic
func get443(packet net.Conn) error {
	packetDataBytes := make([]byte, 5000)
	
	n, err := packet.Read(packetDataBytes) // Read packet
	if err != nil {
		log.Println(err)
		return err
	}					

	sni, err := getHost(packetDataBytes) // Get hostname
	if err != nil {
		log.Println(err)
		return err
	}							
	
	rAddr, err := lookupDomain4(sni) // Lookup domain
	if err != nil {
		log.Println(err)
		return err
	}
	
	target, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: rAddr, Port: 443})
	if err != nil {
		log.Println("Couldn't connect to target", err)
		packet.Close()
		return err
	}
	
	defer target.Close()
	target.Write(packetDataBytes[:n])
	pipe(packet, target)

	return nil
}

// Handle DNS requests ( We have a DNS server here )
func get53(writer dns.ResponseWriter, req *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(req)
	msg.Compress = false

	switch req.Opcode {
	case dns.OpcodeQuery:
		parseQ(msg, *publicIP) // Parse query
	}

	writer.WriteMsg(msg)
}

// Handle errors
func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Run a HTTP server
func runHttp() {
	http.HandleFunc("/", get80)

	server := http.Server{
		Addr: ":80",
	}

	server.ListenAndServe()
}

// Run a HTTPS server
func runHttps() {
	l, err := net.Listen("tcp", *bindIP+":443")
	handleError(err)
	defer l.Close()

	for {
		conn, err := l.Accept()
		handleError(err)
		go get443(conn)
	}
}

// Run a DNS server
func runDns() {
	dns.HandleFunc(".", get53)
	server := &dns.Server{Addr: ":53", Net: "udp"}				// Create server
	log.Printf("Start DNS server on 0.0.0.0:53 -- listening")
	err := server.ListenAndServe()
	defer server.Shutdown()

	if err != nil {
		log.Fatalf("Failed to start DNS server : %s\n ", err.Error())
	}
}

func main() {

	flag.Parse()

	if *domainList == "" || *publicIP == "" || *upstreamDNS == "" {
		log.Fatalln("Give me `-list` and `-PIP`")
	}

	// Run servers
	go runHttp()
	go runHttps()
	go runDns()

	// Load domains list
	routeList = loadDomains(*domainList)
	for range time.NewTicker(*refreshInterval).C {
		routeList = loadDomains(*domainList)
	}
}
