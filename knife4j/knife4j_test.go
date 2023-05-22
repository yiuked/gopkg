package knife4j

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"testing"
)

func TestKnife4j(t *testing.T) {
	r := gin.Default()
	r.GET("/doc/*any", Handler(Config{RelativePath: "/doc"}))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.DocExpansion("none")))

}
