package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strings"
	"time"

	"github.com/enix/x509-certificate-exporter/v2/internal"
	getopt "github.com/pborman/getopt/v2"
	log "github.com/sirupsen/logrus"
)

func main() {
	help := getopt.BoolLong("help", 'h', "show this help message and exit")
	version := getopt.BoolLong("version", 'v', "show version info and exit")
	listenAddress := getopt.StringLong("listen-address", 'b', ":9793", "address on which to bind and expose metrics")
	debug := getopt.BoolLong("debug", 0, "enable debug mode")
	trimPathComponents := getopt.IntLong("trim-path-components", 0, 0, "remove <n> leading component(s) from path(s) in label(s)")
	exposeRelativeMetrics := getopt.BoolLong("expose-relative-metrics", 0, "expose additionnal metrics with relative durations instead of absolute timestamps")
	exposeErrorMetrics := getopt.BoolLong("expose-per-cert-error-metrics", 0, "expose additionnal error metric for each certificate indicating wether it has failure(s)")
	exposeLabels := getopt.StringLong("expose-labels", 'l', "one or more comma-separated labels to enable (defaults to all if not specified)")
	profile := getopt.BoolLong("profile", 0, "optionally enable a pprof server to monitor cpu and memory usage at runtime")

	maxCacheDuration := durationFlag(0)
	getopt.FlagLong(&maxCacheDuration, "max-cache-duration", 0, "maximum cache duration for kube secrets. cache is per namespace and randomized to avoid massive requests.")

	files := stringArrayFlag{}
	getopt.FlagLong(&files, "watch-file", 'f', "watch one or more x509 certificate file")

	directories := stringArrayFlag{}
	getopt.FlagLong(&directories, "watch-dir", 'd', "watch one or more directory which contains x509 certificate files (not recursive)")

	yamls := stringArrayFlag{}
	getopt.FlagLong(&yamls, "watch-kubeconf", 'k', "watch one or more Kubernetes client configuration (kind Config) which contains embedded x509 certificates or PEM file paths")

	getopt.Parse()

	if *help {
		getopt.Usage()
		return
	}

	if *version {
		fmt.Fprintf(os.Stderr, "version %s\n", internal.Version)
		return
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *profile {
		log.Infoln("pprof server listening on :6060")
		go http.ListenAndServe(":6060", nil)
	}

	exporter := internal.Exporter{
		ListenAddress:         *listenAddress,
		Files:                 files,
		Directories:           directories,
		YAMLs:                 yamls,
		YAMLPaths:             internal.DefaultYamlPaths,
		TrimPathComponents:    *trimPathComponents,
		MaxCacheDuration:      time.Duration(maxCacheDuration),
		ExposeRelativeMetrics: *exposeRelativeMetrics,
		ExposeErrorMetrics:    *exposeErrorMetrics,
	}

	if getopt.Lookup("expose-labels").Seen() {
		exporter.ExposeLabels = strings.Split(*exposeLabels, ",")
	}

	log.Infof("starting %s version %s", path.Base(os.Args[0]), internal.Version)
	rand.Seed(time.Now().UnixNano())
	exporter.ListenAndServe()
}
