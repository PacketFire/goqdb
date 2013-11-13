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

func (c App) Index(page models.PageState) revel.Result {

	/*var search string
	var size, page int
	*/

	var savedAuthor string

	if author, ok := c.Session["author"]; ok {
		savedAuthor = author
	}

/*
	c.Params.Bind(&search, "search")
	c.Params.Bind(&size, "size")
	c.Params.Bind(&page, "page")
*/
	if page.Page == 0 {
		page.Page = 1
	}

	if page.Size <= 0 {
		page.Size = 5
	}

	// for pagination
	page.Size += 1
	nextPage := page.Page + 1
	prevPage := page.Page - 1

	search := strings.TrimSpace(page.Search)

	var entries []*models.QdbEntry

	if search == "" {
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry ORDER BY QuoteId DESC LIMIT ?, ?`, (page.Page-1)*(page.Size-1), page.Size))
	} else {
		search = strings.ToLower(search)
		entries = loadEntries(c.Txn.Select(models.QdbEntry{},
			`SELECT * FROM QdbEntry WHERE LOWER(Quote) LIKE ? ORDER BY QuoteId DESC LIMIT ?, ?`, "%"+search+"%", (page.Page-1)*(page.Size), page.Size))

	}

	hasPrevPage := page.Page > 1
	hasNextPage := len(entries) == page.Size
	if hasNextPage {
		entries = entries[:len(entries)-1]
	}

	return c.Render(entries, savedAuthor, page, hasPrevPage, prevPage, hasNextPage, nextPage)
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

func (c App) Post(entry models.QdbEntry, page models.PageState) revel.Result {

	entry.Created = time.Now().Unix()
	entry.Rating = 0

	c.Session["author"] = entry.Author

	entry.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.App.Index(page))
	}
	c.Txn.Insert(&entry)
	return c.Redirect(routes.App.Index(page))
}

func (c App) RatingUp(id int, page models.PageState) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating + 1 WHERE QuoteId = ?", id)

	if err != nil {
	}

	return c.Redirect(routes.App.Index(page))
}

func (c App) RatingDown(id int, page models.PageState) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating - 1 WHERE QuoteId = ?", id)

	if err != nil {
	}

	return c.Redirect(routes.App.Index(page))
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
