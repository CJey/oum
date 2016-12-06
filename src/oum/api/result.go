package api

import (
	js "github.com/bitly/go-simplejson"
)

type Result struct {
	Code    uint64
	Data    *js.Json
	Message string
}

func (r *Result) Bytes() []byte {
	out := js.New()
	out.Set("code", r.Code)
	if r.Code == 0 {
		if r.Data == nil {
			out.Set("data", struct{}{})
		} else {
			out.Set("data", r.Data)
		}
	} else {
		out.Set("data", struct{}{})
		out.Set("msg", r.Message)
	}
	s, err := out.Encode()
	if err == nil {
		return s
	}
	out = js.New()
	out.Set("code", ERR_OUTPUT)
	out.Set("data", struct{}{})
	out.Set("msg", err.Error())
	s, _ = out.Encode()
	return s
}

func (r *Result) String() string {
	return string(r.Bytes())
}
