package trace

import (
	"encoding/json"
	"errors"

	"chuanyun.io/esmeralda/util"
)

func ToSpans(data string) (*[]Span, error) {
	var spans []Span
	err := json.Unmarshal([]byte(data), &spans)
	if err != nil {
		return nil, err
	}

	if len(spans) <= 0 || spans[0].ID == nil {
		return nil, errors.New(util.Message("decode no span"))
	}

	return &spans, nil
}

type MessagePack struct {
	Body      string          `json:"body"`
	Host      json.RawMessage `json:"hostname"`
	Project   json.RawMessage `json:"project"`
	Path      json.RawMessage `json:"fp"`
	Timestamp json.RawMessage `json:"timestamp"`
}

func Unwrap(data []byte) (MessagePack, error) {
	var s MessagePack

	err := json.Unmarshal(data, &s)

	if err != nil {
		return s, err
	}
	return s, nil
}

func GetMessageBody(data []byte) (string, error) {
	s, err := Unwrap(data)
	if err != nil {
		return "", err
	}

	return s.Body, nil
}
