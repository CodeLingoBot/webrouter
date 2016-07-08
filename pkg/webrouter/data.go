package webrouter

import (
	"bytes"
	"encoding/json"
)

type Data struct {
	Sid   int64         `json:"sid,omitempty"`
	Vid   int           `json:"vid,omitempty"`
	Value []interface{} `json:"value,omitempty"`
}

func newData(sid int64, vid int, value interface{}) *Data {
	d := &Data{
		Sid: sid,
		Vid: vid,
	}

	if "" != value {
		d.Value = []interface{}{value}
	}

	return d
}

func newDataValues(sid int64, vid int, value []interface{}) *Data {
	return &Data{
		Sid:   sid,
		Vid:   vid,
		Value: value,
	}
}

func (self *Data) Marshal() string {
	v, _ := json.Marshal(self)
	return string(v)
}

func UnMarshalData(data []byte) (*Data, error) {
	v := &Data{}

	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return v, d.Decode(&v)
}

