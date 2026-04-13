package api

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, httpHandler *HTTPHandler, socketHandler *SocketHandler) {
	router.GET("/ws/:id", socketHandler.HandleServerSideSocket)
	router.POST("/push/:id", httpHandler.HandlePushMessage)
	router.GET("/clients", httpHandler.ReturnClients)
	router.GET("/clients/:id", httpHandler.ReturnClientDetails)
	router.GET("/jobs", httpHandler.ReturnJobs)
	router.POST("/register", httpHandler.HandleRegistration)
	router.DELETE("/delete/jobs", httpHandler.HandleDeleteJobsOfClient)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
