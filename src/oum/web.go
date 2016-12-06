package main

import (
	"compress/gzip"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"conf"
	"db"
	"github.com/cjey/slog"
	"oum/api"
	"utils"
	"webfiles"
)

func getWebCert() tls.Certificate {
	cert := conf.Web.Cert
	key := conf.Web.CertKey
	if len(cert)*len(key) > 0 {
		crt, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		return crt
	}
	if len(cert)|len(key) == 0 {
		cert = "/var/lib/oum/web.crt"
		key = "/var/lib/oum/web.key"
		_, cerr := os.Stat(cert)
		_, kerr := os.Stat(key)
		if os.IsNotExist(cerr) && os.IsNotExist(kerr) {
			cakey, err := genCAKey(false)
			if err != nil {
				slog.Emerg(err.Error())
				os.Exit(1)
			}
			cacrt, err := genCACert(cakey, "CN=OUM Web")
			if err != nil {
				slog.Emerg(err.Error())
				os.Exit(1)
			}
			fallback := func() tls.Certificate {
				crt, err := tls.X509KeyPair([]byte(cacrt), []byte(cakey))
				if err != nil {
					slog.Emerg(err.Error())
					os.Exit(1)
				}
				return crt
			}
			err = ioutil.WriteFile(cert, []byte(cacrt), 0400)
			if err != nil {
				return fallback()
			}
			err = ioutil.WriteFile(key, []byte(cakey), 0400)
			if err != nil {
				return fallback()
			}
		}
		crt, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		return crt
	}

	if len(cert) == 0 {
		slog.Emergf("must given cert")
	} else {
		slog.Emergf("must given certkey")
	}
	os.Exit(1)
	return tls.Certificate{}
}

func web() {
	if len(conf.Web.Restore) > 0 {
		err := webfiles.RestoreAssets(conf.Web.Restore, "")
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		return
	}
	var lsn net.Listener
	laddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%d", conf.Web.Port))
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	tcplsn, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	var srvtype string
	if conf.Web.HTTPS {
		srvtype = "https"
		cert := getWebCert()
		tc := &tls.Config{
			PreferServerCipherSuites: true,
			Certificates:             []tls.Certificate{cert},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,

				// support fallback tls1.0
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		rand.Read(tc.SessionTicketKey[:])
		tc.BuildNameToCertificate()
		lsn = tls.NewListener(tcplsn, tc)
	} else {
		srvtype = "http"
		lsn = tcplsn
	}

	srv := &http.Server{
		Handler:        http.HandlerFunc(defaultHandler),
		MaxHeaderBytes: 1 << 20,
		ConnState: func(c net.Conn, cs http.ConnState) {
			switch cs {
			case http.StateActive:
				c.SetDeadline(time.Time{})
			case http.StateNew:
				fallthrough
			case http.StateIdle:
				c.SetDeadline(time.Now().Add(15 * time.Second))
			}
		},
	}

	slog.Infof("Service running, %s://%s", srvtype, lsn.Addr().String())
	err = srv.Serve(lsn)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/"):
		webAPI(r.URL.Path[5:], w, r)
	case strings.HasPrefix(r.URL.Path, "/download/"):
		webDownloader(r.URL.Path[10:], w, r)
	case len(conf.Web.Root) > 0:
		webStaticHandler(w, r)
	default:
		webBinDataHandler(w, r)
	}
}

func webBinDataHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[1:]
	if len(file) == 0 {
		file = "index.html"
	}
	if file[len(file)-1] == '/' {
		file += "index.html"
	}

	fi, err := webfiles.AssetInfo(file)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			slog.Noticef("%s, %s", r.URL.Path, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// handle 304
	modtime := fi.ModTime()
	var modsince time.Time
	_modsince := r.Header.Get("If-Modified-Since")
	if len(_modsince) > 0 {
		modsince, err = time.Parse(time.RFC1123, _modsince)
		if err != nil {
			modsince = time.Time{}
		}
	}
	if modtime.Equal(modsince) {
		w.WriteHeader(http.StatusNotModified)
		w.Header().Set("Last-Modified", modtime.Format(time.RFC1123))
		return
	}

	body, err := webfiles.Asset(file)
	if err != nil {
		slog.Noticef("%s, %s", r.URL.Path, err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Last-Modified", modtime.Format(time.RFC1123))

	// handle content-type
	ct := mime.TypeByExtension(path.Ext(file))
	if len(ct) == 0 {
		ct = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)

	// handle gzip
	var fgzip bool
	acs := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, ac := range acs {
		ac = strings.TrimSpace(ac)
		if ac == "gzip" {
			fgzip = true
			break
		}
	}

	if fgzip {
		tmp := strings.Split(ct, "/")
		switch tmp[0] {
		case "audio", "image", "video":
			fgzip = false
		}
	}

	if fgzip && len(body) > 1000 {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		gw.Write(body)
		gw.Flush()
	} else {
		w.Write(body)
	}
}

func webStaticHandler(w http.ResponseWriter, r *http.Request) {
	file := conf.Web.Root + r.URL.Path
	if file[len(file)-1] == '/' {
		file += "index.html"
	}
	fi, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			slog.Noticef("%s, %s", r.URL.Path, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// handle 304
	modtime := fi.ModTime()
	var modsince time.Time
	_modsince := r.Header.Get("If-Modified-Since")
	if len(_modsince) > 0 {
		modsince, err = time.Parse(time.RFC1123, _modsince)
		if err != nil {
			modsince = time.Time{}
		}
	}
	if modtime.Equal(modsince) {
		w.WriteHeader(http.StatusNotModified)
		w.Header().Set("Last-Modified", modtime.Format(time.RFC1123))
		return
	}
	f, err := os.Open(file)
	if err != nil {
		slog.Noticef("%s, %s", r.URL.Path, err.Error())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Last-Modified", modtime.Format(time.RFC1123))

	// handle content-type
	var ct string
	if r.FormValue("attach") == "true" {
		ct = "application/octet-stream"
		w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(file))
	} else {
		ct = mime.TypeByExtension(path.Ext(file))
	}
	if len(ct) == 0 {
		ct = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)

	// handle gzip
	var fgzip bool
	acs := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, ac := range acs {
		ac = strings.TrimSpace(ac)
		if ac == "gzip" {
			fgzip = true
			break
		}
	}

	if fgzip {
		tmp := strings.Split(ct, "/")
		switch tmp[0] {
		case "audio", "image", "video":
			fgzip = false
		}
	}

	if fgzip && fi.Size() > 256 {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		io.Copy(gw, f)
		gw.Flush()
	} else {
		io.Copy(w, f)
	}
}

func webAPI(apiName string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	res := api.New(apiName, w, r).Run()
	out := res.Bytes()

	if res.Code > 0 {
		slog.Noticef("API[%s] request failure: %s", apiName, out)
	}

	// handle gzip
	acs := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	var fgzip bool
	for _, ac := range acs {
		ac = strings.TrimSpace(ac)
		if ac == "gzip" {
			fgzip = true
			break
		}
	}

	if fgzip && len(out) > 256 {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		gw.Write(out)
		gw.Flush()
	} else {
		w.Write(out)
	}
}

func webDownloader(name string, w http.ResponseWriter, r *http.Request) {
	switch name {
	case "config":
		//user := r.FormValue("user")
		platform := r.FormValue("os")
		dev := r.FormValue("dev")
		name := r.FormValue("name")
		attach := r.FormValue("attach")

		DB := db.Get()

		var host, port, srvpath string
		var err error
		if len(dev) > 0 {
			err = DB.QueryRow(`
                select remote, port, conffile from ovpn
                where dev=?
            `, dev).Scan(&host, &port, &srvpath)
		} else {
			err = DB.QueryRow(`
                select dev, remote, port, conffile from ovpn
            `).Scan(&dev, &host, &port, &srvpath)
		}

		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("404 Not Found"))
			} else {
				w.Header().Set("X-ERROR", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 Internal Error"))
			}
			return
		}

		remote := fmt.Sprintf("remote %s %s\n", host, port)
		lines, err := utils.StripConf(srvpath)
		if err != nil {
			w.Header().Set("X-ERROR", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Error"))
			return
		}

		res := fetchPatternClient(lines, platform, remote)

		if len(name) == 0 {
			name = dev
		}

		if platform == "linux" {
			name += ".conf"
		} else {
			name += ".ovpn"
		}

		if attach == "true" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename="+name)
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		w.Write(res)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
