package types

import "time"

type Author struct {
	ID      string
	Created time.Time
}

type Haiku struct {
	ID       string
	Text     string
	Created  time.Time
	Rating   int16
	Tags     string
	AuthorID string
}
