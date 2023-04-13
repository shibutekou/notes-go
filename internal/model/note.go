package model

import (
	"time"
)

type Note struct {
	Name      string
	Text      string
	Tag       string
	Author    string
	CreatedAt time.Time
	ID        int32
}

/*
{"text", text},
		{"tag", tag},
		{"created_at", time.Now()},
		{"author", author.Username},
		{"id", r.counter},
*/
