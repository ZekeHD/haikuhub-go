package sql

func CreateAuthorsTable() string {
	return `
CREATE TABLE IF NOT EXISTS authors(
	"ID" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	username text NOT NULL,
	password bytea NOT NULL,
	email text NOT NULL,
	created timestamp without time zone NOT NULL,
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

func InsertHaiku() string {
	return `
INSERT INTO public.haikus
VALUES (default, $1, $2, $3, now(), $4)
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
VALUES (default, $1, $2, $3, now())`
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
