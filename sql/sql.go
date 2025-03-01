package haikusSQL

func InsertHaiku() string {
	return `
INSERT INTO public.haikus
VALUES (default, $1, $2, $3, $4, now())
RETURNING *`
}

func ListAllHaikus() string {
	return `
SELECT * FROM public.haikus
ORDER BY "ID" ASC`
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
