package api

import (
	"db"

	js "github.com/bitly/go-simplejson"
)

// %s/serv_list/<new api name>/g

func init() {
	register("serv.list", func(called string, b base) API {
		return &serv_list{
			base: b,
			name: called,
		}
	})
}

type serv_list struct {
	base
	name string
}

func (api *serv_list) Run() *Result {
	return api.run(api.do)
}

func (api *serv_list) do() *js.Json {
	DB := db.Get()
	rows, err := DB.Query(`
        select dev,name,memo from ovpn
    `)
	if err != nil {
		return api.exit(ERR_UNEXPECTED, err.Error())
	}

	serving := []*js.Json{}
	err = db.RangeRows(rows, func() error {
		var dev, name, memo string
		err = rows.Scan(&dev, &name, &memo)
		if err != nil {
			return err
		}
		item := js.New()
		item.Set("dev", dev)
		item.Set("alias", name)
		item.Set("memo", memo)
		serving = append(serving, item)
		return nil
	})
	if err != nil {
		return api.exit(ERR_UNEXPECTED, err.Error())
	}

	data := js.New()
	data.Set("serving", serving)

	return data
}
