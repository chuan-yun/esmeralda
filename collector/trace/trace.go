package trace

import (
	"encoding/json"
	"errors"

	"github.com/chuan-yun/esmeralda/util"
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
