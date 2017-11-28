package trace

import (
	"chuanyun.io/esmeralda/util"
	"fmt"
	"sort"
	"strings"
)

type WFList []*WaterfallList

func (self WFList) getChild(id string) (*WaterfallList, bool) {
	for _, val := range self {
		if val.Id == id {
			return val, true
		}
	}
	return nil, false
}

func (self WFList) addChild(wf *WaterfallList) WFList {
	for i, val := range self {
		if val.Id == wf.Id {
			self[i] = wf
			return self
		}
	}
	self = append(self, wf)
	return self
}

//为了Map的 Timestamp 排序 .
func (wf WFList) Len() int {
	return len(wf)
}

func (wf WFList) Less(i, j int) bool {
	return wf[i].Timestamp < wf[j].Timestamp
}

//Swap()
func (wf WFList) Swap(i, j int) {
	wf[i], wf[j] = wf[j], wf[i]
}

type WaterResult struct {
	List WFList        `json:"list"`
	Stat WaterfallStat `json:"stat"`
}

type WaterfallStat struct {
	ServiceList    map[string]int `json:"serviceList"`
	Duration       int64          `json:"duration"`
	SpanCount      int            `json:"spanCount"`
	StartTimestamp int64          `json:"startTimestamp"`
	spanIds        map[string]bool
}

type WaterfallList struct {
	AllAnnotations AllAnnotations `json:"allAnnotations"`
	Duration       int64          `json:"duration"`
	Flag           string         `json:"flag"`
	Id             string         `json:"id"`
	ParentId       string         `json:"parentId"`
	Name           string         `json:"name"`
	Nodes          WFList         `json:"nodes"`
	ServiceName    string         `json:"serviceName"`
	Timestamp      int64          `json:"timestamp"`
	TopoUri        string         `json:"topoUri"`
}

type AllAnnotations struct {
	Annotations       []AnnotationsMap       `json:"annotations"`
	Base              ABaseMap               `json:"base"`
	BinaryAnnotations []BinaryAnnotationsMap `json:"binaryAnnotations"`
}

type ABaseMap struct {
	ParentId string `json:"parentId"`
	SpanId   string `json:"spanId"`
	TraceId  string `json:"traceId"`
}

type AnnotationsMap struct {
	Address        string `json:"address"`
	Annotation     string `json:"annotation"`
	DateTime       string `json:"dateTime"`
	RelativeTime   int64  `json:"relativeTime"`
	sortAnnotation int
}

type BinaryAnnotationsMap struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//为了实现cs sr ss cr的顺序排列.
func (s AllAnnotations) Len() int {
	return len(s.Annotations)
}

func (s AllAnnotations) Less(i, j int) bool {
	return s.Annotations[i].sortAnnotation < s.Annotations[j].sortAnnotation
}

//Swap()
func (s AllAnnotations) Swap(i, j int) {
	s.Annotations[i], s.Annotations[j] = s.Annotations[j], s.Annotations[i]
}

func InitWaterResult() *WaterResult {
	return &WaterResult{
		List: WFList{},
		Stat: WaterfallStat{},
	}
}

func TransSort(stype string) int {
	trans := map[string]int{
		"ss": 30,
		"sr": 20,
		"cs": 10,
		"cr": 40,
	}
	if trans[stype] == 0 {
		return 100
	}
	return trans[stype]
}

func Trans(stype string) string {
	trans := map[string]string{
		"ss": "Server Send",
		"sr": "Server Receive",
		"lc": "Local Component",
		"la": "Local Address",
		"sa": "Server Address",
		"cs": "Client Send",
		"cr": "Client Receive",
	}
	if trans[stype] == "" {
		return stype
	}
	return trans[stype]
}

func TranServerName(stype string) string {
	stype = strings.ToUpper(stype)
	trans := map[string]string{
		"MYSQL":     "[MySQL]",
		"MYSQLI":    "[MySQL]",
		"REDIS":     "[Redis]",
		"MEMCACHE":  "[Memcached]",
		"MEMCACHED": "[Memcached]",
	}
	if trans[stype] == "" {
		return stype
	}
	return trans[stype]
}

func (self *WaterResult) SpanStat(span Span) {
	if span.ParentId == "" {
		self.Stat.Duration = span.Duration
		self.Stat.StartTimestamp = span.Timestamp
	}
	if len(span.Annotations) == 0 || len(span.BinaryAnnotations) == 0 {
		fmt.Print("Annotations or BinaryAnnotations is empty")
		return
	}
	if len(self.Stat.spanIds) == 0 {
		self.Stat.spanIds = map[string]bool{span.Id: true}
	} else {
		if _, ok := self.Stat.spanIds[span.Id]; !ok {
			self.Stat.spanIds[span.Id] = true
		}
	}
	self.Stat.SpanCount = len(self.Stat.spanIds)
	serverName := span.Annotations[0].Endpoint.ServiceName
	if checkServerName(serverName) {
		if _, ok := self.Stat.ServiceList[serverName]; !ok {
			if self.Stat.ServiceList == nil {
				self.Stat.ServiceList = map[string]int{serverName: 0}
			}
		}
		self.Stat.ServiceList[serverName]++
	}
}

func getChild(pid string, spans []Span) (ret []Span, left []Span) {
	for _, span := range spans {
		if span.ParentId == pid {
			ret = append(ret, span)
		} else {
			left = append(left, span)
		}
	}
	return
}

func (self *WaterResult) SpanList(spans []Span) {
	tempWaterfallList := map[string]*WaterfallList{} //临时存储的map

	for _, span := range spans {
		wf := &WaterfallList{
			Id:             span.Id,
			Duration:       span.Duration,
			Name:           span.Name,
			Flag:           "default",
			Timestamp:      span.Timestamp,
			ParentId:       span.ParentId,
			ServiceName:    span.getSpanServerName(),
			TopoUri:        span.getSpanTopoUri(),
			Nodes:          []*WaterfallList{},
			AllAnnotations: AllAnnotations{},
		}
		if len(span.BinaryAnnotations) > 0 {
			for _, binaryAnnotation := range span.BinaryAnnotations {
				if binaryAnnotation.Key == "sa" && binaryAnnotation.Endpoint.ServiceName != "" {
					wf.ServiceName = binaryAnnotation.Endpoint.ServiceName
				}
			}
		}

		spanAnnotation := span.formatAnnotations(self.Stat.StartTimestamp)
		spanBinaryAnnotation := span.formatBinaryAnnotations(self.Stat.StartTimestamp)

		wf.AllAnnotations = AllAnnotations{
			Annotations:       spanAnnotation,
			BinaryAnnotations: spanBinaryAnnotation,
			Base: ABaseMap{
				TraceId:  span.TraceId,
				SpanId:   span.Id,
				ParentId: span.ParentId,
			},
		}

		// 合并相同的span @TODO 合并2个span
		if tempWaterfallList[wf.Id] != nil {
			tempWaterfallList[wf.Id].AllAnnotations.Annotations = mergeAnnotation(tempWaterfallList[wf.Id].AllAnnotations.Annotations, spanAnnotation)
			tempWaterfallList[wf.Id].AllAnnotations.BinaryAnnotations = append(tempWaterfallList[wf.Id].AllAnnotations.BinaryAnnotations, spanBinaryAnnotation...)
		} else {
			tempWaterfallList[wf.Id] = wf
		}
	}

	for _, span := range spans {
		pWf, ok1 := tempWaterfallList[span.ParentId]
		curWf, ok2 := tempWaterfallList[span.Id]
		if ok1 && ok2 {
			if pWf.ParentId == "" {
				pWf.ParentId = "0"
				self.List = self.List.addChild(pWf)
			}
			//if _, ok := self.List[pWf.ParentId]; ok {
			if _, ok := self.List.getChild(pWf.ParentId); ok {

				//span合并 以 Server 端信息为主的部分
				if span.isClient() && curWf.Name == "php_curl" {
					curWf.Name = span.Name
				}
				// Span 合并，以 Client 端信息为主的部分
				if span.isServer() {
					//if span.Duration
					if span.Duration != 0 {
						curWf.Duration = span.Duration
					}
					if span.Timestamp != 0 {
						curWf.Timestamp = span.Timestamp
					}
				}

				// 不知道为什么要合并两边 @liupeng70
				//a := span.formatAnnotations(self.Stat.StartTimestamp)
				//curWf.AllAnnotations.Annotations = append(curWf.AllAnnotations.Annotations, a...)
				//b := span.formatBinaryAnnotations(self.Stat.StartTimestamp)
				//curWf.AllAnnotations.BinaryAnnotations = append(curWf.AllAnnotations.BinaryAnnotations, b...)

				curWf.AllAnnotations.Annotations = span.formatAnnotations(self.Stat.StartTimestamp)
				curWf.AllAnnotations.BinaryAnnotations = span.formatBinaryAnnotations(self.Stat.StartTimestamp)
			}

			//if span.isServer() {
			// Liupeng70 此处为何要对 span.isServer 单独处理?
			//curWf.ServiceName = span.getSpanServerName()
			curWf.SetNameFlag()
			//}

			sort.Sort(curWf.AllAnnotations)

			pWf.Nodes = pWf.Nodes.addChild(curWf)
		}
	}
	//对Timestamp 排序
	SortList(self.List)
}

func SortList(list WFList) {
	if len(list) > 0 {
		sort.Sort(list)
	}
	for _, val := range list {
		if len(val.Nodes) > 0 {
			SortList(val.Nodes)
		}
	}
}

func (wf *WaterfallList) SetNameFlag() {
	for _, bA := range wf.AllAnnotations.BinaryAnnotations {
		if bA.Key != "" && bA.Value != "" {
			if bA.Key == "component" || bA.Key == "db.type" {
				wf.ServiceName = TranServerName(bA.Value)
			}
			if bA.Key == "error" || (bA.Key == "http.status_code" && bA.Value != "200" && bA.Value != "201") {
				wf.Flag = "error"
			}
		}
	}
}

func (span *Span) isClient() bool {
	for _, val := range span.Annotations {
		if val.Value == "ss" || val.Value == "sr" {
			return true
		}
	}
	return false
}

func (span *Span) isServer() bool {
	for _, val := range span.Annotations {
		if val.Value == "cs" || val.Value == "cr" {
			return true
		}
	}
	return false
}

func (span *Span) getSpanServerName() string {
	serverName := "[unknown]"
	for _, val := range span.Annotations {
		serverName = val.Endpoint.ServiceName
	}
	return serverName
}

// 拿到 URI 的拓扑地址，http、redis、memcache、mysql
func (span *Span) getSpanTopoUri() string {
	topoUri := ""
	tmp := map[string](string){
		"type":      "",
		"instance":  "",
		"statement": "",
		"address":   "",
	}

	for _, val := range span.BinaryAnnotations {
		if val.Key == "http.url" {
			tmp["type"] = "http"
			tmp["address"] = val.Value
		}
		if val.Key == "db.type" {
			tmp["type"] = val.Value
		}
		if val.Key == "db.instance" {
			tmp["instance"] = val.Value
		}
		if val.Key == "sa" {
			tmp["address"] = val.Endpoint.Ipv4 + ":" + fmt.Sprintf("%d", val.Endpoint.Port)
		}
	}

	if tmp["type"] == "" {
		return topoUri
	}

	// 去除 gateway 的依赖
	if tmp["type"] == "http" {
		// ret := gateway.SearchUri(tmp["address"])
		// if ret.Id != 0 {
		// 	return ret.Uri
		// }
		return tmp["address"]
	} else {

		if tmp["address"] == "" {
			return topoUri
		}

		topoUri = tmp["type"] + "://" + tmp["address"]

		if tmp["type"] == "mysql" {
			topoUri += "/" + tmp["instance"]
		}
	}
	return topoUri
}

//格式化Annotations
func (span *Span) formatAnnotations(traceTimestamp int64) []AnnotationsMap {
	ret := []AnnotationsMap{}
	if len(span.Annotations) > 0 {
		for _, value := range span.Annotations {
			tmp := AnnotationsMap{
				DateTime:       util.FormatInt64TimeNsec(value.Timestamp),
				Annotation:     Trans(value.Value),
				RelativeTime:   value.Timestamp - traceTimestamp,
				Address:        "",
				sortAnnotation: TransSort(value.Value),
			}
			if value.Endpoint.Ipv4 != "" {
				tmp.Address = tmp.Address + value.Endpoint.Ipv4
			}
			if value.Endpoint.Port == 0 {
				tmp.Address = tmp.Address + ":" + fmt.Sprintf("%d", value.Endpoint.Port)
			}

			if value.Endpoint.ServiceName != "" {
				tmp.Address = tmp.Address + "(" + value.Endpoint.ServiceName + ")"
			}
			ret = append(ret, tmp)
		}
	}
	return ret
}

//格式化BinaryAnnotations
func (span *Span) formatBinaryAnnotations(traceTimestamp int64) []BinaryAnnotationsMap {
	ret := []BinaryAnnotationsMap{}
	if len(span.Annotations) > 0 {
		for _, value := range span.BinaryAnnotations {
			tmp := BinaryAnnotationsMap{
				Key:   Trans(value.Key),
				Value: value.Value,
			}
			if value.Key == "ca" || value.Key == "sa" || value.Key == "lc" {
				tmp.Value = ""
				if value.Endpoint.Ipv4 != "" {
					tmp.Value = tmp.Value + value.Endpoint.Ipv4
				}
				if value.Endpoint.Port == 0 {
					tmp.Value = tmp.Value + ":" + fmt.Sprintf("%d", value.Endpoint.Port)
				}
				if value.Endpoint.ServiceName != "" {
					tmp.Value = tmp.Value + "(" + value.Endpoint.ServiceName + ")"
				}
			}

			ret = append(ret, tmp)
		}
	}
	return ret
}

// 合并两个 Annotations
func mergeAnnotation(annotationOrg []AnnotationsMap, annotationNew []AnnotationsMap) []AnnotationsMap {
	if len(annotationNew) > 0 {
		tmpSpanKey := make(map[string]bool)
		for _, value := range annotationOrg {
			tmpSpanKey[value.Annotation] = true
		}
		for _, value := range annotationNew {
			if _, ok := tmpSpanKey[value.Annotation]; !ok {
				annotationOrg = append(annotationOrg, value)
			}
		}
	}
	return annotationOrg
}
