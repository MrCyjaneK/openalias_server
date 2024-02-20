package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

var records = map[string]string{}

func init() {
	b, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(b, &records)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("init(): Loaded", len(records), "records")
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			rr, err := dns.NewRR(fmt.Sprintf("%s A 205.185.118.53", q.Name))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			} else {
				log.Println(err)
			}
		case dns.TypeTXT:
			log.Printf("Query for %s\n", q.Name)
			record := records[strings.ToLower(q.Name)]
			if record != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s TXT \"oa1:xmr recipient_address=%s\";", q.Name, record))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				} else {
					log.Println(err)
				}
			} else {
				log.Println("Result for", q.Name, "not found")
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	// attach request handler func
	dns.HandleFunc(".", handleDnsRequest)

	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
