package db

import (
	ctx "context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/jackc/pgx/v5/pgxpool"
	"haikuhub.net/haikuhubapi/sql"
)

func getConnectionPool() *pgxpool.Pool {
	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Fatal("Error loading env file")
	}

	databaseUrl := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx.Background(), databaseUrl)
	if err != nil {
		log.Fatal("unable to get Postgres pool", err.Error())
	}

	return pool
}

func InitializeTables() {
	_, err := Pool.Exec(ctx.Background(), sql.CreateAuthorsTable())
	if err != nil {
		log.Fatal("Unable to create 'authors' table!", err.Error())
	}

	_, err = Pool.Exec(ctx.Background(), sql.CreateHaikusTable())
	if err != nil {
		log.Fatal("Unable to create 'haikus' table!", err.Error())
	}

	_, err = Pool.Exec(ctx.Background(), sql.CreateVotesTable())
	if err != nil {
		log.Fatal("Unable to create 'votes' table!", err.Error())
	}
}

var Pool = getConnectionPool()
