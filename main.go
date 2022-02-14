package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

var bindIP = flag.String("BIP", "0.0.0.0", "Bind to an IP Address")
var publicIP = flag.String("PIP", "", "Public IP of server")
var domainList = flag.String("list", "", "Domain list")
var allDomains = flag.Bool("all", false, "Do for All Domains")
var upstreamDNS = flag.String("upstream", "1.1.1.1", "Upstream DNS")

// Redirect all HTTP traffic to HTTPS
func get80(writer http.ResponseWriter, req *http.Request) {
	http.Redirect(writer, req, "https://" + req.Host + req.RequestURI, 302)
}

// Handle HTTPS traffic
func get443(packet net.Conn) error {
	packetDataBytes := make([]byte, 5000)
	packet.Read(packetDataBytes)								// Read packet
	sni, _ := getHost(packetDataBytes)							// Get hostname
	destipList, _ := net.LookupIP(sni)							// Get destination IP's from request
	destip := destipList[0]										// We need the first one
	target, err := net.Dial("tcp", destip.String() + ":443")	// Check reachability for target

	if err != nil {
		log.Println("Couldn't connect to target", err)
		packet.Close()

		return err
	}

	defer target.Close()

	go func() { io.Copy(target, packet) }()
	go func() { io.Copy(packet, target) }()

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
	l, err := net.Listen("tcp", ":443")
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

	// reload domain's list every 1 min
	timeticker := time.Tick(60 * time.Second)
	routeList = loadDomains(*domainList)

	for {
		select {
		case <-timeticker:
			routeList = loadDomains(*domainList)
		}
	}
}
