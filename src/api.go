package main

import (
	ctx "context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	haikusSQL "haikuhub.net/haikuhubapi/src/sql"
	"haikuhub.net/haikuhubapi/src/types"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const OK = http.StatusOK
const BAD = http.StatusBadRequest
const INTERNAL = http.StatusInternalServerError

type HaikuPUT struct {
	Text string `json:"text" binding:"required"`
	Tags string `json:"tags"`
}

func main() {
	envLoadErr := godotenv.Load("./.env")
	if envLoadErr != nil {
		log.Fatal("Error loading .env file")

		return
	}

	r := gin.Default()

	r.GET("/allHaikus", listAllHaikus)
	r.GET("/haiku/:id", getHaikuById)
	r.PUT("/haiku", putHaiku)
	r.DELETE("/haiku/:id", deleteHaikuById)

	r.Run()
}

func listAllHaikus(c *gin.Context) {
	sql := haikusSQL.ListAllHaikusSQL()

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	rows, err := conn.Query(ctx.Background(), sql)
	if err != nil {
		fmt.Println("list all failed", err)
		return
	}

	haikus, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Haiku])
	if err != nil {
		c.JSON(INTERNAL, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(OK, gin.H{
		"haikus": haikus,
	})
}

func getHaikuById(c *gin.Context) {
	haikuId := c.Param("id")

	sql := haikusSQL.GetHaikuById()

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	row := conn.QueryRow(ctx.Background(), sql, haikuId)

	haiku := types.Haiku{}
	err := row.Scan(
		&haiku.ID,
		&haiku.Text,
		&haiku.Tags,
		&haiku.Rating,
		&haiku.AuthorID,
		&haiku.Created,
	)

	if err != nil {
		errMessage := err.Error()

		if err.Error() == "no rows in result set" {
			errMessage = "haiku not found!"
		}

		c.JSON(BAD, gin.H{
			"error": errMessage,
		})

		return
	}

	c.JSON(OK, gin.H{
		"haiku": haiku,
	})
}

func putHaiku(c *gin.Context) {
	var body HaikuPUT

	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		c.JSON(BAD, gin.H{"response": err})

		return
	}

	sql := haikusSQL.InsertHaikuSQL()

	conn := getPostgresConn()
	defer conn.Close(ctx.Background())

	row := conn.QueryRow(ctx.Background(), sql, body.Text, body.Tags, 0, "the-authors-id")
	insertedHaiku := types.Haiku{}

	insertErr := row.Scan(
		&insertedHaiku.ID,
		&insertedHaiku.Text,
		&insertedHaiku.Tags,
		&insertedHaiku.Rating,
		&insertedHaiku.AuthorID,
		&insertedHaiku.Created,
	)

	if insertErr != nil {
		c.JSON(INTERNAL, gin.H{
			"error": insertErr.Error(),
		})

		return
	}

	c.JSON(OK, gin.H{
		"haiku": insertedHaiku,
	})
}

func deleteHaikuById(c *gin.Context) {
	haikuId := c.Param("id")

	deletedString := fmt.Sprintf("Haiku with ID '%s' deleted", haikuId)

	c.JSON(OK, gin.H{
		"message": deletedString,
	})
}

func getPostgresConn() *pgx.Conn {
	connectionStr := os.Getenv("POSTGRES_CONNECTION_STR")
	fmt.Println("connectionStr:", connectionStr)
	conn, err := pgx.Connect(ctx.Background(), connectionStr)
	if err != nil {
		fmt.Println("bad things happened!")
		fmt.Println(err)
	}

	return conn
}
