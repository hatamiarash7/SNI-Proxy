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

const maxLengthBytes = 5000

var bindIP = flag.String("BIP", "0.0.0.0", "Bind to an IP Address")
var publicIP = flag.String("PIP", "", "Public IP of server")
var domainList = flag.String("list", "", "Domain list")
var allDomains = flag.Bool("all", false, "Do for All Domains")
var upstreamDNS = flag.String("upstream", "1.1.1.1", "Upstream DNS")

func get80(writer http.ResponseWriter, req *http.Request) {
	http.Redirect(writer, req, "https://" + req.Host + req.RequestURI, 302)
}

func get443(packet net.Conn) error {
	packetDataBytes := make([]byte, maxLengthBytes)
	packet.Read(packetDataBytes)
	sni, _ := getHost(packetDataBytes)
	destipList, _ := net.LookupIP(sni)
	destip := destipList[0]
	target, err := net.Dial("tcp", destip.String()+":443")

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

func get53(writer dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Compress = false

	switch req.Opcode {
	case dns.OpcodeQuery:
		parseQ(m, *publicIP)
	}

	writer.WriteMsg(m)
}

func runHttp() {
	http.HandleFunc("/", get80)

	server := http.Server{
		Addr: ":80",
	}

	server.ListenAndServe()
}

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

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

func runDns() {
	dns.HandleFunc(".", get53)
	server := &dns.Server{Addr: ":53", Net: "udp"}
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

	go runHttp()
	go runHttps()
	go runDns()

	timeticker := time.Tick(60 * time.Second)
	routeDomainList = loadDomains(*domainList)

	for {
		select {
		case <-timeticker:
			routeDomainList = loadDomains(*domainList)
		}
	}
}
