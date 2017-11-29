package util

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

func RequestBodyToString(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)

	return buf.String()
}

func IP(r *http.Request) string {
	ips := Proxy(r)
	if len(ips) > 0 && ips[0] != "" {
		rip := strings.Split(ips[0], ":")
		return rip[0]
	}
	ip := strings.Split(r.RemoteAddr, ":")
	if len(ip) > 0 {
		if ip[0] != "[" {
			return ip[0]
		}
	}

	return "127.0.0.1"
}

func Proxy(r *http.Request) []string {
	if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}
