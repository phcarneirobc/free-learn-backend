package main

import (
	"fmt"

	"github.com/phcarneirobc/free-learn/router"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting Application...")
	godotenv.Load()
	err := router.PrepareApp()
	if err != nil {
		panic(err)
	}

	router.Start(":4430")
}
