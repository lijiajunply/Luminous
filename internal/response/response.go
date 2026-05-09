package response

import "github.com/gin-gonic/gin"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ListData struct {
	Total int         `json:"total"`
	Items interface{} `json:"items"`
}

func Success(c *gin.Context, httpStatus int, message string, data interface{}) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
		Data:    nil,
	})
}

func SuccessList(c *gin.Context, httpStatus int, message string, total int, items interface{}) {
	c.JSON(httpStatus, Response{
		Code:    httpStatus,
		Message: message,
		Data: ListData{
			Total: total,
			Items: items,
		},
	})
}
