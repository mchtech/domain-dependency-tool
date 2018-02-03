package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func querydns(servers []string, name string, qtype uint16, qclass uint16) querydnsresult {
	var ret querydnsresult
	for _, server := range servers {
		if strings.Index(server, ":") != -1 && len(server) > 0 && server[0] != '[' {
			server = "[" + server + "]"
		}
		ok := false
		proto := "udp"
		for i := 0; i < cfg.TimeoutRetry+1; i++ {
			//time.Sleep(100 * time.Millisecond)
			m1 := new(dns.Msg)
			m1.Compress = true
			m1.Id = dns.Id()
			m1.RecursionDesired = false
			m1.Question = []dns.Question{
				dns.Question{
					Name:   dns.Fqdn(name),
					Qtype:  qtype,
					Qclass: qclass}}

			if cfg.EDNSClient != "" {
				subnet := dns.EDNS0_SUBNET{
					Code:          dns.EDNS0SUBNET,
					Family:        cfg.EDNSClientFamily,
					SourceNetmask: cfg.EDNSClientMask,
					SourceScope:   0,
					Address:       cfg.eDNSClientBin,
				}
				opt := new(dns.OPT)
				opt.Hdr.Name = "."
				opt.Hdr.Rrtype = dns.TypeOPT
				opt.Option = append(opt.Option, &subnet)
				m1.Extra = append(m1.Extra, opt)
			}

			var ans1 *dns.Msg
			var err1 error

			var conn *dns.Conn

			conn, err1 = dns.DialTimeout(proto, server+":53", cfg.Timeout)

			if err1 == nil && conn != nil {
				co := new(dns.Conn)
				co.Conn = conn
				co.SetDeadline(time.Now().Add(cfg.Timeout))
				co.WriteMsg(m1)
				ans1, err1 = co.ReadMsg()
				defer co.Close()
			}
			// return result
			if err1 == nil {
				ret = querydnsresult{
					ans: ans1,
					err: err1,
				}
				ok = true
				break
			} else {
				ret.err = err1
				if err1.Error() == "dns: failed to unpack truncated message" {
					proto = "tcp"
				} else {
					fmt.Println("error:", err1)
				}
			}
		}
		if ok {
			break
		}
	}
	if ret.ans == nil {
		fmt.Println("[FATAL ERROR]", ret.err, name, dns.TypeToString[qtype], servers)
	}
	return ret
}

type querydnsresult struct {
	ans *dns.Msg
	err error
}
