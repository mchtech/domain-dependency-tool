package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type EData struct {
	Name       string  `json:"name"`
	Draggable  bool    `json:"draggable"`
	Category   string  `json:"category"`
	Value      string  `json:"value"`
	SymbolSize float32 `json:"symbolSize"`
}
type ECategory struct {
	Name string `json:"name"`
}
type ELink struct {
	Source    string     `json:"source"`
	Target    string     `json:"target"`
	Value     string     `json:"value"`
	LineStyle ELineStyle `json:"lineStyle"`
}

type ELineStyle struct {
	Normal ELineStyleNormal `json:"normal"`
}
type ELineStyleNormal struct {
	Curveness float32 `json:"curveness"`
}

func genjs(name string, mp map[string]*dnsrecord, fname string) {
	var datas []EData
	var links []ELink
	var categories []ECategory
	var categoriesname []string

	catamap := make(map[string]int)

	for _, v := range mp {
		upname := "."
		if v.name != "." {
			// Upper Name
			upname = getuplevelname(v.name)
			if _, ok := catamap[upname]; !ok {
				catamap[upname] = 0
				categoriesname = append(categoriesname, upname)
				categories = append(categories, ECategory{
					Name: upname,
				})
			}
			//CNAME
			if v.cname != "" {
				catamap[v.cname] = 0
				categoriesname = append(categoriesname, v.cname)
				categories = append(categories, ECategory{
					Name: v.cname,
				})
			}
			//CNAME Link
			if v.cname != "" {
				links = append(links, ELink{
					Source: v.name,
					Target: v.cname,
					Value:  "CNAME",
					LineStyle: ELineStyle{
						Normal: ELineStyleNormal{
							Curveness: 0.2,
						},
					},
				})
			}

			// //SOA Link
			// if v.soa != "" {
			// 	links = append(links, ELink{
			// 		Source: v.name,
			// 		Target: v.soa,
			// 		Value:  "SOA",
			// 		LineStyle: ELineStyle{
			// 			Normal: ELineStyleNormal{
			// 				Curveness: 0.2,
			// 			},
			// 		},
			// 	})
			// }
		}

		//Data
		dotcount := float32(strings.Count(v.name, "."))
		cnss := float32(48.00)
		cnss = cnss / dotcount

		data := EData{
			Name:       v.name,
			Draggable:  true,
			Category:   upname,
			Value:      strings.Join(v.ip, "\n"),
			SymbolSize: cnss,
		}
		datas = append(datas, data)

		//NS Link
		for _, ns := range v.ns {
			links = append(links, ELink{
				Source: v.name,
				Target: ns,
				Value:  "NS",
				LineStyle: ELineStyle{
					Normal: ELineStyleNormal{
						Curveness: 0.2,
					},
				},
			})
		}
	}

	for _, v := range mp {
		ln := v.name
		for {
			if ln == "." {
				break
			}
			//Upper Name Link
			un := getuplevelname(ln)
			links = append(links, ELink{
				Source: ln,
				Target: un,
				Value:  "Parent",
				LineStyle: ELineStyle{
					Normal: ELineStyleNormal{
						Curveness: 0.2,
					},
				},
			})
			ln = un
		}
	}

	// categories = make([]string, len(catamap))
	// for i, v := range catamap {
	// 	categories[v] = i
	// }

	cataarrjsonbin, _ := json.Marshal(categoriesname)
	catajsonbin, _ := json.Marshal(categories)
	datajsonbin, _ := json.Marshal(datas)
	linkjsonbin, _ := json.Marshal(links)

	cataarrjson := string(cataarrjsonbin)
	catajson := string(catajsonbin)
	datajson := string(datajsonbin)
	linkjson := string(linkjsonbin)

	jsonstr := fmt.Sprintf(tpl, cataarrjson, datajson, linkjson, catajson, name)
	ioutil.WriteFile(fname, []byte(jsonstr), 0666)
}
