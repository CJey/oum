package api

import (
	js "github.com/bitly/go-simplejson"
)

func init() {
	register(API_UNDEFINED, func(called string, b base) API {
		return &undefined{
			base: b,
			name: called,
		}
	})
}

type undefined struct {
	base
	name string
}

func (api *undefined) Run() *Result {
	return api.run(api.do)
}

func (api *undefined) do() *js.Json {
	return api.exit(ERR_API_UNDEFINED, "API[%s] Undefined", api.name)
}
