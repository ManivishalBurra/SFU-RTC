package routes

import (
	"github.com/gin-gonic/gin"
	"meshRTC/controllers"
	"meshRTC/models"
)

func PrepareRoutes(router *gin.Engine, houseAddress *models.MainHouse) {
	router.GET("/ws/:roomId/:userId", func(context *gin.Context) {
		controllers.ServeWs(context, houseAddress)
	})
	router.GET("/getsocketusers", func(context *gin.Context) {
		controllers.GetUsers(context, houseAddress)
	})
}
