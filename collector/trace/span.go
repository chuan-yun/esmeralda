package trace

import (
	"encoding/json"
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
	RelatedAPI        string             `json:"relatedApi,omitempty"`
	SelfAPI           string             `json:"selfApi,omitempty"`
	InsertTime        string             `json:"insertTime,omitempty"`
}

type Spans []Span

const (
	ServerSpan = iota
	ClientSpan = iota
)
