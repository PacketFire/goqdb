package models

import (
	//        "github.com/coopernurse/gorp"
	//        "github.com/robfig/revel"
	"time"
	"strings"
)

type QdbEntry struct {
	QuoteId int
	Quote   string
	Created int64
	Rating  int
	Author string
	//Tags string
}

func (q *QdbEntry) Clip() string {
	if len(q.Quote) > 256 {
		return q.Quote[:256] + "\r\n..."
	}

	if lines := strings.Split(q.Quote, "\n"); len(lines) > 20 {
		lines = lines[:20]
		lines = append(lines, "\r\n...")
		return strings.Join(lines, " ")
	}

	return q.Quote
}

func (q *QdbEntry) Time() string {
	return time.Unix(q.Created, 0).Format(time.UnixDate)
}

