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

func DeleteHaikuById() string {
	return `
DELETE FROM ONLY public.haikus
WHERE "ID"=$1 AND "authorid"=$2`
}
