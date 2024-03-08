package jrpc

import (
	"encoding/json"
)

type raw struct {
	bytes []byte
}

func (a *raw) UnmarshalJSON(b []byte) error {
	a.bytes = b
	return nil
}

func (a *raw) MarshalJSON() ([]byte, error) {
	return a.bytes, nil
}

type method struct {
	name string
	args []any
	raw  [][]byte
}

func (m *method) MarshalJSON() (b []byte, err error) {
	payload := []any{m.name}
	if m.args != nil && len(m.args) > 0 {
		payload = append(payload, m.args...)
	}
	return json.Marshal(payload)
}

func (m *method) UnmarshalJSON(b []byte) (err error) {
	// decode raw payload
	var r []raw
	if err = json.Unmarshal(b, &r); err != nil {
		return
	}

	// decode name
	if err = json.Unmarshal(r[0].bytes, &m.name); err != nil {
		return
	}

	// copy args
	m.raw = make([][]byte, len(r)-1)
	for i, arg := range r[1:] {
		m.raw[i] = arg.bytes
	}
	return
}

type Failure struct {
	Error string `json:"error"`
}
