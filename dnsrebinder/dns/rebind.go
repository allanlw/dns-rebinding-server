// Copyright 2011 Miek Gieben. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Reflect is a small name server which sends back the IP address of its client, the
// recursive resolver.
// When queried for type A (resp. AAAA), it sends back the IPv4 (resp. v6) address.
// In the additional section the port number and transport are shown.
//
// Basic use pattern:
//
//	dig @localhost -p 8053 whoami.miek.nl A
//
//	;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 2157
//	;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1
//	;; QUESTION SECTION:
//	;whoami.miek.nl.			IN	A
//
//	;; ANSWER SECTION:
//	whoami.miek.nl.		0	IN	A	127.0.0.1
//
//	;; ADDITIONAL SECTION:
//	whoami.miek.nl.		0	IN	TXT	"Port: 56195 (udp)"
//
// Similar services: whoami.ultradns.net, whoami.akamai.net. Also (but it
// is not their normal goal): rs.dns-oarc.net, porttest.dns-oarc.net,
// amiopen.openresolvers.org.
//
// Original version is from: Stephane Bortzmeyer <stephane+grong@bortzmeyer.org>.
//
// Adapted to Go (i.e. completely rewritten) by Miek Gieben <miek@miek.nl>.
package dns

import (
	"dnsrebinder/http"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"net"
	"strconv"
	"strings"
)

var answerState = make(map[string]int) //Maps a requestid to the number of times its been hit
var (
	localhost    = net.ParseIP("127.0.0.1")
	domain_name  = ""
	http_ip      = ""
	capture_ip   = ""
	hitThreshold = 5
	compress     = flag.Bool("compress", false, "compress replies")
)
var HTTPState map[string]int

func makeID(args map[string]string) (string, string) {
	mode := ""
	id := ""
	ip1, ok_ip1 := args["ip1"]
	if !ok_ip1 {
		ip1 = capture_ip
		args["ip1"] = ip1
	}
	ip2, ok_ip2 := args["ip2"]
	uid, ok_uid := args["uid"]
	cidr, ok_cidr := args["cidr"]
	_, ok_httpstate := args["httpstate"]
	if ok_ip2 && ok_uid && !ok_httpstate {
		//Check for ip1,ip2 arguments
		mode = "onetwo"
		id = (ip1 + ip2 + uid)
	} else if ok_uid && ok_cidr {
		// Check for cidr range arguments
		mode = "cidr"
		id = (ip1 + cidr + uid)
	} else if ok_ip2 && ok_uid && ok_httpstate {
		// Check for cidr range arguments
		mode = "http_aware"
		id = ("h" + ip1 + ip2 + uid)
	} else if ok_uid {
		mode = "basic"
		id = (ip1 + "127.0.0.1" + uid)
	}

	// Catch queries that didn't match out to anything
	if id == "" {
		id = "err"
	}
	if mode == "" {
		mode = "basic"
	}
	glog.Info("Failed to parse hostname arg map:", args)
	return id, mode
}
func makeAnswer(name string, args map[string]string) net.IP {
	id, mode := makeID(args)
	hits, ok := answerState[id]
	localHitThreshold := hitThreshold
	localHitThresholdStr, ok := args["hits"]
	if !ok {
		localHitThreshold = hitThreshold
	} else {
		tmp, err := strconv.Atoi(localHitThresholdStr)
		if err == nil {
			localHitThreshold = tmp
		}
	}
	glog.Infof("Building Answer for Request ID:%s\n", id)
	glog.Infof("Current hits for ID %s: %d\n", id, hits)
	if !ok { // Never seen it before
		answerState[id] = 0
		hits = 0
	}
	answerState[id] += 1
	switch mode {
	default:
		fallthrough
	case "basic":
		glog.Infof("Basic mode: Current hits: %d, Threshold for returning localhost: %d\n", hits, localHitThreshold)
		if hits < localHitThreshold {
			ipstr, ok := args["ip1"]
			if !ok {
				glog.Infof("Failed to parse ip1 from hostname, defaulting to 127.0.0.1\n")
				return localhost
			}
			glog.Infof("Hits %d under threshold %d, returning %s\n", hits, localHitThreshold, ipstr)
			return net.ParseIP(strings.Replace(ipstr, "x", ".", -1))
		} else {
			glog.Infof("Hits %d over threshold %d, returning localhost\n", hits, localHitThreshold)
			return localhost
		}
	case "cidr":
		glog.Infof("Cidr mode: current hits:%d\n", hits)
		cidr, ok := args["cidr"]
		if !ok {
			glog.Infof("Failed to parse cidr from hostname, defaulting to 127.0.0.1\n")
			return localhost
		}
		cidr = strings.Replace(cidr, "x", ".", -1)
		cidr = strings.Replace(cidr, "s", "/", -1)
		ip, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			glog.Infof("Failed to parse cidr from hostname, defaulting to 127.0.0.1\n")
			return localhost
		}
		returnIP := intToIpv4(ipv4ToInt(ip.To4()) + uint32(hits))
		glog.Infof("Hits: %d, next ip: %s\n", hits, returnIP.String())
		if ipnet.Contains(returnIP) {
			return returnIP
		} else {
			glog.Infof("Reached end of CIDR walk, returning 127.0.0.1\n")
			return localhost
		}
	case "onetwo":
		glog.Infof("ip1,ip2 mode: Current hits: %d, Threshold for returning localhost: %d\n", hits, localHitThreshold)
		if hits < localHitThreshold {
			ipstr, ok := args["ip1"]
			if !ok {
				glog.Infof("Failed to parse ip1 from hostname, defaulting to 127.0.0.1\n")
				return localhost
			}

			glog.Infof("Hits %d under threshold %d, returning %s\n", hits, localHitThreshold, ipstr)
			return net.ParseIP(strings.Replace(ipstr, "x", ".", -1))
		} else {
			ipstr, ok := args["ip2"]
			if !ok {
				glog.Infof("Failed to parse ip2 from hostname, defaulting to 127.0.0.1\n")
				return localhost
			}

			glog.Infof("Hits %d over threshold %d, returning %s\n", hits, localHitThreshold, ipstr)
			return net.ParseIP(strings.Replace(ipstr, "x", ".", -1))
		}
	case "http_aware":
		glog.Infof("HTTP reactive mode\n")
		if strings.HasSuffix(name, ".") {
			name = name[:len(name)-1]
		}
		_, http_ok := HTTPState[name]
		glog.Infof("HTTP State Map Size: %d", len(HTTPState))
		if !http_ok {
			ipstr, ok := args["ip1"]
			if !ok {
				glog.Infof("Failed to parse ip1 from hostname, defaulting to 127.0.0.1\n")
				return localhost
			}

			glog.Infof("Payload hasn't been delivered for %s yet (not present in HTTP state), returning ip1: %s\n", name, ipstr)
			return net.ParseIP(strings.Replace(ipstr, "x", ".", -1))
		} else {

			ipstr, ok := args["ip2"]
			if !ok {

				glog.Infof("Failed to parse ip2 from hostname, defaulting to 127.0.0.1\n")
				return localhost
			}

			glog.Infof("Payload was delivered for %s! Returning ip2: %s\n", name, ipstr)
			return net.ParseIP(strings.Replace(ipstr, "x", ".", -1))
		}

	}

}

// These functions are basically from https://github.com/EvilSuperstars/go-cidrman/blob/master/ipv4.go , but there's no need to import that whole module for two basic functions, and they're licensed very permissively so copying isn't an issue.
func ipv4ToInt(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}
func intToIpv4(i uint32) net.IP {
	ip := make([]byte, net.IPv4len)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}
func lookupName(name string) net.IP {
	if name == domain_name {
		glog.Infof("Got a request for the base domain, returning HTTP server IP: %s", http_ip)
		return net.ParseIP(http_ip)
	}
	pieces := strings.Split(name, ".dconf."+domain_name)
	first := strings.Join(pieces[:len(pieces)-1], "")
	params := strings.Split(first, "-")
	arguments := make(map[string]string)

	for i, str := range params {
		if i%2 != 0 {
			arguments[params[i-1]] = str
		}
	}
	fmt.Println("args:", arguments)
	ip := makeAnswer(name, arguments)
	http.BroadcastToListeners(fmt.Sprintf("DNS request: \"%s\" response: \"%s\"", name, ip))
	return ip
}

func serve(net, port string) {
	server := &dns.Server{Addr: ":8053", Net: net, TsigSecret: nil}
	if err := server.ListenAndServe(); err != nil {
		glog.Errorf("Failed to setup the "+net+" server: %s\n", err.Error())
	}
}

func handleRebind(w dns.ResponseWriter, r *dns.Msg) {
	var (
		v4 bool
		rr dns.RR
		a  net.IP
	)
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = *compress

	// Assign IP for our reply to a
	a = lookupName(r.Question[0].Name)
	v4 = true
	// Build the actual response object we're going to return.
	if v4 {
		rr = &dns.A{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
			A:   a.To4(),
		}
	} else {
		rr = &dns.AAAA{
			Hdr:  dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0},
			AAAA: a,
		}
	}

	switch r.Question[0].Qtype {
	default:
		fallthrough
	case dns.TypeAAAA, dns.TypeA:
		m.Answer = append(m.Answer, rr)
	}
	w.WriteMsg(m)
}

func StartDNS(domain, port, server_ip, _capture_ip string) {
	//fmt.Printf("Started\n")
	domain_name = domain
	http_ip = server_ip
	capture_ip = _capture_ip
	dns.HandleFunc(".", handleRebind)
	go serve("tcp", port)
	go serve("udp", port)

}
