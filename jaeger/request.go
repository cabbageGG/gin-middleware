package jaeger

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type HttpOps struct {
	Method string
	Url    string // http://127.0.0.1/?a=1&b=2   contains params
	data   []byte // post data - json marshal
}

// Transport wraps a RoundTripper. If a request is being traced with
// Tracer, Transport will inject the current span into the headers,
// and set HTTP related tags on the span.
type Transport struct {
	// The actual RoundTripper to use for the request. A nil
	// RoundTripper defaults to http.DefaultTransport.
	http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	if !opentracing.IsGlobalTracerRegistered() {
		return rt.RoundTrip(req)
	}

	operationName := "HTTP " + req.Method + " " + req.URL.Path

	span, _ := opentracing.StartSpanFromContext(req.Context(), operationName)
	span.SetTag("span.kind", "client")

	defer span.Finish()

	// 将span的信息，传递到http-header，用于下次请求
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))

	resp, err := rt.RoundTrip(req)

	if err != nil {
		return resp, err
	}
	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	if resp.StatusCode >= http.StatusInternalServerError {
		ext.Error.Set(span, true)
	}

	return resp, nil
}

func HttpDo(c *gin.Context, ops HttpOps) ([]byte, error) {
	// get context.Context
	SpanHttpContext, _ := c.Get("SpanHttpContext")
	ctx, ok := SpanHttpContext.(context.Context)
	if !ok {
		log.Println("ctx not ok", ctx)
		ctx = context.Background()
	}

	client := &http.Client{Transport: &Transport{}}

	req, err := http.NewRequest(ops.Method, ops.Url, bytes.NewBuffer(ops.data))

	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx) // extend existing trace, if any

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
