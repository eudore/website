package tracer

import (
	"context"
	"io"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
	Logger eudore.Logger
	Hander protocol.HandlerHTTP
}

func NewTracer(app *eudore.App, handler protocol.HandlerHTTP) *Tracer {

	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		ServiceName: "eudore-website",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		panic(err)
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)

	return &Tracer{
		tracer: tracer,
		closer: closer,
		Logger: app.Logger,
		Hander: handler,
	}
}

func (t *Tracer) EudoreHTTP(ctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
	headers := make(opentracing.HTTPHeadersCarrier)
	r.Header().Range(func(k, v string) {
		headers.Set(k, v)
	})
	spnCtx, err := t.tracer.Extract(opentracing.HTTPHeaders, headers)
	err = nil
	if err != nil {
		t.Logger.Error(err)
		t.Hander.EudoreHTTP(ctx, w, r)
		return
	}
	serverSpan := t.tracer.StartSpan("EudoreHTTP", ext.RPCServerOption(spnCtx))
	serverSpan.SetTag("http.user-agent", r.Header().Get("User-Agent"))
	serverSpan.SetTag("http.method", r.Method())
	serverSpan.SetTag("http.url", r.RequestURI())
	t.Hander.EudoreHTTP(opentracing.ContextWithSpan(ctx, serverSpan), w, r)
	serverSpan.SetTag("http.status", w.Status())
	serverSpan.Finish()

}

func NewTracerFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		span := opentracing.SpanFromContext(ctx.Context())
		if span == nil || ctx.GetParam("action") == "" {
			return
		}
		serverSpan := span.Tracer().StartSpan(ctx.GetParam("action"), opentracing.ChildOf(span.Context()))
		serverSpan.SetTag("action", ctx.GetParam("action"))
		serverSpan.SetTag("route", ctx.GetParam("route"))
		serverSpan.SetTag("ram", ctx.GetParam("ram"))
		ctx.Next()
		serverSpan.Finish()
	}
}
