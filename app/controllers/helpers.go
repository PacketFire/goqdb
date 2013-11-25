package controllers

import (
	"github.com/robfig/revel"
	"github.com/coopernurse/gorp"
	"github.com/PacketFire/goqdb/app/models"
	"strings"
	"errors"
)

// TODO: implement form input limit hints so that truncation is less silent

// helper functions

// TagArray insertion helper, also does truncation
// TODO: tag entries must be unique (SQL unique key on [QuoteId,Tag] pair)
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

	var err error

	for i := range tags {
		err = c.Txn.Insert(&models.TagEntry{QuoteId: parent, Tag: tags[i]})

		if err != nil {
			return err
		}
	}
	return nil
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

type QdbTypeConverter struct {}

func (me QdbTypeConverter) ToDb (val interface{}) (interface{}, error) {
	return val, nil
}

func (me QdbTypeConverter) FromDb (target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {

		// split csv values from QdbView Tags column
		case *models.TagArray:
			binder := func (holder, target interface{}) error {
				s, ok := holder.(*string)

				if !ok {
					return errors.New("FromDb: Unable to convert holder")
				}

				sl, ok := target.(*models.TagArray)

				if !ok {
					return errors.New("FromDb: Unable to convert target")
				}

				if *s == "" {
					*sl = models.TagArray{}
				} else {
					*sl = strings.Split(*s, ",")
				}
				return nil
			}
			return gorp.CustomScanner{new(string), target, binder}, true
		}

	return gorp.CustomScanner{}, false
}
