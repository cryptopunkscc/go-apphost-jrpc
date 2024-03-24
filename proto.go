package jrpc

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

type Failure struct {
	Error string `json:"error"`
}
