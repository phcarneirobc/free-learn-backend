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

    // Inicializa o campo "cursos" como um array vazio de primitive.ObjectID
    userToInsert := model.User{
        Id:        primitive.NewObjectID(),
        Name:      user.Name,
        Password:  hashedPassword,
        Date:      primitive.NewDateTimeFromTime(time.Now()),
        Cursos:    []primitive.ObjectID{}, // Corrigido para []primitive.ObjectID
    }

    ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
    result, err := db.InsertOne(db.Instance.Client, ctx, db.Instance.Dbname, db.UserCollection, userToInsert)
    if err != nil {
        return nil, err
    }

    return result, nil
}


type LoginResponse struct {
    ID    primitive.ObjectID
    Token string
    Professor  bool
}

func Login(user model.User) (LoginResponse, error) {
    collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)

    var result model.User
    err := collection.FindOne(context.Background(), bson.M{"name": user.Name}).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return LoginResponse{}, errors.New("user not found")
        }
        return LoginResponse{}, err
    }

    match := auth.CheckPasswordHash(user.Password, result.Password)
    if !match {
        return LoginResponse{}, errors.New("invalid password")
    }

    token, err := auth.GenerateToken(result)
    if err != nil {
        return LoginResponse{}, err
    }
    return LoginResponse{
        ID:    result.Id,
        Token: token,
        Professor:  result.Professor,
    }, nil
}

func AddCourseToUser(userID primitive.ObjectID, courseID primitive.ObjectID) error {
    // Encontrar o usuário pelo ID
    userCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
    filter := bson.M{"_id": userID}
    update := bson.M{"$push": bson.M{"cursos": courseID}} // Corrigido para "cursos"
    _, err := userCollection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        return err
    }
    return nil
}