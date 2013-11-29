package models

import (
	"github.com/robfig/revel"
	"github.com/coopernurse/gorp"
	"strings"
	"time"
)

type QdbEntry struct {
	QuoteId int
	Quote   string
	Created int64
	Rating  int
	Author  string
}

func (q *QdbEntry) PreInsert (s gorp.SqlExecutor) error {
	q.Created = time.Now().Unix()
	q.Rating  = 0
	return nil
}

// the tag entry table model
type TagEntry struct {
	QuoteId int
	Tag     string
}

// a list of tags
type TagArray []string

// QdbView represents the QdbView SQL VIEW representation 
// of QdbEntry and TagEntry tables
type QdbView struct {
	QuoteId int
	Quote   string
	Created int64
	Rating  int
	Author  string

	Tags TagArray
}

func (quote *QdbView) Validate (v *revel.Validation) {
	v.Required(quote.Author)
	v.Required(quote.Quote)
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

type Pagination struct {
	Page int
	Size int

	Search string
	Tag string

	HasNext bool
	HasPrev bool

	Order string
	OrderDir string
}

func (p Pagination) NextPage () Pagination {
	t := p
	t.Page += 1
	return t
}

func (p Pagination) PrevPage () Pagination {
	t := p
	t.Page -= 1
	return t
}

func (p Pagination) OrderByDate () Pagination {
	t := p
	t.Order = ""
	return t
}

func (p Pagination) OrderByRating () Pagination {
	t := p
	t.Order = "rating"
	return t
}

func (p Pagination) OrderDesc () Pagination {
	t := p
	t.OrderDir = ""
	return t
}

func (p Pagination) OrderAsc () Pagination {
	t := p
	t.OrderDir = "asc"
	return t
}

type DateRange struct {
	Lower, Upper int64
}

