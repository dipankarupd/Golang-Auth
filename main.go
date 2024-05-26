package main

import (
	"os"

	"github.com/dipankarupd/authapp/pkg/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// get the port
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)

	router.Run(":" + port)
}
