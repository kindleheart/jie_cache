package handlers

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
	"jie_cache/cache"
	"jie_cache/pb"
	"log"
	"net/http"
)

func HTTPHandler(c *gin.Context) {
	log.Println(c.Request.Method, c.Request.URL.Path)
	groupName := c.Query("group")
	key := c.Query("key")
	if groupName == "" || key == "" {
		c.String(http.StatusBadRequest, "group and key must can not be empty")
		return
	}

	group := cache.GetGroup(groupName)
	if group == nil {
		c.String(http.StatusNotFound, "no such group: "+groupName)
		return
	}
	val, err := group.Get(key)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// 编码
	body, err := proto.Marshal(&pb.Response{Value: val.ByteSlice()})
	if err != nil {
		c.String(http.StatusInternalServerError, "proto marshal fail")
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", body)
}
