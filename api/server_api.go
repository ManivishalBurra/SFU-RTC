package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"meshRTC/models"
	"meshRTC/routes"
)

func StartServer(houseAddress *models.MainHouse) {
	server := gin.New()
	server.Use(CORSMiddleware())
	routes.PrepareRoutes(server, houseAddress)
	err := server.Run(":6303")
	if err != nil {
		fmt.Println(err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
