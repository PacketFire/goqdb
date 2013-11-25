package controllers

import (
//	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"
	"github.com/coopernurse/gorp"
	"time"
	"errors"
	"strings"
)

type QdbTypeConverter struct {}

func (me QdbTypeConverter) ToDb (val interface{}) (interface{}, error) {
	return val, nil
}

func (me QdbTypeConverter) FromDb (target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
		case *models.TagList:
			binder := func (holder, target interface{}) error {
				s, ok := holder.(*string)

				if !ok {
					return errors.New("FromDb: Unable to convert holder")
				}

				sl, ok := target.(*models.TagList)

				if !ok {
					return errors.New("FromDb: Unable to convert target")
				}

				if *s == "" {
					*sl = make(models.TagList, 0)
				} else {
					*sl = strings.Split(*s, ",")
				}
				return nil
			}
			return gorp.CustomScanner{new(string), target, binder}, true
	}

	return gorp.CustomScanner{}, false
}

type Core struct {
	GorpController
}

func (c *Core) insertView (view *models.QdbView) error {

	view.Created = time.Now().Unix()
	view.Rating = 0

	entry := models.QdbEntry{
		Quote:   view.Quote,
		Author:  view.Author,
		Created: view.Created,
		Rating:  view.Rating}

	err := c.Txn.Insert(&entry)

	if err != nil {
		return err
	}

	view.QuoteId = entry.QuoteId

	for _, s := range view.Tags {
		s = strings.TrimSpace(s)
		if s != "" {
			c.Txn.Insert(&models.TagEntry{QuoteId: entry.QuoteId, Tag: s})
		}
	}

	return err
}

func (c *Core) getEntry (id int) (models.QdbView, error) {

	var entries []models.QdbView
	_, err := c.Txn.Select(&entries, `
		SELECT 
			*
		FROM 
			QdbView
		WHERE
			QuoteId = ?
		LIMIT 1
		`, id,
	)

	if len(entries) == 0 {
		return models.QdbView{}, err
	}

	return entries[0], err
}

func (c *Core) getEntries (state models.PageState, r models.DateRange) ([]models.QdbView, error) {

	params := make(map[string]interface{})

	var entries []models.QdbView
	var err error

	query := `
		SELECT
			*
		FROM 
			QdbView`

	if state.Search != "" {
		query += `
		WHERE 
			Quote LIKE :search`

		params["search"] = "%"+state.Search+"%"
	}

	if state.Tag != "" {
		if state.Search != "" {
			query += `
		AND`
		} else {
			query += `
		WHERE`
		}

		query += `
			QuoteId IN (
				SELECT 
					TagEntry.QuoteId
				FROM
					TagEntry
				WHERE
					TagEntry.Tag = :tag
			)`

		params["tag"] = state.Tag
	}

	if r.Lower != 0 && r.Upper != 0 {
		if state.Search != "" || state.Tag != "" {
			query += `
		AND`
		} else {
			query += `
		WHERE`
		}

		query += `
			Created BETWEEN :dateLower AND :dateUpper`
		params["dateLower"] = r.Lower
		params["dateUpper"] = r.Upper
	}

	var lower int

	if state.Size > 0 {
		lower = state.Size * (state.Page - 1)
	} else {
		lower = 0
	}
	query += `
		LIMIT :limitLower, :size`

	params["limitLower"] = lower
	params["size"]  = state.Size

	_, err = c.Txn.Select(&entries, query, params)
	return entries, err
}

func (c *Core) upVote (id int) (int64, error) {

	_, err := c.Txn.Exec(
		`UPDATE
			QdbEntry
		SET
			Rating = Rating + 1
		WHERE
			QuoteId = ?`,
		id)

	if err != nil {
		return 0, err
	}

	return c.Txn.SelectInt(`SELECT CHANGES()`)
}

func (c *Core) downVote (id int) (int64, error) {

	_, err := c.Txn.Exec(
		`UPDATE
			QdbEntry
		SET
			Rating = Rating - 1
		WHERE
			QuoteId = ?`,
		id)

	if err != nil {
		return 0, err
	}

	return c.Txn.SelectInt(`SELECT CHANGES()`)
}

