package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/elastic/beats/v7/libbeat/common/transform/typeconv"
)

type entry struct {
	K     string                 `json:"k"`
	Value map[string]interface{} `json:"v"`
}

func (e entry) Decode(to interface{}) error {
	return typeconv.Convert(to, e.Value)
}

type state struct {
	TTL     time.Duration
	Updated time.Time
	Cursor  interface{}
	Meta    interface{}
}

type stateInternal struct {
	TTL     time.Duration
	Updated time.Time
}

func f(d []byte) (regEntry, error) {
	var dec entry
	var st Value

	if err := json.Unmarshal(d, &dec); err != nil {
		panic(err)
	}

	if err := dec.Decode(&st); err != nil {
		log.Fatalf("Failed to read regisry state for '%q', error: %#v", dec.K, err)
		return regEntry{}, nil
	}

	ret := regEntry{
		K: dec.K,
		V: st,
	}
	return ret, nil
}
