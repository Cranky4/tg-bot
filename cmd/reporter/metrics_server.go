package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func startMetricsHTTPServer(url string, port int) error {
	http.Handle(url, promhttp.Handler())

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "ошибка старта сервера метрик")
	}

	return nil
}
