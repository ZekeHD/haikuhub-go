package haikusSQL

func InsertHaikuSQL() string {
	return `
INSERT INTO public.haikus
VALUES (default, $1, $2, $3, $4, now())
RETURNING *`
}

func ListAllHaikusSQL() string {
	return `
SELECT * FROM public.haikus
ORDER BY "ID" ASC`
}

func GetHaikuById() string {
	return `
SELECT * FROM public.haikus
WHERE "ID"=$1`
}
