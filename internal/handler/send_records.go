package handler

import (
	"github.com/gin-gonic/gin"
	"fsvchart-notify/internal/service"
)

func HandleGetSendRecords(c *gin.Context) {
	c.JSON(200, service.GetSendRecords())
} 