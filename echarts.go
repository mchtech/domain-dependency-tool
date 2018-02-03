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

const tpl = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
	<title>ECharts</title>
	<style>
		html, body, #main {
			width: 99%%;
			height: 99%%;
		}
	</style>
    <!-- 引入 echarts.js -->
    <script src="echarts.min.js"></script>
</head>
<body>
    <!-- 为ECharts准备一个具备大小（宽高）的Dom -->
    <div id="main"></div>
	<script type="text/javascript">
        // 基于准备好的dom，初始化echarts实例
        var myChart = echarts.init(document.getElementById('main'));
        // 指定图表的配置项和数据
        var option = {
			// backgroundColor: new echarts.graphic.RadialGradient(0.3, 0.3, 0.8, [{
			// 	offset: 0,
			// 	color: '#f7f8fa'
			// }, {
			// 	offset: 1,
			// 	color: '#cdd0d5'
			// }]),
			title: {
			},
			tooltip: {
				formatter: '{b0}:<br/><pre>{c0}</pre>'
			},
			legend: [{
				formatter: function(name) {
					return name;
				},
				tooltip: {
					show: true
				},
				selectedMode: 'false',
				top: 20,
				data: %s
			}],
			toolbox: {
				show: true,
				feature: {
					dataView: {
						show: true,
						readOnly: true
					},
					restore: {
						show: true
					},
					saveAsImage: {
						show: true
					}
				}
			},
			//animationDuration: 3000,
			//animationEasingUpdate: 'quinticInOut',
			series: [	
				{
					//symbol: 'rect',
					//symbolSize: [200,40],	
					edgeSymbol: ['circle', 'arrow'],				
					type: 'graph',
					layout: 'force',					
					focusNodeAdjacency: true,
					roam: true,
					lineStyle: {
						normal: {
							color: 'source',
							curveness: 0,
							type: "solid"
						}
					},
					force: {
						edgeLength: 125,
						repulsion: 200,
						gravity: 0.075
					},
					edgeLabel: {
						normal: {
							show: true,
							formatter: "{c}"
						}
					},
					data: %s,
					links: %s,
					categories: %s,

					label: {
						normal: {
							show: true,
							position: 'top',

						}
					},
					lineStyle: {
						normal: {
							color: 'source',
							curveness: 0,
							type: "solid"
						}
					}
				}
			]
		};
		var h = document.body.offsetHeight;
		var w = document.body.offsetWidth;
		option.series[0].data.forEach(function(fe){
			if(fe.name === "."){
				// fe.x = w/2;
				// fe.y = h/2;
				// fe.fixed = true;
				// fe.label = {
				// 	fontSize: 20
				// };
			}else if(fe.name === "%s"){
				fe.x = w/2;
				//fe.y = h/8;
				fe.y = h/2;
				fe.fixed = true;
				fe.label = {
					fontSize: 20
				};
			}
			// fe.tooltip = {
			// 	formatter: '{b0}: <br/> <pre>{c0}</pre>'
			// };
		});
        // 使用刚指定的配置项和数据显示图表。
        myChart.setOption(option);
    </script>
</body>
</html>`
