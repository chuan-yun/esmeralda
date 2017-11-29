package trace

type Endpoint struct {
	ServiceName string `json:"serviceName"`
	Ipv4        string `json:"ipv4"`
	Port        int16  `json:"port"`
}

type Annotation struct {
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Endpoint  Endpoint
}

type BinaryAnnotation struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Endpoint Endpoint
}

type Span struct {
	ID                string             `json:"id"`
	Timestamp         int64              `json:"timestamp"`
	ParentID          string             `json:"parentId"`
	Duration          int64              `json:"duration"`
	Name              string             `json:"name"`
	TraceID           string             `json:"traceId"`
	Annotations       []Annotation       `json:"annotations"`
	BinaryAnnotations []BinaryAnnotation `json:"binaryAnnotations"`
	Version           string             `json:"version"`
	RelatedAPI        string             `json:"relatedApi"`
	InsertTime        string             `json:"insertTime"`
}
