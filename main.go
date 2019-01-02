package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/logger"
	"github.com/nocquidant/demo-hello/api"
	"github.com/nocquidant/demo-hello/env"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.Init("hello", true, false, ioutil.Discard)

	// Load environment
	env.Load()

	logger.Info("Environment used...")
	logger.Infof(" - env.version: %s\n", env.VERSION)
	logger.Infof(" - env.build: %s\n", env.GITCOMMIT)
	logger.Infof(" - env.name: %s\n", env.NAME)
	logger.Infof(" - env.port: %d\n", env.PORT)
	logger.Infof(" - env.remoteUrl: %s\n", env.REMOTE_URL)

	logger.Infof("HTTP service: %s, is running using port: %d\n", env.NAME, env.PORT)
	logger.Info("Available GET endpoints are: '/health', '/hello', '/refresh' and '/remote'")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", api.HandlerHealth)
	mux.HandleFunc("/hello", api.HandlerHello)
	mux.HandleFunc("/remote", api.HandlerRemote)
	mux.HandleFunc("/refresh", api.HandlerRefresh)

	mux.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":"+strconv.Itoa(env.PORT), mux)
}
