package trace

import (
	"fmt"
	"sort"
	"strings"

	"chuanyun.io/esmeralda/util"
)

type WFList []*WaterfallList

func (WFList WFList) getChild(id string) (*WaterfallList, bool) {
	for _, val := range WFList {
		if val.ID == id {
			return val, true
		}
	}
	return nil, false
}

func (WFList WFList) addChild(wf *WaterfallList) WFList {
	for i, val := range WFList {
		if val.ID == wf.ID {
			WFList[i] = wf
			return WFList
		}
	}
	WFList = append(WFList, wf)
	return WFList
}

func (WFList WFList) Len() int {
	return len(WFList)
}

func (WFList WFList) Less(i, j int) bool {
	return WFList[i].Timestamp < WFList[j].Timestamp
}

func (WFList WFList) Swap(i, j int) {
	WFList[i], WFList[j] = WFList[j], WFList[i]
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
	ID             string         `json:"id"`
	ParentID       string         `json:"parentId"`
	Name           string         `json:"name"`
	Nodes          WFList         `json:"nodes"`
	ServiceName    string         `json:"serviceName"`
	Timestamp      int64          `json:"timestamp"`
	TopoURI        string         `json:"topoUri"`
}

type AllAnnotations struct {
	Annotations       []AnnotationsMap       `json:"annotations"`
	Base              ABaseMap               `json:"base"`
	BinaryAnnotations []BinaryAnnotationsMap `json:"binaryAnnotations"`
}

type ABaseMap struct {
	ParentID string `json:"parentId"`
	SpanID   string `json:"spanId"`
	TraceID  string `json:"traceId"`
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

func (s AllAnnotations) Len() int {
	return len(s.Annotations)
}

func (s AllAnnotations) Less(i, j int) bool {
	return s.Annotations[i].sortAnnotation < s.Annotations[j].sortAnnotation
}

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

func (WaterResult *WaterResult) SpanStat(span Span) {
	if span.ParentID == "" {
		WaterResult.Stat.Duration = span.Duration
		WaterResult.Stat.StartTimestamp = span.Timestamp
	}
	if len(span.Annotations) == 0 || len(span.BinaryAnnotations) == 0 {
		fmt.Print("Annotations or BinaryAnnotations is empty")
		return
	}
	if len(WaterResult.Stat.spanIds) == 0 {
		WaterResult.Stat.spanIds = map[string]bool{span.ID: true}
	} else {
		if _, ok := WaterResult.Stat.spanIds[span.ID]; !ok {
			WaterResult.Stat.spanIds[span.ID] = true
		}
	}
	WaterResult.Stat.SpanCount = len(WaterResult.Stat.spanIds)
	serverName := span.Annotations[0].Endpoint.ServiceName
	if checkServerName(serverName) {
		if _, ok := WaterResult.Stat.ServiceList[serverName]; !ok {
			if WaterResult.Stat.ServiceList == nil {
				WaterResult.Stat.ServiceList = map[string]int{serverName: 0}
			}
		}
		WaterResult.Stat.ServiceList[serverName]++
	}
}

func getChild(pid string, spans []Span) (ret []Span, left []Span) {
	for _, span := range spans {
		if span.ParentID == pid {
			ret = append(ret, span)
		} else {
			left = append(left, span)
		}
	}
	return
}

func (WaterResult *WaterResult) SpanList(spans []Span) {
	tempWaterfallList := map[string]*WaterfallList{}

	for _, span := range spans {
		wf := &WaterfallList{
			ID:             span.ID,
			Duration:       span.Duration,
			Name:           span.Name,
			Flag:           "default",
			Timestamp:      span.Timestamp,
			ParentID:       span.ParentID,
			ServiceName:    span.getSpanServerName(),
			TopoURI:        span.getSpanTopoURI(),
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

		spanAnnotation := span.formatAnnotations(WaterResult.Stat.StartTimestamp)
		spanBinaryAnnotation := span.formatBinaryAnnotations(WaterResult.Stat.StartTimestamp)

		wf.AllAnnotations = AllAnnotations{
			Annotations:       spanAnnotation,
			BinaryAnnotations: spanBinaryAnnotation,
			Base: ABaseMap{
				TraceID:  span.TraceID,
				SpanID:   span.ID,
				ParentID: span.ParentID,
			},
		}

		if tempWaterfallList[wf.ID] != nil {
			tempWaterfallList[wf.ID].AllAnnotations.Annotations = mergeAnnotation(tempWaterfallList[wf.ID].AllAnnotations.Annotations, spanAnnotation)
			tempWaterfallList[wf.ID].AllAnnotations.BinaryAnnotations = append(tempWaterfallList[wf.ID].AllAnnotations.BinaryAnnotations, spanBinaryAnnotation...)
		} else {
			tempWaterfallList[wf.ID] = wf
		}
	}

	for _, span := range spans {
		pWf, ok1 := tempWaterfallList[span.ParentID]
		curWf, ok2 := tempWaterfallList[span.ID]
		if ok1 && ok2 {
			if pWf.ParentID == "" {
				pWf.ParentID = "0"
				WaterResult.List = WaterResult.List.addChild(pWf)
			}

			if _, ok := WaterResult.List.getChild(pWf.ParentID); ok {

				if span.isClient() && curWf.Name == "php_curl" {
					curWf.Name = span.Name
				}

				if span.isServer() {
					if span.Duration != 0 {
						curWf.Duration = span.Duration
					}
					if span.Timestamp != 0 {
						curWf.Timestamp = span.Timestamp
					}
				}

				curWf.AllAnnotations.Annotations = span.formatAnnotations(WaterResult.Stat.StartTimestamp)
				curWf.AllAnnotations.BinaryAnnotations = span.formatBinaryAnnotations(WaterResult.Stat.StartTimestamp)
			}

			curWf.SetNameFlag()
			sort.Sort(curWf.AllAnnotations)
			pWf.Nodes = pWf.Nodes.addChild(curWf)
		}
	}
	//对Timestamp 排序
	SortList(WaterResult.List)
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

func (span *Span) getSpanTopoURI() string {
	topoURI := ""
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
		return topoURI
	}

	if tmp["type"] == "http" {
		return tmp["address"]
	}

	if tmp["address"] == "" {
		return topoURI
	}

	topoURI = tmp["type"] + "://" + tmp["address"]
	if tmp["type"] == "mysql" {
		topoURI += "/" + tmp["instance"]
	}

	return topoURI
}

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
