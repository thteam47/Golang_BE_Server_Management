package drive

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type MongoDB struct {
	DB *mongo.Database
}

var Mongo = &MongoDB{}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func ConnectMongo(dburl string, dbname string) *MongoDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dburl))
	if err != nil {
		log.Fatal("err connect to db", err.Error())
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	Mongo.DB = client.Database(dbname)

	// password := "admin"
	// passHash, _ := HashPassword(password)
	// user := User {
	// 	Username: "admin",
	// 	Password: passHash,
	// }
	// DB.Collection("user").InsertOne(ctx, user)
	// if err != nil {
	// 	panic(err)
	// }
	return Mongo
}
