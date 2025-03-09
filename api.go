package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"haikuhub.net/haikuhubapi/authors"
	"haikuhub.net/haikuhubapi/db"
	"haikuhub.net/haikuhubapi/haikus"
	"haikuhub.net/haikuhubapi/types"
	"haikuhub.net/haikuhubapi/votes"
)

func main() {
	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Fatal("Error loading env file")
	}

	db.InitializeTables()

	r := gin.Default()
	r.HandleMethodNotAllowed = true

	r.PUT("/haiku", haikus.PutHaiku)
	r.GET("/haiku/:id", haikus.GetHaikuById)
	r.POST("/allHaikus", haikus.ListAllHaikus)
	r.DELETE("/haiku/:id", haikus.DeleteHaikuById)

	r.PUT("/author", authors.RegisterAuthor)

	r.POST("/vote", votes.PostVote)

	r.GET("/haiku", func(c *gin.Context) {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "haiku id parameter required: GET https://haikuhub.net/haikus/121",
		})
	})

	r.DELETE("/haiku", func(c *gin.Context) {
		c.JSON(types.HTTP_BAD, gin.H{
			"error": "haiku id parameter required: DELETE https://haikuhub.net/haikus/121",
		})
	})

	r.Run()
}
