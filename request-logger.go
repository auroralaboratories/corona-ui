package main

import (
	"github.com/urfave/negroni"
	"net/http"
	"time"
)

type RequestLogger struct {
}

func NewRequestLogger() *RequestLogger {
	return &RequestLogger{}
}

func (self *RequestLogger) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, req)

	response := rw.(negroni.ResponseWriter)
	status := response.Status()
	duration := time.Since(start)

	if status < 400 {
		log.Debugf("[HTTP %d] %s to %v took %v", status, req.Method, req.URL, duration)
	} else {
		log.Warningf("[HTTP %d] %s to %v took %v", status, req.Method, req.URL, duration)
	}
}
