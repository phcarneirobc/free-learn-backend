package router

import (

	"github.com/phcarneirobc/free-learn/db"
	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/handlers"
	"github.com/phcarneirobc/free-learn/model"
	"net/http"
	"github.com/phcarneirobc/free-learn/auth"
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
	pg.Use(auth.AuthenticateToken) // Add this line
	pg.POST("/post", PostCourse)
	pg.GET("/get", GetAllCourses)
	pg.GET("/get/:id", GetCourseByID)
	pg.PUT("/update/:id", UpdateCourseValue)
	pg.DELETE("/delete/:id", DeleteCourse)


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
