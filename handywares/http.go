package handywares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/janstoon/toolbox/tricks"
	"github.com/rs/cors"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type HttpMiddlewareStack middleware.Builder

type PanicRecoverHttpMiddlewareOpt = tricks.InPlaceOption[any]

func (stk *HttpMiddlewareStack) PushPanicRecover(options ...PanicRecoverHttpMiddlewareOpt) *HttpMiddlewareStack {
	return stk.Push(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(rw, req)
		})
	})
}

type BlindLoggerHttpMiddlewareOpt = tricks.InPlaceOption[any]

func (stk *HttpMiddlewareStack) PushBlindLogger(options ...BlindLoggerHttpMiddlewareOpt) *HttpMiddlewareStack {
	return stk.Push(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			log.Printf("requested %s", req.URL)

			next.ServeHTTP(rw, req)
		})
	})
}

type CorsHttpMiddlewareOpt = tricks.InPlaceOption[cors.Options]

func (stk *HttpMiddlewareStack) PushCrossOriginResourceSharingPolicy(
	options ...CorsHttpMiddlewareOpt,
) *HttpMiddlewareStack {
	cfg := cors.Options{}
	cfg = tricks.PtrVal(tricks.ApplyOptions(&cfg,
		tricks.Map(options, func(src CorsHttpMiddlewareOpt) tricks.Option[cors.Options] {
			return src
		})...))

	return stk.Push(cors.New(cfg).Handler)
}

var CorsAllowOrigins = func(origins ...string) CorsHttpMiddlewareOpt {
	return func(s *cors.Options) {
		s.AllowedOrigins = origins
	}
}

var CorsAllowMethods = func(methods ...string) CorsHttpMiddlewareOpt {
	return func(s *cors.Options) {
		s.AllowedMethods = methods
	}
}

var CorsAllowHeaders = func(headers ...string) CorsHttpMiddlewareOpt {
	return func(s *cors.Options) {
		s.AllowedHeaders = headers
	}
}

var CorsDebug = func(debug bool) CorsHttpMiddlewareOpt {
	return func(s *cors.Options) {
		s.Debug = debug
	}
}

type OpenTelemetryHttpMiddlewareOpt = tricks.Option[any]

func (stk *HttpMiddlewareStack) PushOpenTelemetry(
	tracer trace.Tracer, options ...OpenTelemetryHttpMiddlewareOpt,
) *HttpMiddlewareStack {
	return stk.Push(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx, span := tracer.Start(req.Context(), fmt.Sprintf("[%s] %s", req.Method, req.URL.String()))
			defer span.End()

			span.SetAttributes(
				semconv.HTTPRequestMethodKey.String(req.Method),
			)

			traceableReq := req.WithContext(ctx)
			next.ServeHTTP(rw, traceableReq)
		})
	})
}

func (stk *HttpMiddlewareStack) Push(mw middleware.Builder) *HttpMiddlewareStack {
	current := *stk
	if current == nil {
		current = middleware.PassthroughBuilder
	}

	*stk = func(next http.Handler) http.Handler {
		return current(mw(next))
	}

	return stk
}

func (stk *HttpMiddlewareStack) NotNil() middleware.Builder {
	if *stk == nil {
		return middleware.PassthroughBuilder
	}

	return middleware.Builder(*stk)
}
