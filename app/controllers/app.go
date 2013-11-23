package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/routes"
	"github.com/PacketFire/goqdb/app/models"
	"fmt"
	"strings"
	"net/http"
)

type App struct {
	Core
}

func (c *App) Index (page models.PageState) revel.Result {

	var savedAuthor string

	if author, ok := c.Session["author"]; ok {
		savedAuthor = author
	}

	if page.Page == 0 {
		page.Page = 1
	}

	page.Search = strings.TrimSpace(page.Search)

	entries, err := c.getEntries(page.Page, page.Size, page.Tag, page.Search)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	}

	nextPage := page.Page + 1
	prevPage := page.Page - 1


	hasPrevPage := page.Page > 1
	hasNextPage := len(entries) == page.Size

	return c.Render(entries, savedAuthor, page, hasPrevPage, prevPage, hasNextPage, nextPage)
}

func (c *App) Post (entry models.QdbEntry, tags string, page models.PageState) revel.Result {

	var view models.QdbView

	c.Validation.Required(entry.Quote)
	c.Validation.Required(entry.Author)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.App.Index(page))
	}

	view.Quote = entry.Quote
	view.Author = entry.Author

	if tags != "" {
		view.Tags = strings.Split(tags, ",")

		for _, t := range view.Tags {
			t = strings.TrimSpace(t)
		}
	}

	err := c.insertView(&view)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.Redirect(routes.App.Index(page))
	}

	c.Session["author"] = entry.Author

	return c.Redirect(routes.App.Index(page))
}

func (c *App) One (id int) revel.Result {

	var quote string
	entries, err := c.getEntryById(id);

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	} else {
		if len(entries) == 0 {
			c.Flash.Error(fmt.Sprintf("No such id: %d", id))
		} else {
			quote = entries[0].Quote
		}
	}

	return Utf8Result(quote)
}

func (c *App) UpVote (id int, page models.PageState) revel.Result {

	_, err := c.upVote(id)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	}

	return c.Redirect(routes.App.Index(page))
}

func (c *App) DownVote (id int, page models.PageState) revel.Result {

	_, err := c.downVote(id)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	}

	return c.Redirect(routes.App.Index(page))
}

