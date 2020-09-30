package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: scaffod

// base controller for gin framework
type BaseController struct{}

// TODO: return controller
// 1. return data
// 2. return 500
// 3. return 404
// 4. return 401

// TODO: get params
// methods:
// 1. must
// 2. should
// 3. default
// types:
// 1. string
// 2. int
// 3. int64
// 4. bool

func (*BaseController) JSONData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
		"msg":    "success",
	})
}

func (*BaseController) BadRequest(c *gin.Context, message string, a ...interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status": "bad request",
		"msg":    fmt.Sprintf(message, a...),
	})
}
