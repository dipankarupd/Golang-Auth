package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dipankarupd/authapp/pkg/db"
	"github.com/dipankarupd/authapp/pkg/models"
	"github.com/dipankarupd/authapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = db.OpenCollection(db.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	return string(bytes)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))

	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password incorrect")
		check = false
	}
	return check, msg

}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, bson.M{"error": err.Error()})
			return
		}
		// for validation with the user model
		validationError := validate.Struct(user)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, bson.M{"error": validationError.Error()})
			return
		}

		// check for the email and phone number if exist or not:
		// create a var count for checking

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, bson.M{"error": "could not fetch from db"})
			return
		}

		// check if count is more
		//if more, the user already exist:
		if count > 0 {
			c.JSON(http.StatusInternalServerError, bson.M{"error": "email already exist"})
		}

		// encrypt the password:
		password := HashPassword(*user.Password)
		user.Password = &password

		// get the timestamp for created at and updated at
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// give the id
		user.Id = primitive.NewObjectID()

		// user_id to access the user:
		user.User_id = user.Id.Hex()

		// generate the access and refresh tokens:
		accessToken, refreshToken, _ := utils.GenerateTokens(*user.Email, *user.Name, *user.UserType, user.User_id)

		user.AccessToken = &accessToken
		user.RefreshToken = &refreshToken

		res, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			msg := fmt.Sprintf("User not created:")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		c.JSON(http.StatusOK, res)
	}
}

func Signin() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		defer cancel()
		// decode the JSON
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)

		// check if the password matches, decode password:
		passwordValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if passwordValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user exist"})
		}

		// generate the tokens:
		accessToken, refrestToken, _ := utils.GenerateTokens(
			*foundUser.Email,
			*foundUser.Name,
			*foundUser.UserType,
			foundUser.User_id,
		)

		utils.UpdateAllToken(accessToken, refrestToken, foundUser.User_id)

		// at last recheck with id
		err := userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		}

		// finally give the response
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		// check if the current user is an admin
		// only admin can get all users:
		if err := utils.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// pagination default 10 record per page

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))

		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		// aggregation pipelines

		// match
		matchStage := bson.D{{
			Key: "$match", Value: bson.D{{}},
		}}

		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}

		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"tota;_count", 1},
					{"user-items", bson.D{
						{"$slice", []interface{}{"$data", startIndex, recordPerPage}},
					}},
				},
			},
		}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured"})
		}

		defer cancel()

		var allUsers []bson.M

		err = result.All(ctx, &allUsers)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, allUsers[0])

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		userId := c.Param("id")

		// match to the user
		// only admin or the user with the userId can access this
		if err := utils.MatchUserType(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// create a context:
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, user)
	}

}
