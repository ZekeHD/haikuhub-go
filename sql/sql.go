package sql

func CreateAuthorsTable() string {
	return `
CREATE TABLE IF NOT EXISTS authors(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username text NOT NULL,
	password bytea NOT NULL,
	email text,
	created timestamp without time zone NOT NULL,
	favorite_haikus integer[] NOT NULL,
	CONSTRAINT email_unique UNIQUE (email),
	CONSTRAINT username_unique UNIQUE (username)
)`
}

func CreateHaikusTable() string {
	return `
CREATE TABLE IF NOT EXISTS haikus(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	text text NOT NULL,
	tags text NOT NULL,
	rating integer NOT NULL,
	created timestamp without time zone NOT NULL,
	authorid integer references authors("ID")
)`
}

func CreateVotesTable() string {
	return `
CREATE TABLE IF NOT EXISTS votes(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	upvoted boolean NOT NULL,
	voted_timestamp timestamp without time zone NOT NULL,
	authorid integer references authors("ID"),
	haikuid integer references haikus("ID"),
	UNIQUE(authorid, haikuid)
)`
}

func InsertHaiku() string {
	return `
INSERT INTO public.haikus
VALUES (default, $1, $2, $3, now()::timestamp, $4)
RETURNING *`
}

func ListAllHaikus() string {
	return `
SELECT * FROM public.haikus
ORDER BY "ID" ASC
LIMIT ($1) OFFSET ($2)`
}

func GetHaikuById() string {
	return `
SELECT * FROM public.haikus
WHERE "ID"=$1`
}

func InsertAuthor() string {
	return `
INSERT INTO public.authors
VALUES (default, $1, $2, $3, now()::timestamp, '{}')`
}

func GetAuthorByUsername() string {
	return `
SELECT * FROM public.authors
WHERE "username"=$1`
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
	return `
DELETE FROM ONLY public.haikus
WHERE "ID"=$1 AND "authorid"=$2`
}

func GetVoteByHaikuAndAuthor() string {
	return `
SELECT * FROM public.votes
WHERE "authorid"=$1 AND "haikuid"=$2`
}

/*
Insert vote using haikuid and authorid if a record with a combination
of those fields does not exist. If one exists, update the vote direction
and the voted_timestamp.
*/
func UpsertVote() string {
	return `
INSERT INTO votes
VALUES (default, $1, now()::timestamp, $2, $3)
ON CONFLICT (authorid, haikuid)
WHERE (upvoted != $1)
DO UPDATE SET
	upvoted = EXCLUDED.upvoted
	voted_timestamp = EXCLUDED.voted_timestamp`
}

func DeleteVoteById() string {
	return `
DELETE FROM ONLY public.votes
WHERE "ID"=$1 AND "authorid"=$2`
}
