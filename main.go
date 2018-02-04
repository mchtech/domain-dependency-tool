package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type dnsrecord struct {
	name     string
	ip       []string
	ns       []string
	cname    string
	verifyed bool
	//soa      string
}

var mp = make(map[string]*dnsrecord)
var cla = uint16(dns.ClassINET)

//得到域名的第nsindex的权威DNS的IP 和 权威DNS的个数
// return IPs, NS count
func getnsip(nm string, nsindex int) ([]string, int) {
	var ret []string
	var nslen int
	// 获取 name 的 NS记录
	record := mp[nm]

	if record != nil && len(record.ns) > 0 {
		nslen = len(record.ns)
		for i := nsindex; i < nslen; i++ {
			ns := record.ns[i]
			// 获取 ns 的 IP地址
			record2 := mp[ns]
			// IP 存在
			if record2 != nil && len(record2.ip) > 0 {
				ret = record2.ip
				break
			} else {
				ret = analyze(ns, false)
			}
		}

	} else { // 如果 nm (ccc.bbb.aaa) 的 NS记录 不存在
		// uplevelnm (bbb.aaa)
		uplevelnm := getuplevelname(nm)
		var nsip []string
		maxloop := 1
		for i := 0; i < maxloop; i++ {
			fmt.Println("--Loop", i, "for getting NS of", uplevelnm)
			// nsip: 得到 bbb.aaa 的 NS
			nsip, maxloop = getnsip(uplevelnm, i)
			if len(nsip) > 0 {
				// 向 nsip (bbb.aaa 的 NS) 查询 nm (ccc.bbb.aaa) 的 NS (nsres)
				nsres := querydns(nsip, nm, dns.TypeNS, cla)
				if nsres.ans != nil {
					if len(nsres.ans.Ns) == 0 {
						//使用上级name的NS
						uplevelrecord := mp[uplevelnm]
						if uplevelrecord != nil {
							mergenstodnsrecord(nm, uplevelrecord.ns)
							//fmt.Println("get result", mp[nm])
							ret = nsip
							//是指针不用重取
							nslen = len(uplevelrecord.ns)
						} else {
							//bug
							fmt.Println("bug 1!")
						}
					} else {
						fmt.Println("---Using NS records in answer of", nm)
						// nm(ccc.bbb.aaa) 的 NS (nsrecords)
						nsrecords := searchnsfromns(nm, nsres.ans.Ns)
						nslen = len(nsrecords)
						if nslen > 0 {
							var mergednamefromns []string
							for _, nsrecord := range nsrecords {
								nsrecord = strings.ToLower(nsrecord)
								nsi := searchipfromextra(nsrecord, nsres.ans.Extra)
								if len(nsi) > 0 {
									mergeaaaaatodnsrecord(nsrecord, nsi)
								} else {
									//fmt.Println("bug3")
								}
								mergednamefromns = append(mergednamefromns, nsrecord)
								appendnstodnsrecord(nm, nsrecord)
								fmt.Println("---Loop", i, "Get results(from Authority RRs)", mp[nm])
								ip := analyze(nsrecord, false)
								ret = dupmerge(ret, ip)
							}
							//验证每一个NS
							if cfg.VerifyNS {
								for _, mname := range mergednamefromns {
									mname = strings.ToLower(mname)
									if stat, ok := mp[mname]; !ok || !stat.verifyed {
										fmt.Println("----Loop", i, "Verify NS of", mname)
										//重新验证
										analyze(mname, true)
										if stat, ok = mp[mname]; ok {
											stat.verifyed = true
										}
									}
								}
							}
						} else { // NS record 查找失败
							fmt.Println("---Loop", i, "NS Record 查找失败")
							//return ret
						}
					}
				}
				break
			} else { // 本级 NS 查找失败
				//fmt.Println("here 4")
				//return ret
				continue
			}
		}
	}
	return ret, nslen
}

//reserved : must false
func analyze(nm string, reserved bool) []string {

	nm = strings.ToLower(nm)
	var res []string

	fmt.Println("-Analyze", nm)
	record := mp[nm]

	var nsip []string
	// 得到域名的权威DNS IP
	// nm = ns1.aaa.com.
	//upname = aaa.com.
	upname := getuplevelname(nm)
	uprecord := mp[upname]
	//uprecord.ns = ns1.aaa.com.
	//nsip = record.ip
	if record != nil && (!reserved || (reserved && uprecord != nil && len(uprecord.ns) > 0 && uprecord.ns[0] == nm)) {
		//if record != nil && (!reserved || (reserved && record.soa == nm)) {
		if len(record.ip) > 0 {
			res = record.ip
			fmt.Println("-Analyze Results (from cache)", res)
			return res
		}
	}

	maxloop := 1
	for i := 0; i < maxloop; i++ {
		fmt.Println("--Begin Analyze Loop", i, "Name", nm)
		nsip, maxloop = getnsip(upname, i)

		if len(nsip) > 0 {
			ok := true
			//使用权威DNS解析
			qtypes := []uint16{dns.TypeA, dns.TypeAAAA}
			//qtypes := []uint16{dns.TypeA}

			for _, qtype := range qtypes {
				nmip := querydns(nsip, nm, qtype, cla)
				if nmip.ans != nil && len(nmip.ans.Answer) > 0 {
					if nmip.ans.Answer[0].Header().Rrtype == dns.TypeA || nmip.ans.Answer[0].Header().Rrtype == dns.TypeAAAA {
						ips := searchaaaaafromansver(nmip.ans.Answer)
						mergeaaaaatodnsrecord(nm, ips)
						res = dupmerge(res, ips)
						//fmt.Println("analyze result", mp[nm])
					} else if nmip.ans.Answer[0].Header().Rrtype == dns.TypeCNAME {
						cname := nmip.ans.Answer[0].(*dns.CNAME).Target
						res = dupmerge(res, analyze(cname, false))
						mergeaaaaatodnsrecord(nm, res)
						mp[nm].cname = cname
					} else {
						// mal package
						fmt.Println("mal package")
					}
				} else {
					if nmip.ans == nil {
						ok = false
						break
					} else if nmip.ans != nil && !nmip.ans.Authoritative {
						//非权威应答
						fmt.Println("!! non-authoritive answer !!")
						nsrecords := searchnsfromns(nm, nmip.ans.Ns)
						nsrecordslen := len(nsrecords)
						if nsrecordslen > 0 {
							mergenstodnsrecord(upname, nsrecords)
							for _, nsrecord := range nsrecords {
								nsrecord = strings.ToLower(nsrecord)
								//同源
								if strings.Index(nsrecord, upname) != -1 {
									nsi := searchipfromextra(nsrecord, nmip.ans.Extra)
									nsiplen := len(nsi)
									if nsiplen > 0 {
										mergeaaaaatodnsrecord(nsrecord, nsi)
									}
								}
								if strings.Count(nsrecord, ".")-strings.Count(upname, ".") > 1 {
									getnsip(getuplevelname(nsrecord), 0)
								}
							}
							maxloop += nsrecordslen
							i = maxloop - nsrecordslen - 1
						}
						ok = false
						break
					} else {
						// OK
					}
				}
			}
			if !ok {
				continue
			}
		} else {
			fmt.Println("bug2!", nsip)
		}
		break
	}
	fmt.Println("-Analyze Results", res)
	return res
}

type Config struct {
	Name              string
	EDNSClient        string
	eDNSClientBin     net.IP
	EDNSClientMask    uint8
	EDNSClientFamily  uint16
	timeout           int
	Timeout           time.Duration
	TimeoutRetry      int
	VerifyNS          bool
	ResolveRootServer bool
}

var cfg Config

func main() {
	start := time.Now()

	flag.BoolVar(&cfg.VerifyNS, "nv", false, "No not verify NS records in Authority Records (will take a short time and get less relationships)")
	flag.BoolVar(&cfg.ResolveRootServer, "root", false, "Resolve root-servers Records")
	flag.StringVar(&cfg.EDNSClient, "eip", "", "IPv4 or IPv6 address for EDNS-Client-Subnet")
	flag.IntVar(&cfg.timeout, "t", 2, "timeout of resolving domain (in second)")
	flag.IntVar(&cfg.TimeoutRetry, "c", 4, "retry count (timeout)")
	flag.Parse()

	cfg.VerifyNS = !cfg.VerifyNS
	cfg.Timeout = time.Duration(cfg.timeout) * time.Second
	cfg.Name = flag.Arg(0)

	//cfg.Name = "www.amazon.com."
	// cfg.VerifyNS = true

	if cfg.Name == "" {
		flag.Usage()
		fmt.Println("  DomainName string\n\na domain name is required.\n")
		return
	}
	// fqdn format
	if !strings.HasSuffix(cfg.Name, ".") {
		cfg.Name += "."
	}
	//edns
	if cfg.EDNSClient != "" {
		if strings.Index(cfg.EDNSClient, ":") != -1 {
			cfg.EDNSClientFamily = 2
			cfg.EDNSClientMask = 128
			cfg.eDNSClientBin = net.ParseIP(cfg.EDNSClient)
		} else {
			cfg.EDNSClientFamily = 1
			cfg.EDNSClientMask = 32
			cfg.eDNSClientBin = net.ParseIP(cfg.EDNSClient).To4()
		}
	}
	// setup root-servers

	rootservers := []string{}
	for i := 'a'; i <= 'm'; i++ {
		rootserver := string(i) + ".root-servers.net."
		rootservers = append(rootservers, rootserver)
		mp[rootserver] = &dnsrecord{
			name: rootserver,
		}
	}
	mp["a.root-servers.net."] = &dnsrecord{
		name: "a.root-servers.net.",
		ip:   []string{"198.41.0.4", "2001:503:ba3e::2:30"},
	}
	for i := 'a'; i <= 'm'; i++ {
		rootserver := string(i) + ".root-servers.net."
		mp[rootserver].ns = rootservers
	}

	mp["."] = &dnsrecord{
		name: ".",
		ns:   rootservers,
	}

	if cfg.ResolveRootServer {
		for i := 'a'; i <= 'm'; i++ {
			analyze(string(i)+".root-servers.net.", true)
		}
	}
	//process
	result := analyze(cfg.Name, false)
	getnsip(cfg.Name, 0)

	//dump
	fmt.Println("----------------------------------------------------------------------------------------------------")
	for i, data := range mp {
		//fmt.Println(i, "Fqdn", data.name, "SOA", data.soa, "NS", data.ns, "CNAME", data.cname, "IP", data.ip)
		fmt.Println(i, "Fqdn", data.name, "NS", data.ns, "CNAME", data.cname, "IP", data.ip)
	}
	fmt.Println("----------------------------------------------------------------------------------------------------")
	fmt.Println(result)

	//gen map html
	fname := cfg.Name
	if !cfg.VerifyNS {
		fname += "-noverifyns"
	}
	if cfg.EDNSClient != "" {
		ip := strings.Replace(cfg.EDNSClient, ":", "-", -1)
		fname += "-edns-" + ip
	}
	fname += ".html"

	genjs(cfg.Name, mp, fname)
	fmt.Println(fname + " generated")
	OpenURL(fname)
	fmt.Println("Spent", time.Now().Sub(start))
}

var commands = map[string][]string{
	"windows": []string{"cmd", " /c start "},
	"darwin":  []string{"open"},
	"linux":   []string{"xdg-open"},
}

// Open calls the OS default program for uri
func OpenURL(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}

	cmd := exec.Command(run[0], run[1]+uri)
	return cmd.Start()
}
