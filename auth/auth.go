package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/phcarneirobc/free-learn/db"
	"github.com/phcarneirobc/free-learn/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

const (
	ErrInvalidTokenSignature = "Invalid token signature"
	ErrCouldNotParseToken    = "Could not parse token"
)

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func AuthenticateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
		return
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	token, err := jwt.ParseWithClaims(bearerToken[1], &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	var user models.User
	collection := db.Instance.Client.Database(db.Instance.Dbname).Collection(db.UserCollection)
	err = collection.FindOne(context.Background(), bson.M{"email": claims.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.Set("userID", user.Id)
	c.Set("userProfessor", user.Professor)
	c.Next()
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tknStr string) (bool, string) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false, ErrInvalidTokenSignature
		}
		return false, ErrCouldNotParseToken
	}

	if !tkn.Valid {
		return false, ErrInvalidTokenSignature
	}

	return true, claims.Email
}

func RequireProfessor(c *gin.Context) {
	userProfessor, exists := c.Get("userProfessor")
	if !exists || !userProfessor.(bool) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access forbidden: professors only"})
		return
	}
	c.Next()
}
