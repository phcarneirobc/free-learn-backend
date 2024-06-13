package router

import (
	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/auth"
	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/handlers"
)

func Start(port string) {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "api is working!!",
		})
	})

	r.GET("/get", handlers.GetAllCourses)
	r.GET("/search", handlers.SearchCourses)
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	pg := r.Group("/courses")
	pg.Use(auth.AuthenticateToken)
	pg.POST("/post", auth.RequireProfessor, handlers.PostCourse)
	pg.GET("/get/:id", handlers.GetCourseByID)
	pg.PUT("/update/:id", auth.RequireProfessor, handlers.UpdateCourseValue)
	pg.DELETE("/delete/:id", auth.RequireProfessor, handlers.DeleteCourse)
	pg.POST("/add-course-to-user/:id", handlers.AddCourseToUser)
	pg.POST("/rate/:id", handlers.RateCourse)
    pg.GET("/get-user-courses/:id", handlers.GetUserCourses)
    
	err := r.Run(port)
	if err != nil {
		panic(err)
	}
   
}

func PrepareApp() error {
    return db.StartDB()
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
