package controllers

import (
	"github.com/PacketFire/goqdb/app/models"
	"time"
)

var DEFAULT_SIZE = 5

type Core struct {
	GorpController
}

func loadEntries (result []interface{}) []*models.QdbView {

	var entries []*models.QdbView

	for _, r := range result {
		entries = append(entries, r.(*models.QdbView))
	}

	return entries
}

func (c *Core) insertEntry (entry *models.QdbEntry) error {

	entry.Created = time.Now().Unix()
	entry.Rating = 0

	c.Session["author"] = entry.Author

	return c.Txn.Insert(entry)
}

func (c *Core) getEntryById (id int) ([]*models.QdbView, error) {

	result, err := c.Txn.Select(models.QdbView{}, `
		SELECT 
			QdbEntry.*, G.Tags 
		FROM 
			QdbEntry
		LEFT JOIN
			(
				SELECT 
					TagEntry.QuoteId,
					GROUP_CONCAT(TagEntry.Tag, ',') AS Tags
				FROM 
					TagEntry
				GROUP BY 
					TagEntry.QuoteId
			) AS G
		ON
			G.QuoteId = QdbEntry.QuoteId
		WHERE
			QdbEntry.QuoteId = ?
		LIMIT 1
		`, id)

	return loadEntries(result), err
}

func (c *Core) getEntries (page, size int, search string) ([]*models.QdbView, error) {

	if size == 0 {
		size = DEFAULT_SIZE
	}

	var lower int
	if size > 0 {
		lower = size * (page - 1)
	} else {
		lower = 0
	}

	var result []interface{}
	var err error

	if search == "" {
		result, err = c.Txn.Select(models.QdbView{}, `
			SELECT 
				QdbEntry.*, G.Tags 
			FROM 
				QdbEntry
			LEFT JOIN
				(
					SELECT 
						TagEntry.QuoteId,
						GROUP_CONCAT(TagEntry.Tag, ',') AS Tags
					FROM 
						TagEntry
					GROUP BY 
						TagEntry.QuoteId
				) AS G
			ON
				G.QuoteId = QdbEntry.QuoteId
			LIMIT 
				?, ?`,
			lower, size)
	} else {
		result, err = c.Txn.Select(models.QdbEntry{}, `
			SELECT 
				QdbEntry.*, G.Tags 
			FROM 
				QdbEntry
			LEFT JOIN
				(
					SELECT 
						TagEntry.QuoteId,
						GROUP_CONCAT(TagEntry.Tag, ',') AS Tags
					FROM 
						TagEntry
					GROUP BY 
						TagEntry.QuoteId
				) AS G
			ON
				G.QuoteId = QdbEntry.QuoteId
			WHERE
				LOWER(QdbEntry.Quote) LIKE ?
			LIMIT
				?, ?`,
			"%"+search+"%", lower, size)
	}

	return loadEntries(result), err
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

	var changes int64
	changes, err = c.Txn.SelectInt(`SELECT CHANGES()`)

	return changes, err
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

	var changes int64
	changes, err = c.Txn.SelectInt(`SELECT CHANGES()`)

	return changes, err
}

