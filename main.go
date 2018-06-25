// Copyright (c) 2017 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package main

import (
	"flag"
	"fmt"
	"github.com/aristanetworks/glog"
	"github.com/aristanetworks/goarista/gnmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

// TODO: Make this more clear
var help = `Usage of gnmi:
gnmi -addr [<VRF-NAME>/]ADDRESS:PORT [options...]
  capabilities
  get PATH+
  subscribe PATH+
  ((update|replace PATH JSON)|(delete PATH))+
`

func usageAndExit(s string) {
	flag.Usage()
	if s != "" {
		fmt.Fprintln(os.Stderr, s)
	}
	os.Exit(1)
}

func main() {
	cfg := &gnmi.Config{}
	flag.StringVar(&cfg.Addr, "addr", "", "Address of gNMI gRPC server with optional VRF name")
	flag.StringVar(&cfg.CAFile, "cafile", "", "Path to server TLS certificate file")
	flag.StringVar(&cfg.CertFile, "certfile", "", "Path to client TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "keyfile", "", "Path to client TLS private key file")
	flag.StringVar(&cfg.Password, "password", "", "Password to authenticate with")
	flag.StringVar(&cfg.Username, "username", "", "Username to authenticate with")
	flag.BoolVar(&cfg.TLS, "tls", false, "Enable TLS")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, help)
		flag.PrintDefaults()
	}
	flag.Parse()
	if cfg.Addr == "" {
		usageAndExit("error: address not specified")
	}

	args := flag.Args()

	gi, err := newGNMIInfo(cfg)
	if err != nil {
		glog.Fatal(err)
	}

	go func() {
		foo := newGNMICollector()
		prometheus.MustRegister(foo)
		http.Handle("/metrics", promhttp.Handler())
		glog.Info("Begining to serve on port :8082")
		glog.Fatal(http.ListenAndServe(":8082", nil))
	}()

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "capabilities":
			gi.handleOpCapabilities(args[i+1:])
			return
		case "get":
			gi.handleOpGet(args[i+1:])
			return
		case "subscribe":
			gi.handleOpSubscribe(args[i+1:])
		default:
			usageAndExit(fmt.Sprintf("error: unknown operation %q", args[i]))
		}
	}

}
