package db

import (
	"os"
)

const CourseCollection = "course"
const UserCollection = "users"

func getDbConnectionString() string {
	str := os.Getenv("DB_CONNECTION_STRING")
	if str == "" {
		str = "mongodb://localhost:27017"
	}

	return str
}

func getDbName() string {
	str := os.Getenv("DBNAME")
	if str == "" {
		str = "db"
	}

	return str
}
