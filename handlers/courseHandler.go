package handlers

import (
	"context"
	"errors"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/db"
	models "github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SearchCourses(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}
	courses, err := searchCoursesByQuery(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func searchCoursesByQuery(query string) ([]models.Course, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var courses []models.Course
	if err = cursor.All(context.Background(), &courses); err != nil {
		return nil, err
	}
	return courses, nil
}

func PostCourse(c *gin.Context) {
	var reading models.Course
	if err := c.BindJSON(&reading); err != nil {
		c.JSON(400, gin.H{"message": "Bad Request"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	object, err := postCourse(reading, userID.(primitive.ObjectID))
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, object)
}

func postCourse(read models.Course, creatorID primitive.ObjectID) (models.Course, error) {
	toInsert := models.Course{
		Id:          primitive.NewObjectID(),
		Date:        primitive.NewDateTimeFromTime(time.Now()),
		Name:        read.Name,
		Description: read.Description,
		Image:       read.Image,
		Link:        read.Link,
		Modules:     read.Modules,
		CreatorID:   creatorID,
	}

	_, err := db.InsertOne(
		db.Instance.Client,
		db.Instance.Context,
		db.Instance.Dbname,
		db.CourseCollection,
		toInsert,
	)

	return toInsert, err
}

func GetCourseByID(c *gin.Context) {
	courseID := c.Param("id")

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid course ID"})
		return
	}

	result, err := getCourseByID(courseObjectID)
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

func getCourseByID(id primitive.ObjectID) (*models.Course, error) {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var result models.Course
	err := collection.FindOne(db.Instance.Context, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func GetAllCourses(c *gin.Context) {
	readings, err := getAllCourses()
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, readings)
}

func getAllCourses() ([]models.Course, error) {
	cur, err := readAllCourse(db.Instance.Client, db.Instance.Context, db.Instance.Dbname, db.CourseCollection)
	if err != nil {
		return nil, err
	}

	var results []models.Course
	for cur.Next(db.Instance.Context) {
		var elem models.Course
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func readAllCourse(client *mongo.Client, ctx context.Context, dataBase, col string) (*mongo.Cursor, error) {
	collection := client.Database(dataBase).Collection(col)
	cur, err := collection.Find(ctx, bson.M{})
	return cur, err
}

func UpdateCourseValue(c *gin.Context) {
	courseID := c.Param("id")

	var courseUpdate struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Link        string          `json:"link"`
		Image       string          `json:"image"`
		Modules     []models.Module `json:"modules"`
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

	err = updateCourseValue(courseObjectID, courseUpdate.Name, courseUpdate.Description, courseUpdate.Link, courseUpdate.Image, courseUpdate.Modules)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update course", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Course updated successfully"})
}

func updateCourseValue(courseID primitive.ObjectID, name, description, link, image string, modules []models.Module) error {
	updateData := bson.M{
		"name":        name,
		"description": description,
		"link":        link,
		"image":       image,
		"modules":     modules,
	}
	return updateCourseByID(courseID, updateData)
}

func updateCourseByID(id primitive.ObjectID, updateData bson.M) error {
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	_, err := collection.UpdateOne(db.Instance.Context, bson.M{"_id": id}, bson.M{"$set": updateData})
	return err
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

	err = deleteCourse(userID.(primitive.ObjectID), courseObjectID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete course", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Course deleted successfully"})
}

func deleteCourse(userID, courseID primitive.ObjectID) error {
	courseCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var course models.Course

	err := courseCollection.FindOne(context.Background(), bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("course not found")
		}
		return err
	}

	if course.CreatorID != userID {
		return errors.New("user is not the creator of the course")
	}

	_, err = courseCollection.DeleteOne(context.Background(), bson.M{"_id": courseID})
	return err
}

func RateCourse(c *gin.Context) {
	courseID := c.Param("id")
	var rating models.Rating
	if err := c.ShouldBindJSON(&rating); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if rating.UserID != userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID in request body does not match authenticated user"})
		return
	}

	if rating.Score < 1 || rating.Score > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Score must be between 1 and 5"})
		return
	}

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	if err := rateCourse(userID.(primitive.ObjectID), courseObjectID, rating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course rated successfully"})
}

func rateCourse(userID, courseID primitive.ObjectID, rating models.Rating) error {
	courseCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.CourseCollection)
	var course models.Course

	err := courseCollection.FindOne(context.Background(), bson.M{"_id": courseID}).Decode(&course)
	if err != nil {
		return err
	}

	if course.CreatorID == userID {
		return errors.New("professors cannot rate their own courses")
	}

	for _, existingRating := range course.Ratings {
		if existingRating.UserID == userID {
			return errors.New("user has already rated this course")
		}
	}

	course.Ratings = append(course.Ratings, rating)
	_, err = courseCollection.UpdateOne(context.Background(), bson.M{"_id": courseID}, bson.M{"$set": bson.M{"ratings": course.Ratings}})
	return err
}

func GetUserCourses(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	courses, err := getUserCourses(objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func getUserCourses(userID primitive.ObjectID) ([]models.Course, error) {
	userCollection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	var user models.User
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

	var courses []models.Course
	if err = cursor.All(context.Background(), &courses); err != nil {
		return nil, err
	}

	return courses, nil
}
