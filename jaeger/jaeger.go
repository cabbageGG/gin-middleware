package jaeger

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerClient "github.com/uber/jaeger-client-go"
)

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
			c.Set("SpanHttpContext", opentracing.ContextWithSpan(context.Background(), serverSpan))

			spanContext := serverSpan.Context()
			if spanContext, ok := spanContext.(jaegerClient.SpanContext); ok {
				c.Set("trace_id", spanContext.TraceID().String())
				c.Set("span_id", spanContext.SpanID().String())
			}
		}

		c.Next()

		// add tags
		if JaegerOpen == true {
			serverSpan.SetTag("http.status_code", c.Writer.Status())
		}
	}
}
