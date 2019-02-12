// Copyright 2017 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter_shared

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	sslCertFileF = kingpin.Flag("web.ssl-cert-file", "Path to SSL certificate file.").Default("").String()
	sslKeyFileF  = kingpin.Flag("web.ssl-key-file", "Path to SSL key file.").Default("").String()

	landingPage = template.Must(template.New("home").Parse(strings.TrimSpace(`
<html>
<head>
	<title>{{ .name }} exporter</title>
</head>
<body>
	<h1>{{ .name }} exporter</h1>
	<p><a href="{{ .path }}">Metrics</a></p>
</body>
</html>
`)))
)

// RunServer runs server for exporter with given name (it is used on landing page) on given address,
// exposing metrics under given path.
// Function never returns.
func RunServer(name, addr, path string, handler func(http.ResponseWriter, *http.Request)) {
	if (*sslCertFileF == "") != (*sslKeyFileF == "") {
		log.Fatal("One of the flags --web.ssl-cert-file or --web.ssl-key-file is missing to enable HTTPS.")
	}

	ssl := false
	if *sslCertFileF != "" && *sslKeyFileF != "" {
		if _, err := os.Stat(*sslCertFileF); os.IsNotExist(err) {
			log.Fatalf("SSL certificate file does not exist: %s", *sslCertFileF)
		}
		if _, err := os.Stat(*sslKeyFileF); os.IsNotExist(err) {
			log.Fatalf("SSL key file does not exist: %s", *sslKeyFileF)
		}
		ssl = true
	}

	var buf bytes.Buffer
	data := map[string]string{"name": name, "path": path}
	if err := landingPage.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	if ssl {
		runHTTPS(addr, path, http.HandlerFunc(handler), buf.Bytes())
	} else {
		runHTTP(addr, path, http.HandlerFunc(handler), buf.Bytes())
	}
}

func runHTTPS(addr, path string, handler http.Handler, landing []byte) {
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Write(landing)
	})

	tlsCfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		TLSConfig:    tlsCfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)), // disable HTTP/2
	}
	log.Infof("Starting HTTPS server for https://%s%s ...", addr, path)
	log.Fatal(srv.ListenAndServeTLS(*sslCertFileF, *sslKeyFileF))
}

func runHTTP(addr, path string, handler http.Handler, landing []byte) {
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landing)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Infof("Starting HTTP server for http://%s%s ...", addr, path)
	log.Fatal(srv.ListenAndServe())
}
