package sql

import (
	"fmt"
	"os"
)

func CreateAuthorsTable() string {
	schema := os.Getenv("DATABASE_SCHEMA")
	fmt.Println("schema", schema)

	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.authors(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username text NOT NULL,
	password bytea NOT NULL,
	email text,
	created timestamp without time zone NOT NULL,
	favorite_haikus integer[] NOT NULL,
	CONSTRAINT email_unique UNIQUE (email),
	CONSTRAINT username_unique UNIQUE (username)
)`, schema)
}

func CreateHaikusTable() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.haikus(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	text text NOT NULL,
	tags text[] NOT NULL,
	rating integer NOT NULL,
	created timestamp without time zone NOT NULL,
	authorid integer references authors("ID")
)`, schema)
}

func CreateVotesTable() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.votes(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	upvoted boolean NOT NULL,
	voted_timestamp timestamp without time zone NOT NULL,
	authorid integer references authors("ID"),
	haikuid integer references haikus("ID"),
	UNIQUE(authorid, haikuid)
)`, schema)
}

func InsertHaiku() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
INSERT INTO %s.haikus
VALUES (default, $1, $2, $3, now()::timestamp, $4)
RETURNING *`, schema)
}

func ListAllHaikus() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
SELECT * FROM %s.haikus
ORDER BY "ID" ASC
LIMIT ($1) OFFSET ($2)`, schema)
}

func GetHaikuById() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
SELECT * FROM %s.haikus
WHERE "ID"=$1`, schema)
}

func InsertAuthor() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
INSERT INTO %s.authors
VALUES (default, $1, $2, $3, now()::timestamp, '{}')`, schema)
}

func GetAuthorByUsername() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
SELECT * FROM %s.authors
WHERE "username"=$1`, schema)
}

/*
Maybe refactor this? This "implicitly" checks that the Haiku
belongs to the AuthorID from the Authorization header.

Without the 'AND "authorid"=$2', any user that is authenticated
will be able to delete ANY Haiku - adding the authorid clause here
shortcuts around that vulnerability (because the authorid passed here will be
the same username passed in the Authorization header, no Author will be able to
delete any Haikus but their own), but I think I maybe need to check that specifically
before executing this SQL.
*/
func DeleteHaikuById() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
DELETE FROM ONLY %s.haikus
WHERE "ID"=$1 AND "authorid"=$2`, schema)
}

func GetVoteByHaikuAndAuthor() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
SELECT * FROM %s.votes
WHERE "authorid"=$1 AND "haikuid"=$2`, schema)
}

/*
Insert vote using haikuid and authorid if a record with a combination
of those fields does not exist. If one exists, update the vote direction
and the voted_timestamp.
*/
func UpsertVote() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
INSERT INTO %s.votes
VALUES (default, $1, now()::timestamp, $2, $3)
ON CONFLICT (authorid, haikuid)
WHERE (upvoted != $1)
DO UPDATE SET
	upvoted = EXCLUDED.upvoted
	voted_timestamp = EXCLUDED.voted_timestamp`, schema)
}

func DeleteVoteById() string {
	schema := os.Getenv("DATABASE_SCHEMA")

	return fmt.Sprintf(`
DELETE FROM ONLY %s.votes
WHERE "ID"=$1 AND "authorid"=$2`, schema)
}

func DropAllTestTables() string {
	return "DROP TABLE IF EXISTS test.authors, test.haikus, test.votes CASCADE"
}
