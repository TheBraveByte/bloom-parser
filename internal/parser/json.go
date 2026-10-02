package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// jsonParser handles a JSON array of objects. The union of object keys, in
// first-seen order, becomes the column list.
type jsonParser struct{}

func (jsonParser) Format() Format { return FormatJSON }

func (jsonParser) Parse(_ context.Context, in Input) (*ParsedDocument, error) {
	dec := json.NewDecoder(bytes.NewReader(in.Content))
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil || t != json.Delim('[') {
		return nil, fmt.Errorf("%w: expected a json array of objects", ErrNotParseable)
	}

	var headers []string
	seen := map[string]bool{}
	var objects [][]kv
	for dec.More() {
		obj, err := decodeObject(dec)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
		for _, e := range obj {
			if !seen[e.k] {
				seen[e.k] = true
				headers = append(headers, e.k)
			}
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("%w as json: %v", ErrNotParseable, err)
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("%w: empty json array", ErrNotParseable)
	}

	rows := make([][]string, len(objects))
	for i, obj := range objects {
		row := make([]string, len(headers))
		vals := make(map[string]string, len(obj))
		for _, e := range obj {
			vals[e.k] = stringifyJSON(e.v)
		}
		for j, h := range headers {
			row[j] = vals[h]
		}
		rows[i] = row
	}
	return normalize(rawTable{headers: headers, rows: rows})
}

type kv struct {
	k string
	v any
}

// decodeObject streams one {...} object preserving key order.
func decodeObject(dec *json.Decoder) ([]kv, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("%w as json: %v", ErrNotParseable, err)
	}
	if t != json.Delim('{') {
		return nil, fmt.Errorf("%w: array elements must be objects", ErrNotParseable)
	}
	var out []kv
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("%w as json: %v", ErrNotParseable, err)
		}
		var v any
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("%w as json: %v", ErrNotParseable, err)
		}
		out = append(out, kv{kt.(string), v})
	}
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("%w as json: %v", ErrNotParseable, err)
	}
	return out, nil
}

// stringifyJSON flattens a decoded scalar to canonical text; non-scalars are
// re-encoded so nothing is silently dropped.
func stringifyJSON(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		return strconv.FormatBool(x)
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}
