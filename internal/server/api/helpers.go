package api

import (
	"github.com/gin-gonic/gin"
)

func writeError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, ErrorResponse{Error: message})
}
