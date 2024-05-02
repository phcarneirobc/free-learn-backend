package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/phcarneirobc/free-learn/auth"
	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Register(user model.User) (*mongo.InsertOneResult, error) {
    hashedPassword, err := auth.HashPassword(user.Password)
    if err != nil {
        return nil, err
    }

    userToInsert := model.User{
        Id:       primitive.NewObjectID(),
        Name:     user.Name,
        Password: hashedPassword,
        Date:     primitive.NewDateTimeFromTime(time.Now()),
    }

    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, err := db.InsertOne(db.Instance.Client, ctx, db.Instance.Dbname, db.UserCollection, userToInsert)
    if err != nil {
        return nil, err
    }

    return result, nil
}



func Login(user model.User) (string, error) {
    collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)

    var result model.User
    err := collection.FindOne(context.Background(), bson.M{"name": user.Name}).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return "", errors.New("user not found")
        }
        return "", err
    }

    match := auth.CheckPasswordHash(user.Password, result.Password)
    if !match {
        return "", errors.New("invalid password")
    }

    token, err := auth.GenerateToken(result)
    if err != nil {
        return "", err
    }

    return token, nil
}