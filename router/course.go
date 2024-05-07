package router

import (

	"github.com/gin-gonic/gin"

	"github.com/phcarneirobc/free-learn/handlers"
	"github.com/phcarneirobc/free-learn/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostCourse(c *gin.Context) {
	var reading model.Course
	err := c.BindJSON(&reading)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Bad Request",
		})
		return
	}

	object, err := handlers.PostCourse(reading)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, object)
}

func GetAllCourses(c *gin.Context) {
	readings, err := handlers.GetAllCourses()
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
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

	// Get the machine by ID
	result, err := handlers.GetCourseByID(courseObjectID)
	if err != nil {
		c.JSON(
			500, gin.H{
				"error":   "Failed to get course by ID",
				"details": err.Error(),
			},
		)
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
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Link       string `json:"link"`
		Image string `json:"image"`
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

	// Update the machine
	err = handlers.UpdateCourseValue(
		courseObjectID,
		courseUpdate.Name,
		courseUpdate.Description,
		courseUpdate.Link,
		courseUpdate.Image,
	)
	if err != nil {
		c.JSON(
			500,
			gin.H{
				"error":   "Failed to update course",
				"details": err.Error(),
			},
		)
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

    // Delete the course
    err = handlers.DeleteCourse(courseObjectID) // Call the correct function here
    if err != nil {
        c.JSON(
            500,
            gin.H{
                "error":   "Failed to delete course",
                "details": err.Error(),
            },
        )
        return
    }

    c.JSON(200, gin.H{"message": "Course deleted successfully"})
}

