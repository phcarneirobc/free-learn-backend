package auth

import (
    "os"
    "time"
    "errors"
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    
    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
    "github.com/phcarneirobc/free-learn/model"
  
)

var jwtKey = []byte(os.Getenv("JWT_SECRET")) // Get secret key from environment variable

const (
    ErrInvalidTokenSignature = "Invalid token signature"
    ErrCouldNotParseToken    = "Could not parse token"
)

type Claims struct {
    Username string `json:"username"`
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

    token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("Unexpected signing method")
        }
        return jwtKey, nil
    })

    if err != nil {
        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    if !token.Valid {
        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    c.Next()
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) // Reduce cost to 12
    return string(bytes), err
}

// CheckPasswordHash checks a password against a hashed password
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// GenerateToken generates a JWT token for a given user
func GenerateToken(user model.User) (string, error) {
   

    expirationTime := time.Now().Add(24 * time.Hour)

    claims := &Claims{
        Username: user.Name,
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

// ValidateToken validates the JWT token
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

    return true, claims.Username
}