package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/handlers"
	"github.com/phcarneirobc/free-learn/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostCourse(c *gin.Context) {
	var reading model.Course
	if err := c.BindJSON(&reading); err != nil {
		c.JSON(400, gin.H{"message": "Bad Request"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	object, err := handlers.PostCourse(reading, userID.(primitive.ObjectID))
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, object)
}

func GetAllCourses(c *gin.Context) {
	readings, err := handlers.GetAllCourses()
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, readings)
}

func GetCourseByID(c *gin.Context) {
	courseID := c.Param("id")

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid course ID"})
		return
	}

	result, err := handlers.GetCourseByID(courseObjectID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get course by ID", "details": err.Error()})
		return
	}

	if result == nil {
		c.JSON(404, gin.H{"error": "Course not found"})
		return
	}

	c.JSON(200, result)
}

func UpdateCourseValue(c *gin.Context) {
	courseID := c.Param("id")

	var courseUpdate struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Link        string         `json:"link"`
		Image       string         `json:"image"`
		Modules     []model.Module `json:"modules"`
	}

	if err := c.BindJSON(&courseUpdate); err != nil {
		c.JSON(400, gin.H{"error": "Failed to bind JSON data"})
		return
	}

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid course ID"})
		return
	}

	err = handlers.UpdateCourseValue(courseObjectID, courseUpdate.Name, courseUpdate.Description, courseUpdate.Link, courseUpdate.Image, courseUpdate.Modules)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update course", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Course updated successfully"})
}

func DeleteCourse(c *gin.Context) {
	courseID := c.Param("id")

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid course ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	err = handlers.DeleteCourse(userID.(primitive.ObjectID), courseObjectID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete course", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Course deleted successfully"})
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

	if err := handlers.AddCourseToUser(userIDObj, courseIDObj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course added to user successfully"})
}
