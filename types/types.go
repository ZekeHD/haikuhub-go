package types

import (
	"net/http"
	"time"
)

type Author struct {
	ID       string
	Username string
	Password []byte
	Email    string
	Created  time.Time
}

type Haiku struct {
	ID       string
	Text     string
	Created  time.Time
	Rating   int16
	Tags     string
	AuthorID string
}

const HTTP_OK = http.StatusOK
const HTTP_BAD = http.StatusBadRequest
const HTTP_UNAUTHORIZED = http.StatusUnauthorized
const HTTP_INTERNAL = http.StatusInternalServerError
