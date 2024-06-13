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
	"github.com/gin-gonic/gin"
	"net/http"
)

func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := registerUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func registerUser(user models.User) (*mongo.InsertOneResult, error) {
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

	userToInsert := models.User{
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

func getUserByEmail(email string) (*models.User, error) {
	email = strings.ToLower(email)
	var user models.User
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

func Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := loginUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func loginUser(user models.User) (LoginResponse, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)

	email := strings.ToLower(user.Email)
	var result models.User
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

func AddCourseToUser(c *gin.Context) {
	userID := c.Param("id")

	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var courseID struct {
		CourseID string `json:"course_id"`
	}
	if err := c.ShouldBindJSON(&courseID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	courseIDObj, err := primitive.ObjectIDFromHex(courseID.CourseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := addCourseToUser(userIDObj, courseIDObj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course added to user successfully"})
}

func addCourseToUser(userID primitive.ObjectID, courseID primitive.ObjectID) error {
	userCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	filter := bson.M{"_id": userID}
	update := bson.M{"$push": bson.M{"cursos": courseID}}
	_, err := userCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}


