package models

import (
	//        "github.com/coopernurse/gorp"
	//	"github.com/robfig/revel"
	"strings"
	"time"
	"database/sql"
)

type QdbEntry struct {
	QuoteId int
	Quote   string
	Created int64
	Rating  int
	Author  string
}

type PageState struct {
	Search string
	Page   int
	Size   int
}

type TagEntry struct {
	TagId   int
	QuoteId int
	Tag     string
}

type QdbView struct {
	QuoteId int
	Quote   string
	Created int64
	Rating  int
	Author  string

	Tags    sql.NullString
}

func (q *QdbView) Time() string {
	return time.Unix(q.Created, 0).Format(time.UnixDate)
}

func (q *QdbView) Clip() string {
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
