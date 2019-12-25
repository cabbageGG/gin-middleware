# example for jaeger
```go
package main

import (
	"encoding/json"
	"log"

	"github.com/cabbageGG/gin-middleware/jaeger"
	"github.com/gin-gonic/gin"
)


type HttpRes struct {
	Message string `json: "message"`
}

func index(c *gin.Context) {
	// call child
	ops := jaeger.HttpOps{
		Method: "GET",
		Url:    "http://localhost:8080/child",
	}
	content, err := jaeger.HttpDo(c, ops)
	if err != nil {
		log.Printf("call /child error: %v", err)
		c.JSON(500, gin.H{"message": "call /child error"})
	}
	res := HttpRes{}
	err = json.Unmarshal(content, &res)
	if err != nil {
		log.Printf("call /child  data unmarshal error %v", err)
		c.JSON(500, gin.H{"message": "call /child  data unmarshal error"})
	}
	c.JSON(200, gin.H{"message": res.Message})
}

func child(c *gin.Context) {
	c.JSON(200, gin.H{"message": "call child success!"})
}

func main() {
	r := gin.New()
	r.Use(jaeger.SetUp())

	r.GET("/", index)
	r.GET("/child", child)

	r.Run(":8080")
}
```


# example for prometheus
```go
package main

import (
        "github.com/gin-gonic/gin"
        ginprom "github.com/cabbageGG/gin-middleware/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
        r := gin.Default()

        r.Use(ginprom.NewGinPrometheus(r).Middleware()) // 添加prometheus 监控
        r.GET("/metrics", gin.WrapH(promhttp.Handler()))
        
        r.GET("/", func(c *gin.Context){
                c.JSON(200, gin.H{"message": "success"})
        })
        r.GET("/test/", func(c *gin.Context){
                c.JSON(500, gin.H{"message": "error"})
        })
        r.Run()
}
```