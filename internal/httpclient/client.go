package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	InsecureSkipVerify bool
}

func New(config *Config) *http.Client {
	transport, _ := http.DefaultTransport.(*http.Transport)

	httpTransport := transport.Clone()
	httpTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify, //nolint:gosec
	}

	return &http.Client{
		Transport: otelhttp.NewTransport(httpTransport),
		Timeout:   time.Second * 30,
	}
}
