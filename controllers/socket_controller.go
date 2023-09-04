package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"meshRTC/models"
	"meshRTC/services"
)

func GetUsers(c *gin.Context, house *models.MainHouse) {
	users := house.Users["test"]
	var response []models.Clients

	for c := range users {
		if users[c] != nil {
			var x models.Clients
			x.User = c
			response = append(response, x)
		}
	}

	c.JSON(200, response)
}

func ServeWs(c *gin.Context, houseAddress *models.MainHouse) {
	fmt.Println("Entered ServeWs")
	roomId := c.Param("roomId")
	userId := c.Param("userId")
	services.ServeWs(c.Writer, c.Request, roomId, userId, houseAddress)
}
