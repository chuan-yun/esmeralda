package trace

import (
	"fmt"

	"chuanyun.io/esmeralda/util"
)

type ListResult struct {
	TraceID         string                  `json:"traceId"`
	SpanCount       int                     `json:"spanCount"`
	Duration        int64                   `json:"duration"`
	ServiceNameList map[string]*ServiceName `json:"serviceNameList"`
	ComponentList   map[string]*Component   `json:"componentList"`
	spanIds         map[string]bool
	Timestamp       int64  `json:"timestamp"`
	Root            bool   `json:"root"`
	TraceStatus     string `json:"traceStatus"`
}

type ServiceName struct {
	Name        string          `json:"name"`
	Duration    int64           `json:"duration"`    //最大耗时
	AllDuration int64           `json:"allDuration"` //总耗时
	Method      string          `json:"method"`
	URI         string          `json:"uri"`
	Ratio       string          `json:"ratio"`
	Count       int             `json:"count"`
	Error       map[string]bool `json:"error"` //该服务错误类别统计
	Status      string          `json:"status"`
	HostIP      string          `json:"hostIp"`
}

type Component struct {
	Count  int             `json:"count"`
	Error  map[string]bool `json:"error"`
	Status string          `json:"status"`
}

func InitResult(traceID, spanID string) *ListResult {
	return &ListResult{
		TraceID:         traceID,
		SpanCount:       0,
		Duration:        0,
		Root:            false,
		TraceStatus:     "normal",
		Timestamp:       0,
		spanIds:         map[string]bool{spanID: true},
		ServiceNameList: map[string]*ServiceName{},
		ComponentList:   map[string]*Component{},
	}
}

func (ListResult *ListResult) SetDuration(duration int64) {
	ListResult.Duration = duration
}

func (ListResult *ListResult) SetTimestamp(Timestamp int64) {
	ListResult.Timestamp = Timestamp
}

func (ListResult *ListResult) SetRoot(Root bool) {
	ListResult.Root = Root
}

func (ListResult *ListResult) SpanPlus(spanID string) {
	if _, ok := ListResult.spanIds[spanID]; !ok {
		ListResult.spanIds[spanID] = true
	}
	ListResult.SpanCount = len(ListResult.spanIds)
}

func (ListResult *ListResult) SetServiceName(serverName, uri string) {
	ListResult.ServiceNameList[serverName] = &ServiceName{
		Name:        serverName,
		Count:       0,
		Duration:    0,
		AllDuration: 0,
		Method:      "GET",
		URI:         uri,
		Error:       map[string]bool{},
		Status:      "normal",
		HostIP:      "",
	}
}

func (ListResult *ListResult) ServiceNameUri(serverName string, binaryAnnotations []BinaryAnnotation) {
	errorType := ""
	errorVal := ""
	for _, val := range binaryAnnotations {
		if val.Key == "http.url" {
			ListResult.ServiceNameList[serverName].URI = val.Value
		}
		if val.Key == "http.status_code" && val.Value != "200" && val.Value != "201" {
			errorType = "http"
			errorVal = val.Value
		}
		if val.Key == "error" {
			errorVal = val.Value
		}
		if val.Key == "sa" {
			ListResult.ServiceNameList[serverName].HostIP = val.Endpoint.Ipv4
		}
		if val.Key == "http.url" {
			ListResult.ServiceNameList[serverName].HostIP = val.Value
		}
	}
	if errorType == "http" && errorVal != "" {
		ListResult.TraceStatus = "danger"
		ListResult.ServiceNameList[serverName].Status = "danger"
		ListResult.ServiceNameList[serverName].Error[errorVal] = true
	}
}

func (ListResult *ListResult) ServiceNamePlus(serverName string) {
	ListResult.ServiceNameList[serverName].Count++
}

func (ListResult *ListResult) ServiceNameDuration(serverName string, duration int64) {
	if ListResult.Root == false {
		ListResult.Duration = util.MaxInt64(ListResult.Duration, duration)
	}
	ListResult.ServiceNameList[serverName].AllDuration = ListResult.ServiceNameList[serverName].AllDuration + duration
	ListResult.ServiceNameList[serverName].Duration = util.MaxInt64(duration, ListResult.ServiceNameList[serverName].Duration)
}

func (ListResult *ListResult) TraceRatio() {
	for _, sval := range ListResult.ServiceNameList {
		var Ratio int64
		Ratio = 0
		if ListResult.Duration != 0 {
			Ratio = sval.Duration / ListResult.Duration
		}
		sval.Ratio = fmt.Sprintf("%.2f", float64(Ratio))
	}
}

func (ListResult *ListResult) initComponent(cType string) {
	if _, ok := ListResult.ComponentList[cType]; !ok {
		ListResult.ComponentList[cType] = &Component{
			Count:  0,
			Error:  map[string]bool{},
			Status: "normal",
		}
	}
}

func (ListResult *ListResult) ComponentPlus(cType string) {
	ListResult.initComponent(cType)
	ListResult.ComponentList[cType].Count++
}

func (ListResult *ListResult) ComponentError(cType string, error string) {
	if cType == "" {
		return
	}
	ListResult.initComponent(cType)
	ListResult.ComponentList[cType].Error[error] = true
	ListResult.ComponentList[cType].Status = "danger"
}

func formatComponentType(stype string) string {
	trans := map[string]string{
		"redis":    "Redis",
		"mysql":    "MySQL",
		"memcache": "Memcached",
	}
	if trans[stype] == "" {
		return stype
	}
	return trans[stype]
}

func (ListResult *ListResult) setComponentInfo(binaryAnnotations []BinaryAnnotation) {
	componentType := ""
	for _, val := range binaryAnnotations {
		if val.Key == "db.type" && (val.Value == "redis" || val.Value == "mysql" || val.Value == "memcache") {
			componentType = formatComponentType(val.Value)
			ListResult.ComponentPlus(componentType)
		}
		if val.Key == "error" {
			ListResult.TraceStatus = "danger"
			ListResult.ComponentError(componentType, val.Value)
		}
	}
}
