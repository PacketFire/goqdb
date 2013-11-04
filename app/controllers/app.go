package controllers

import (
	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"
	"github.com/robfig/revel"
	"strings"
)

type App struct {
	GorpController
}

/*
func (c App) Index() revel.Result {
	return c.Render()
}
*/

func (c App) Index(search string, size, page int) revel.Result {
	if page == 0 {
		page = 1
	}

	if size <= 0 {
		size = 6
	}

	nextPage := page + 1
	prevPage := page - 1
	search = strings.TrimSpace(search)

	var entries []*models.QdbEntry

	if search == "" {
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry ORDER BY QuoteId DESC LIMIT ?, ?`, (page-1)*size, size))
	} else {
		search = strings.ToLower(search)
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry WHERE LOWER(Quote) LIKE ? ORDER BY QuoteId LIMIT ?, ?`, "%"+search+"%", (page-1)*size, size))

	}

	hasPrevPage := page > 1
	hasNextPage := len(entries) == size
	if hasNextPage {
		entries = entries[:len(entries)-1]
	}

	return c.Render(entries, search, size, page, hasPrevPage, prevPage, hasNextPage, nextPage)
}

func loadEntries(results []interface{}, err error) []*models.QdbEntry {
	if err != nil {
		panic(err)
	}

	var entries []*models.QdbEntry

	for _, r := range results {
		entries = append(entries, r.(*models.QdbEntry))
	}

	return entries
}

func (c App) Post() revel.Result {
	var quote models.QdbEntry
	c.Params.Bind(&quote.Quote, "quote")
	c.Txn.Insert(&quote)
	return c.Redirect(routes.App.Index("", 0, 0))
}

func (c App) One(id int) revel.Result {
	var entries []*models.QdbEntry
	entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry WHERE QuoteId = ? ORDER BY QuoteId DESC LIMIT 1`, id))
	if len(entries) == 0 {
		c.Flash.Error("no such id")
	}
	quote := entries[0]
	return c.Render(quote)
}
