package trace

import (
	"chuanyun.io/esmeralda/util"
	"fmt"
)

//返回结果struct
type ListResult struct {
	TraceId         string                  `json:"traceId"`
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
	Uri         string          `json:"uri"`
	Ratio       string          `json:"ratio"`
	Count       int             `json:"count"`
	Error       map[string]bool `json:"error"` //该服务错误类别统计
	Status      string          `json:"status"`
	HostIp      string          `json:"hostIp"`
}

type Component struct {
	Count  int             `json:"count"`
	Error  map[string]bool `json:"error"`
	Status string          `json:"status"`
}

func InitResult(traceId, spanId string) *ListResult {
	return &ListResult{
		TraceId:         traceId,
		SpanCount:       0,
		Duration:        0,
		Root:            false,
		TraceStatus:     "normal",
		Timestamp:       0,
		spanIds:         map[string]bool{spanId: true},
		ServiceNameList: map[string]*ServiceName{},
		ComponentList:   map[string]*Component{},
	}
}

func (self *ListResult) SetDuration(duration int64) {
	self.Duration = duration
}

func (self *ListResult) SetTimestamp(Timestamp int64) {
	self.Timestamp = Timestamp
}

func (self *ListResult) SetRoot(Root bool) {
	self.Root = Root
}

func (self *ListResult) SpanPlus(spanId string) {
	if _, ok := self.spanIds[spanId]; !ok {
		self.spanIds[spanId] = true
	}
	self.SpanCount = len(self.spanIds)
}

func (self *ListResult) SetServiceName(serverName, uri string) {
	self.ServiceNameList[serverName] = &ServiceName{
		Name:        serverName,
		Count:       0,
		Duration:    0,
		AllDuration: 0,
		Method:      "GET",
		Uri:         uri,
		Error:       map[string]bool{},
		Status:      "normal",
		HostIp:      "",
	}
}

func (self *ListResult) ServiceNameUri(serverName string, binaryAnnotations []BinaryAnnotation) {
	errorType := ""
	errorVal := ""
	for _, val := range binaryAnnotations {
		if val.Key == "http.url" {
			self.ServiceNameList[serverName].Uri = val.Value
		}
		if val.Key == "http.status_code" && val.Value != "200" && val.Value != "201" {
			errorType = "http"
			errorVal = val.Value
		}
		if val.Key == "error" {
			errorVal = val.Value
		}
		if val.Key == "sa" {
			self.ServiceNameList[serverName].HostIp = val.Endpoint.Ipv4
		}
		if val.Key == "http.url" {
			self.ServiceNameList[serverName].HostIp = val.Value
		}
	}
	if errorType == "http" && errorVal != "" {
		self.TraceStatus = "danger"
		self.ServiceNameList[serverName].Status = "danger"
		self.ServiceNameList[serverName].Error[errorVal] = true
	}
}

func (self *ListResult) ServiceNamePlus(serverName string) {
	self.ServiceNameList[serverName].Count++
}

//计算最大耗时及耗时占比
func (self *ListResult) ServiceNameDuration(serverName string, duration int64) {
	if self.Root == false { //如果没有根节点则取最大的为总耗时
		self.Duration = util.MaxInt64(self.Duration, duration)
	}
	self.ServiceNameList[serverName].AllDuration = self.ServiceNameList[serverName].AllDuration + duration
	self.ServiceNameList[serverName].Duration = util.MaxInt64(duration, self.ServiceNameList[serverName].Duration)
}

func (self *ListResult) TraceRatio() {
	for _, sval := range self.ServiceNameList {
		var Ratio int64
		Ratio = 0
		if self.Duration != 0 {
			Ratio = sval.Duration / self.Duration
		}
		sval.Ratio = fmt.Sprintf("%.2f", float64(Ratio))
	}
}

func (self *ListResult) initComponent(cType string) {
	if _, ok := self.ComponentList[cType]; !ok {
		self.ComponentList[cType] = &Component{
			Count:  0,
			Error:  map[string]bool{},
			Status: "normal",
		}
	}
}

func (self *ListResult) ComponentPlus(cType string) {
	self.initComponent(cType)
	self.ComponentList[cType].Count++
}

func (self *ListResult) ComponentError(cType string, error string) {
	if cType == "" {
		return
	}
	self.initComponent(cType)
	self.ComponentList[cType].Error[error] = true
	self.ComponentList[cType].Status = "danger"
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

//计算trace的组件信息
func (self *ListResult) setComponentInfo(binaryAnnotations []BinaryAnnotation) {
	componentType := ""
	for _, val := range binaryAnnotations {
		if val.Key == "db.type" && (val.Value == "redis" || val.Value == "mysql" || val.Value == "memcache") {
			componentType = formatComponentType(val.Value)
			self.ComponentPlus(componentType)
		}
		if val.Key == "error" {
			self.TraceStatus = "danger"
			self.ComponentError(componentType, val.Value)
		}
	}
}
