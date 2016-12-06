package api

import (
	js "github.com/bitly/go-simplejson"
)

// %s/sample/<new api name>/g

func init() {
	register("sample", func(called string, b base) API {
		return &sample{
			base: b,
			name: called,
		}
	})
}

type sample struct {
	base
	name string
}

func (api *sample) Run() *Result {
	return api.run(api.do)
}

func (api *sample) do() *js.Json {
	// optional string
	argA := api.oStr("argA", "default value of argA")
	// optional string array, seprated by ",", trimspace all element and filter the empty element
	argB := api.oStrArray("argB", "default elem1", "default elem2")
	// optional string set, seprated by ",", trimspace all element and filter the empty element, then unique all elements
	argC := api.oStrSet("argC", "default elem1", "default elem2")

	data := js.New()
	data.Set("argA", argA)
	data.Set("argB", argB)
	data.Set("argC", argC)
	return data
}
