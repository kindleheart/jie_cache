package app

import "github.com/gin-gonic/gin"

func NewGinEngine(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}
	return gin.Default()
}
