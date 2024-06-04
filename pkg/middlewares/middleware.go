package middlewares

import (
	"fmt"
	"net/http"

	"github.com/dipankarupd/authapp/pkg/utils"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("no authorization token provided")})
			c.Abort()
			return
		}

		// check for the token
		var claims, err = utils.ValidateToken(clientToken)

		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("user_type", claims.UserType)
		c.Set("uid", claims.Id)
		c.Next()
	}

}
