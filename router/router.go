package router

import (

	"github.com/phcarneirobc/free-learn/db"
	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/handlers"
	"github.com/phcarneirobc/free-learn/model"
	"net/http"
	"github.com/phcarneirobc/free-learn/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HealthRoute(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "api is working!!",
	})
}


func AddCourseToUserHandler(c *gin.Context) {
    // Obter o ID do usuário a partir dos parâmetros da URL
    userID := c.Param("id")

    // Converter o ID do usuário para o tipo primitive.ObjectID
    userIDObj, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Obter o ID do curso a partir dos dados do corpo da solicitação
    var courseID struct {
        CourseID string `json:"course_id"`
    }
    if err := c.ShouldBindJSON(&courseID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
        return
    }

    // Converter o ID do curso para o tipo primitive.ObjectID
    courseIDObj, err := primitive.ObjectIDFromHex(courseID.CourseID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
        return
    }

    // Adicionar o curso ao usuário usando a função AddCourseToUser da package handlers
    if err := handlers.AddCourseToUser(userIDObj, courseIDObj); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Course added to user successfully"})
}




func Start(port string) {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.GET("/ping", HealthRoute)

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
	pg.Use(CORSMiddleware())
	pg.Use(auth.AuthenticateToken) 
	pg.POST("/post", PostCourse)
	pg.GET("/get", GetAllCourses)
	pg.GET("/get/:id", GetCourseByID)
	pg.PUT("/update/:id", UpdateCourseValue)
	pg.DELETE("/delete/:id", DeleteCourse)
	pg.POST("/add-course-to-user/:id", AddCourseToUserHandler)

	err := r.Run(port)
	if err != nil {
		panic(err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().
			Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
			)
		c.Writer.Header().
			Set(
				"Access-Control-Allow-Methods",
				"POST, OPTIONS, GET, PUT,DELETE",
			)
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
