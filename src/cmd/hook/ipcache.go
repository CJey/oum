package hook

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"db"
)

type IP struct {
	ip       net.IP
	country  string
	province string
	city     string
	isp      string
}

func (ip *IP) Same(cmp *IP) bool {
	return ip.ip.String() == cmp.ip.String()
}

func (ip *IP) SameCountry(cmp *IP) bool {
	if len(ip.country) == 0 {
		return false
	}
	return ip.country == cmp.country
}

func (ip *IP) Country() string {
	return ip.country
}

func (ip *IP) SameProvince(cmp *IP) bool {
	if len(ip.country)&len(ip.province) == 0 {
		return false
	}
	return ip.country == cmp.country && ip.province == cmp.province
}

func (ip *IP) SameISP(cmp *IP) bool {
	if len(ip.isp) == 0 {
		return false
	}
	return ip.isp == cmp.isp
}

func (ip *IP) Province() string {
	return ip.country + ">" + ip.province
}

func (ip *IP) SameCity(cmp *IP) bool {
	if len(ip.country)&len(ip.province)&len(ip.city) == 0 {
		return false
	}
	return ip.country == cmp.country &&
		ip.province == cmp.province &&
		ip.city == cmp.city
}

func (ip *IP) City() string {
	return ip.country + ">" + ip.province + ">" + ip.city
}

func (ip *IP) ISP() string {
	return ip.isp
}

var httpcli *http.Client = &http.Client{
	Timeout: 5 * time.Second,
}

var ipiplock *sync.Mutex = &sync.Mutex{}
var lastquery time.Time

func NewIP(ip net.IP) (*IP, error) {
	if ip == nil {
		return nil, fmt.Errorf("Invalid ip address")
	}
	ipdot := ip.String()
	ret := &IP{
		ip: ip,
	}
	DB := db.Get()
	var updated time.Time
	err := DB.QueryRow(`
        select country,province,city,isp,updated from ipcache
        where ip=?
    `, ipdot).Scan(&ret.country, &ret.province, &ret.city, &ret.isp, &updated)
	now := time.Now()
	if err == nil {
		if updated.AddDate(0, 1, 0).After(now) {
			return ret, nil
		}
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	ipiplock.Lock()
	defer ipiplock.Unlock()
	// defend access rate
	wait := 300*time.Millisecond - now.Sub(lastquery)
	if wait > 0 {
		time.Sleep(wait)
	}
	var info [5]string
	for i := 0; true; i++ {
		resp, err := httpcli.Get("http://freeapi.ipip.net/" + ipdot)
		lastquery = now
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(body, &info)
			if err != nil {
				return nil, err
			}
			break
		}
		if i >= 2 {
			return nil, fmt.Errorf("ipip.net service expecetion")
		}
		time.Sleep(500 * time.Millisecond)
	}
	ret.country = info[0]
	ret.province = info[1]
	ret.city = info[2]
	ret.isp = info[4]
	_, err = DB.Exec(`
        insert or replace into ipcache
            (ip, country, province, city, isp, updated)
        values
            (?, ?, ?, ?, ?, datetime('now'))
    `, ipdot, info[0], info[1], info[2], info[4])
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ReservedIPv4(ip net.IP) bool {
	n := binary.BigEndian.Uint32(ip)
	return (n >= 167772160 && n <= 184549375) ||
		(n >= 3232235520 && n <= 3232301055) ||
		(n >= 2130706432 && n <= 2147483647) ||
		(n >= 2886729728 && n <= 2887778303) ||
		(n >= 2851995648 && n <= 2852061183)
}
