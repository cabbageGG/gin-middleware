package jaeger

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
			c.Set("SpanContext", opentracing.ContextWithSpan(context.Background(), serverSpan))
		}

		c.Next()

		// add tags
		if JaegerOpen == true {
			serverSpan.SetTag("http.status_code", c.Writer.Status())
		}
	}
}
