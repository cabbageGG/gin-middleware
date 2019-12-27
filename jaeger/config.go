package jaeger

import (
	"io"
	"os"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
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

func init() {

	if os.Getenv("AppName") != "" {
		AppName = os.Getenv("AppName")
	}

	if os.Getenv("JaegerHostPort") != "" {
		JaegerHostPort = os.Getenv("JaegerHostPort")
	}

	if os.Getenv("JaegerOpen") == "false" {
		JaegerOpen = false
	}

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

	// Initialize tracer
	closer, err := cfg.InitGlobalTracer(
		AppName,
	)
	if err != nil {
		return
	}
	//defer closer.Close()
	tracerCloser = closer
}
