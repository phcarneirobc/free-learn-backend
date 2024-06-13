package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/phcarneirobc/free-learn/auth"
	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Register(user model.User) (*mongo.InsertOneResult, error) {
	existingUser, err := getUserByEmail(user.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	userToInsert := model.User{
		Id:        primitive.NewObjectID(),
		Email:     user.Email,
		Password:  hashedPassword,
		Professor: user.Professor,
		Date:      primitive.NewDateTimeFromTime(time.Now()),
		Cursos:    []primitive.ObjectID{},
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := db.InsertOne(db.Instance.Client, ctx, db.Instance.Dbname, db.UserCollection, userToInsert)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getUserByEmail(email string) (*model.User, error) {
	email = strings.ToLower(email)
	var user model.User
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

type LoginResponse struct {
	ID        primitive.ObjectID
	Token     string
	Professor bool
}

func Login(user model.User) (LoginResponse, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)

	email := strings.ToLower(user.Email)
	var result model.User
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return LoginResponse{}, errors.New("email not found")
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
		ID:        result.Id,
		Token:     token,
		Professor: result.Professor,
	}, nil
}

func AddCourseToUser(userID primitive.ObjectID, courseID primitive.ObjectID) error {
	userCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	filter := bson.M{"_id": userID}
	update := bson.M{"$push": bson.M{"cursos": courseID}}
	_, err := userCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func GetUserCourses(userID primitive.ObjectID) ([]model.Course, error) {
	userCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	var user model.User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	courseCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	filter := bson.M{"_id": bson.M{"$in": user.Cursos}}
	cursor, err := courseCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var courses []model.Course
	if err = cursor.All(context.Background(), &courses); err != nil {
		return nil, err
	}

	return courses, nil
}
