package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/auth"
	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/handlers"
	"github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HealthRoute(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "api is working!!",
	})
}

func Start(port string) {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.GET("/ping", HealthRoute)
	r.GET("/get",  GetAllCourses)
	r.POST("/register", func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := handlers.Register(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	r.POST("/login", func(c *gin.Context) {
		var user model.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := handlers.Login(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	pg := r.Group("/courses")
	pg.Use(auth.AuthenticateToken)
	pg.POST("/post", auth.RequireProfessor, PostCourse)
	pg.GET("/get/:id", auth.RequireProfessor, GetCourseByID)
	pg.PUT("/update/:id", auth.RequireProfessor, UpdateCourseValue)
	pg.DELETE("/delete/:id", auth.RequireProfessor, DeleteCourse)
	pg.POST("/add-course-to-user/:id", AddCourseToUser)
	pg.GET("/get-user-courses/:id", func(c *gin.Context) {
		userID := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		courses, err := GetUserCourses(objectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, courses)
	})

	err := r.Run(port)
	if err != nil {
		panic(err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func PrepareApp() error {
	return db.StartDB()
}

func AddCourseToUserHandler(c *gin.Context) {
	userID := c.Param("id")
	courseID := c.Param("courseId")

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	courseObjectID, err := primitive.ObjectIDFromHex(courseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	err = handlers.AddCourseToUser(userObjectID, courseObjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add course to user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course added to user successfully"})
}



