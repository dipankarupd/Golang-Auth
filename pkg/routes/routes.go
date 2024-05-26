package routes

import (
	"github.com/dipankarupd/authapp/pkg/controllers"
	"github.com/dipankarupd/authapp/pkg/middlewares"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {

	// signup and signin routes:
	r.POST("users/signup", controllers.Signup())
	r.POST("users/signin", controllers.Signin())
}

func UserRoutes(r *gin.Engine) {

	// add middleware to all the routes:
	r.Use(middlewares.Authenticate())
	r.GET("/users", controllers.GetUsers())
	r.GET("/user/{id}", controllers.GetUser())
}
