package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbName string
var Client *mongo.Client

// to get the mongodb client
func init() {

	// get the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading the .env file")
	}
	connectionString := os.Getenv("MONGO_URI")
	dbName = os.Getenv("DB_NAME")

	// connect to mongodb getting client
	clientOptions := options.Client().ApplyURI(connectionString)

	// connect to mongo
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to mongodb successfully")
	Client = client
}

func OpenCollection(client *mongo.Client, colName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database(dbName).Collection(colName)
	fmt.Println("Collection instance ready")
	return collection
}
