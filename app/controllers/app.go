package controllers

import (
	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"
	"github.com/robfig/revel"
	"strings"
	"time"
)

type App struct {
	GorpController
}

func (c App) Index() revel.Result {
	var search string
	var size, page int

	c.Params.Bind(&search, "search")
	c.Params.Bind(&size, "size")
	c.Params.Bind(&page, "page")

	if page == 0 {
		page = 1
	}

	if size <= 0 {
		size = 5
	}

	// for pagination
	size += 1
	nextPage := page + 1
	prevPage := page - 1

	search = strings.TrimSpace(search)

	var entries []*models.QdbEntry

	if search == "" {
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry ORDER BY QuoteId DESC LIMIT ?, ?`, (page-1)*(size-1), size))
	} else {
		search = strings.ToLower(search)
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry WHERE LOWER(Quote) LIKE ? ORDER BY QuoteId DESC LIMIT ?, ?`, "%"+search+"%", (page-1)*(size), size))

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

	quote.Created = time.Now().Unix()
	quote.Rating = 0

	c.Params.Bind(&quote.Author, "author")
	c.Params.Bind(&quote.Quote, "quote")

	quote.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.App.Index())
	}
	c.Txn.Insert(&quote)
	return c.Redirect(routes.App.Index())
}

func (c App) RatingUp (id int) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating + 1 WHERE QuoteId = ?", id)

	if err != nil {
	}

	return c.Redirect(routes.App.Index())
}


func (c App) RatingDown (id int) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating - 1 WHERE QuoteId = ?", id)

	if err != nil {
	}

	return c.Redirect(routes.App.Index())
}

func (c App) One(id int) revel.Result {
	var entries []*models.QdbEntry
	entries = loadEntries(c.Txn.Select(models.QdbEntry{},
		`SELECT * FROM QdbEntry WHERE QuoteId = ? ORDER BY QuoteId DESC LIMIT 1`, id))
	if len(entries) == 0 {
		c.Flash.Error("no such id")
	}
	quote := entries[0]
	return c.RenderText(quote.Quote)
}
