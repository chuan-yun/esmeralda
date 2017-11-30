package trace

import (
	"encoding/json"
	"time"
)

type Endpoint struct {
	ServiceName string `json:"serviceName"`
	Ipv4        string `json:"ipv4"`
	Port        int    `json:"port"`
}

type Annotation struct {
	Value     string   `json:"value"`
	Timestamp int64    `json:"timestamp"`
	Endpoint  Endpoint `json:"endpoint"`
}

type BinaryAnnotation struct {
	Value    interface{} `json:"value"`
	Key      string      `json:"key"`
	Endpoint Endpoint    `json:"endpoint"`
}

type Span struct {
	ID                json.RawMessage    `json:"id"`
	ParentID          string             `json:"parentId,omitempty"`
	Timestamp         int64              `json:"timestamp"`
	Name              json.RawMessage    `json:"name"`
	Duration          json.RawMessage    `json:"duration"`
	Version           string             `json:"version,omitempty"`
	TraceID           json.RawMessage    `json:"traceId"`
	BinaryAnnotations []BinaryAnnotation `json:"binaryAnnotations,omitempty"`
	Annotations       []Annotation       `json:"annotations"`
	InsertTime        string             `json:"insertTime,omitempty"`
}

type Spans []Span

const (
	ServerSpan = iota
	ClientSpan = iota
)

func (span *Span) ToJson() (string, error) {
	str, err := json.Marshal(span)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func (span *Span) GetStoreMeta() (indexName string, typeName string, indexBaseName string) {
	microsecond := span.Timestamp
	date := ""

	if microsecond < 100000000000000 {
		date = time.Now().Local().Format("20060102")
	} else {
		date = time.Unix(0, microsecond*int64(time.Microsecond)).Local().Format("20060102")
	}

	indexName = "chuanyun" + "-" + date
	typeName = "span"

	return indexName, typeName, "chuanyun"
}

type Document struct {
	IndexName     string
	TypeName      string
	IndexBaseName string
	Payload       string
}

type DocumentQueue []Document

func (span *Span) AssembleDocument() (doc *Document, err error) {
	spanJSON, err := span.ToJson()
	if err != nil {
		return doc, err
	}

	doc.IndexName, doc.TypeName, doc.IndexBaseName = span.GetStoreMeta()
	doc.Payload = spanJSON

	return doc, nil
}
