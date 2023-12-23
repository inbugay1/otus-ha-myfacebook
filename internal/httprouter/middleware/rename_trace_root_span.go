package middleware

import (
	"net/http"

	"github.com/inbugay1/httprouter"
	"go.opentelemetry.io/otel/trace"
)

func NewRenameTraceRootSpan() httprouter.MiddlewareFunc {
	return func(next httprouter.Handler) httprouter.Handler {
		handler := func(responseWriter http.ResponseWriter, request *http.Request) error {
			ctx := request.Context()

			span := trace.SpanFromContext(ctx)
			name := "HTTP " + request.Method + " " + httprouter.RouteName(ctx)

			span.SetName(name)

			return next.Handle(responseWriter, request) //nolint:wrapcheck
		}

		return httprouter.HandlerFunc(handler)
	}
}
