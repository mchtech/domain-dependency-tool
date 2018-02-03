package main

import (
	"strings"

	"github.com/miekg/dns"
)

func inarray(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func searchipfromextra(name string, exrr []dns.RR) []string {
	name = strings.ToLower(name)
	var res []string
	for _, v := range exrr {
		if strings.ToLower(v.Header().Name) == name {
			if v.Header().Rrtype == dns.TypeA {
				res = append(res, v.(*dns.A).A.String())
			} else if v.Header().Rrtype == dns.TypeAAAA {
				res = append(res, v.(*dns.AAAA).AAAA.String())
			}
		}
	}
	return res
}

func dupmerge(src []string, dst []string) []string {
	for _, d := range dst {
		dup := false
		for _, s := range src {
			if s == d {
				dup = true
				break
			}
		}
		if !dup {
			src = append(src, d)
		}
	}
	return src
}

func searchnsfromns(name string, nss []dns.RR) []string {
	name = strings.ToLower(name)
	var res []string
	for _, v := range nss {
		if v.Header().Rrtype == dns.TypeNS {
			res = append(res, strings.ToLower(v.(*dns.NS).Ns))
		} else if v.Header().Rrtype == dns.TypeSOA {
			soaname := v.(*dns.SOA).Hdr.Name
			soaname = strings.ToLower(soaname)
			if soaname != name {
				//getnsip(soaname, 0)
				//analyze(soaname)
				soarecord := mp[soaname]
				if soarecord != nil {
					res = dupmerge(res, soarecord.ns)
				}
				// if sv, ok := mp[name]; ok {
				// 	sv.soa = soaname
				// } else {
				// 	mp[name] = &dnsrecord{
				// 		name: name,
				// 		ns:   []string{},
				// 		soa:  soaname,
				// 		ip:   []string{},
				// 	}
				// }
			}
			res = dupmerge(res, []string{strings.ToLower(v.(*dns.SOA).Ns)})
		}
	}
	return res
}

func searchaaaaafromansver(ans []dns.RR) []string {
	var res []string
	for _, v := range ans {
		if v.Header().Rrtype == dns.TypeA {
			res = append(res, v.(*dns.A).A.String())
		} else if v.Header().Rrtype == dns.TypeAAAA {
			res = append(res, v.(*dns.AAAA).AAAA.String())
		}
	}
	return res
}

func appendnstodnsrecord(name string, nsrecord string) {
	name = strings.ToLower(name)
	nsrecord = strings.ToLower(nsrecord)
	record := mp[name]
	if record == nil {
		mp[name] = &dnsrecord{
			name: name,
			ns:   []string{nsrecord},
			ip:   []string{},
		}
	} else {
		for _, ns := range record.ns {
			if ns == nsrecord {
				return
			}
		}
		record.ns = append(record.ns, nsrecord)
	}
}

func mergeaaaaatodnsrecord(name string, aaaaa []string) {
	name = strings.ToLower(name)
	record := mp[name]
	if record == nil {
		mp[name] = &dnsrecord{
			name: name,
			ip:   aaaaa,
			ns:   []string{},
		}
	} else {
		for _, aaa := range aaaaa {
			exist := false
			for _, ip := range record.ip {
				if ip == aaa {
					exist = true
				}
			}
			if !exist {
				record.ip = append(record.ip, aaa)
			}
		}
	}
}

func mergenstodnsrecord(name string, aaaaa []string) {
	name = strings.ToLower(name)
	for i, v := range aaaaa {
		aaaaa[i] = strings.ToLower(v)
	}

	record := mp[name]
	if record == nil {
		mp[name] = &dnsrecord{
			name: name,
			ns:   aaaaa,
			ip:   []string{},
		}
	} else {
		for _, aaa := range aaaaa {

			exist := false
			for _, ns := range record.ns {
				if ns == aaa {
					exist = true
				}
			}
			if !exist {
				record.ns = append(record.ns, aaa)
			}
		}
	}
}

func getuplevelname(name string) string {
	idx := strings.Index(name, ".")
	ret := name[idx+1:]
	if ret == "" {
		ret = "."
	}
	return ret
}
