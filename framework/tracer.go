package framework

import (
	"io"

	"github.com/eudore/eudore"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

func NewTracerFunc() eudore.HandlerFunc {
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

	return (&Tracer{
		tracer: tracer,
		closer: closer,
	}).HandlerHTTP
}

func (t *Tracer) HandlerHTTP(ctx eudore.Context) {
	spnCtx, err := t.tracer.Extract(opentracing.HTTPHeaders, ctx.Request().Header)
	if err != nil {
		ctx.Error(err)
		return
	}
	serverSpan := t.tracer.StartSpan("EudoreHTTP", ext.RPCServerOption(spnCtx))
	serverSpan.SetTag("http.user-agent", ctx.GetHeader("User-Agent"))
	serverSpan.SetTag("http.method", ctx.Method())
	serverSpan.SetTag("http.url", ctx.Request().RequestURI)
	ctx.WithContext(opentracing.ContextWithSpan(ctx.GetContext(), serverSpan))
	ctx.Next()
	serverSpan.SetTag("http.status", ctx.Response().Status())
	serverSpan.SetTag("route", ctx.GetParam("route"))
	action := ctx.GetParam("action")
	if action != "" {
		serverSpan.SetTag("action", action)
		serverSpan.SetTag("ram", ctx.GetParam("ram"))
	}
	serverSpan.Finish()

}