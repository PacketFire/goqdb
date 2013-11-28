package controllers

import (
	"github.com/robfig/revel"
	_"github.com/coopernurse/gorp"
	"github.com/PacketFire/goqdb/app/models"
	"strings"
	_"errors"
)

// TODO: implement form input limit hints so that truncation is less silent

// helper functions

// TagArray insertion helper, also does truncation
func (c *GorpController) insertTagArray (parent int, tags models.TagArray) error {

	for i := range tags {
		// delete entries that are too long
		if len(tags[i]) > INPUT_TAG_MAX {
			tags = append(tags[:i-1], tags[i+1:]...)
		} else {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	if len(tags) > INPUT_TAG_LIST_MAX {
		tags = tags[:INPUT_TAG_LIST_MAX]
	}

	var entries []interface{}
	var values string

	for i := range tags {
		entries = append(entries, parent)
		entries = append(entries, tags[i])

		values += " (?, ?)"

		if i < len(tags) - 1 {
			values += ","
		}
	}

	_, err := c.Txn.Exec(`INSERT OR IGNORE INTO TagEntry VALUES ` +  values, entries...)
	return err
}

func (c *GorpController) insertView (q *models.QdbView) error {

	if len(q.Quote) > INPUT_QUOTE_MAX {
		q.Quote = q.Quote[:INPUT_QUOTE_MAX]
	}

	if len(q.Author) > INPUT_AUTHOR_MAX {
		q.Author = q.Author[:INPUT_AUTHOR_MAX]
	}

	e := models.QdbEntry{
		Quote: q.Quote,
		Author: q.Author,
	}

	err := c.Txn.Insert(&e)

	if err != nil {
		return err
	}

	q.QuoteId = e.QuoteId
	q.Created = e.Created
	q.Rating  = e.Rating

	return c.insertTagArray(q.QuoteId, q.Tags)
}

func (c *GorpController) upVote (id int) (bool, error) {
	_, err := c.Txn.Exec(`UPDATE QdbEntry SET Rating = Rating + 1 WHERE QuoteId = ?`, id)

	if err != nil {
		return false, err
	}

	changes, err := c.Txn.SelectInt(`SELECT CHANGES()`)

	return changes != 0, err
}

func (c *GorpController) downVote (id int) (bool, error) {
	_, err := c.Txn.Exec(`UPDATE QdbEntry SET Rating = Rating - 1 WHERE QuoteId = ?`, id)

	if err != nil {
		return false, err
	}

	changes, err := c.Txn.SelectInt(`SELECT CHANGES()`)

	return changes != 0, err
}

type Utf8Result string

func (u Utf8Result) Apply(req *revel.Request, resp *revel.Response) {
	resp.WriteHeader(resp.Status, "text/plain; charset=utf-8")
	resp.Out.Write([]byte(u))
}

