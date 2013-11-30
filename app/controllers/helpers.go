package controllers

import (
	"github.com/robfig/revel"
	_"github.com/coopernurse/gorp"
	"github.com/PacketFire/goqdb/app/models"
	"strings"
	_"errors"
	"fmt"
)

// TODO: implement form input limit hints so that truncation is less silent

// helper functions

// TagArray insertion helper, also does truncation
func (c *GorpController) insertTagArray (parent int, tags models.TagArray) error {

	var t []interface{}

	for i := range tags {
		// skip entries that are too long or empty
		if n := len(tags[i]); n < INPUT_TAG_MAX && n > 0 {
			t = append(t,
				interface{}(strings.TrimSpace(tags[i])),
			)
		}

		if i > INPUT_TAG_LIST_MAX {
			break
		}
	}

	if len(t) == 0 {
		return nil
	}

	s := fmt.Sprintf("(%d, ?)", parent)
	s += strings.Repeat(", " + s, len(t) - 1)

	_, err := c.Txn.Exec(`INSERT OR IGNORE INTO TagEntry VALUES ` +  s, t...)
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

