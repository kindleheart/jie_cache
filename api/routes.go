package api

import (
	"github.com/gin-gonic/gin"
	"jie_cache/api/handlers"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) SetupRouter(engine *gin.Engine) {
	engine.GET("/jie_cache", handlers.HTTPHandler)
}
