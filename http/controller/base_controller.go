package controller

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lostyear/go-toolkits/http/response"
)

// BaseController for gin framework
type BaseController struct{}

// JSON return json response by default response
func (ctl BaseController) JSON(c *gin.Context, resp *response.DefaultResponse) {
	c.JSON(resp.Status, resp)
}

// DefaultURLParamStr get url param.
// if param is empty, it will return default string.
func (ctl BaseController) DefaultURLParamStr(c *gin.Context, field string, defaultVal string) string {
	val := c.Param(field)
	if val == "" {
		return defaultVal
	}

	return val
}

// MustURLParamStr get url param string,
// if param is empty, it will panic.
func (ctl BaseController) MustURLParamStr(c *gin.Context, field string) string {
	val := c.Param(field)
	if val == "" {
		panic(response.NewBadRequestResponse(fmt.Sprintf(
			"url param[%s] is blank", field)))
	}

	return val
}

// MustURLParamInt64 get url param, and convert it to int64.
// if param is empty, it will panic;
// if param val can not convert to int64 it will panic.
func (ctl BaseController) MustURLParamInt64(c *gin.Context, field string) int64 {
	strval := ctl.MustURLParamStr(c, field)
	val, err := strconv.ParseInt(strval, 10, 64)
	if err != nil {
		panic(response.NewBadRequestResponse(
			fmt.Sprintf(
				"cannot convert %s to int64", strval)))
	}

	return val
}

// MustURLParamInt get url param, and convert it to int.
// if param is empty, it will panic;
// if param val can not convert to int it will panic.
func (ctl BaseController) MustURLParamInt(c *gin.Context, field string) int {
	return int(ctl.MustURLParamInt64(c, field))
}

// DefaultQueryInt64 get query param, and convert it to int64,
// if query is empty use defaultVal instand,
// if query val can not convert to int64 it will panic.
func (ctl BaseController) DefaultQueryInt64(c *gin.Context, key string, defaultVal int64) int64 {
	strval := c.Query(key)
	if strval == "" {
		return defaultVal
	}

	val, err := strconv.ParseInt(strval, 10, 64)
	if err != nil {
		panic(response.NewBadRequestResponse(fmt.Sprintf(
			"cannot convert [%s] to int64", strval)))
	}

	return val
}

// DefaultQueryInt get query param string, and convert it to int.
// if query is empty use defaultVal instand,
// if query val can not convert to int it will panic.
func (ctl BaseController) DefaultQueryInt(c *gin.Context, key string, defaultVal int) int {
	return int(ctl.DefaultQueryInt64(
		c, key, int64(defaultVal),
	))
}

// MustQueryStr get query param string,
// if query is empty, it will panic.
func (ctl BaseController) MustQueryStr(c *gin.Context, key string) string {
	val := c.Query(key)
	if val == "" {
		panic(response.NewBadRequestResponse(fmt.Sprintf(
			"query param[%s] is necessary", key)))
	}

	return val
}

// MustQueryInt64 get query param, and convert it to int64,
// if query is empty, it will panic;
// if query val can not convert to int64 it will panic.
func (ctl BaseController) MustQueryInt64(c *gin.Context, key string) int64 {
	strval := ctl.MustQueryStr(c, key)

	val, err := strconv.ParseInt(strval, 10, 64)
	if err != nil {
		panic(response.NewBadRequestResponse(fmt.Sprintf(
			"cannot convert [%s] to int64", strval)))
	}

	return val
}

// MustQueryInt get query param, and convert it to int,
// if query is empty, it will panic;
// if query val can not convert to int it will panic.
func (ctl BaseController) MustQueryInt(c *gin.Context, key string) int {
	return int(ctl.MustQueryInt64(c, key))
}
