package api

import (
	"regexp"
	"sync"
)

var apis_lock sync.Mutex
var apis map[string]func(string, base) API
var RegexpMap map[string]*regexp.Regexp

func register(name string, creator func(string, base) API) {
	apis_lock.Lock()
	defer apis_lock.Unlock()
	if apis == nil {
		apis = make(map[string]func(string, base) API)
	}
	apis[name] = creator
}

func init() {
	regexprecompile()
}
func regexprecompile() { //正则表达式预编译
	RegexpMap = make(map[string]*regexp.Regexp, 4)
	RegexpMap["MatchSMSCode"] = regexp.MustCompile(`^[0-9a-zA-Z]{4,8}$`)
	RegexpMap["MatchCallCode"] = regexp.MustCompile(`^[0-9]{4,6}$`)
	RegexpMap["MatchSource"] = regexp.MustCompile(`^\w{1,20}$`)
	RegexpMap["MatchPhoneNum"] = regexp.MustCompile(`^(\+86|86|0086)?\d{11}$`)
	RegexpMap["MatchEmail"] = regexp.MustCompile(`
			^([\w\!\#\$\%\^\&\*\-\+]+)([\.\w!#\$%\^&\*\-\+]+)*@([a-zA-Z]([-a-zA-Z\d]{0,61}[a-zA-Z\d])?(\.[a-zA-Z]([-a-zA-Z\d]{0,61}[a-zA-Z\d])?)*\.?)$
	`)
}
