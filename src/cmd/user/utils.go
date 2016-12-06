package user

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

func StdUsername(username string) (name string, device string) {
	ok, err := regexp.MatchString(`^!?[a-zA-Z0-9\-\._]+(%[a-zA-Z0-9\-\._]+)?$`, username)
	if err != nil {
		return
	}
	if !ok {
		return
	}
	tmp := strings.SplitN(username, "%", 2)
	name = strings.ToLower(tmp[0])
	if len(tmp) > 1 {
		device = strings.ToLower(tmp[1])
	}
	if name[0] == '!' {
		name = name[1:]
	}
	return
}

func StdPassword(password string) (code string, pass string) {
	tmp := strings.SplitN(password, "%", 2)
	code = tmp[0]
	if len(tmp) == 2 {
		pass = tmp[1]
	}
	return
}

func hashPassword(pass string) string {
	salt := make([]byte, 8)
	rand.Read(salt)
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(pass))
	res := mac.Sum(nil)
	return fmt.Sprintf("%x:%x", salt, res)
}
