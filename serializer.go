package rpc

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

type Serializer struct {
	io.ReadWriteCloser
	enc      *json.Encoder
	dec      *json.Decoder
	remoteID id.Identity
}

func (conn *Serializer) RemoteIdentity() id.Identity {
	return conn.remoteID
}

func (conn *Serializer) Encode(value any) (err error) {
	r := value
	switch v := value.(type) {
	case error:
		r = Failure{v.Error()}
	}
	return conn.enc.Encode(r)
}

func (conn *Serializer) Decode(value any) (err error) {
	// decode raw value
	r := raw{}
	if err = conn.dec.Decode(&r); err != nil {
		return
	}

	// try decode as failure
	f := Failure{}
	if err = json.Unmarshal(r.bytes, &f); err == nil && f.Error != "" {
		return errors.New(f.Error)
	}

	// decode value
	return json.Unmarshal(r.bytes, value)
}
