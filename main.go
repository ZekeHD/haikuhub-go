package main

import (
	"log"

	"github.com/joho/godotenv"

	"haikuhub.net/haikuhubapi/api"
	"haikuhub.net/haikuhubapi/db"
)

func main() {
	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Fatal("Error loading env file", envLoadErr.Error())
	}

	db.InitializeTables()

	r := api.GetRouter()

	r.Run()
}
