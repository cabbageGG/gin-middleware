package jaeger

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

var (
	AppName        = "test"
	JaegerHostPort = "127.0.0.1:6831"
	JaegerOpen     = true
)

var tracerCloser io.Closer

// tracerCloser io.Closer can be used in shutdown hooks to ensure that the internal
// queue of the Reporter is drained and all buffered spans are submitted to collectors.
func Close() {
	tracerCloser.Close()
}

func config() {
	if os.Getenv("AppName") != "" {
		AppName = os.Getenv("AppName")
	}

	if os.Getenv("JaegerHostPort") != "" {
		JaegerHostPort = os.Getenv("JaegerHostPort")
	}

	if os.Getenv("JaegerOpen") == "false" {
		JaegerOpen = false
	}
}

func init() {

	config()

	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: JaegerHostPort, // config your jaeger-agent host port
		},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		AppName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		jLogger.Infof("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	//defer closer.Close()
	tracerCloser = closer
}

func SetUp() gin.HandlerFunc {

	return func(c *gin.Context) {

		var serverSpan opentracing.Span

		if JaegerOpen == true {
			operationName := c.Request.Method + " " + c.Request.URL.Path
			wireContext, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
			serverSpan = opentracing.StartSpan(
				operationName,
				ext.RPCServerOption(wireContext),
				opentracing.Tag{Key: string(ext.Component), Value: "HTTP"})
			defer serverSpan.Finish()
			c.Set("Tracer", opentracing.GlobalTracer())
			c.Set("ParentSpanContext", serverSpan.Context())
		}

		c.Next()

		// add tags
		if JaegerOpen == true {
			serverSpan.SetTag("http.status_code", c.Writer.Status())
		}
	}
}
