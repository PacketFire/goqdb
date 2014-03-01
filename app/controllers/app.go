package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"
	"time"
)

type App struct {
	Base
}

var appDefaults = models.Args{
	To: models.UnixEpoch(time.Now().Unix()),
	Sort: models.SORT_DATE,
	Page: 1,
	Size: models.SIZE_DEFAULT,
}

func (c *App) Index (arg models.Args) revel.Result {

	var entries []models.Quote

	c.RenderArgs["hasNext"] = false
	c.RenderArgs["hasPrev"] = false

	if arg.Id > 0 {
		entries = c.getEntry(arg.Id)
	} else {

		argMerged := arg.Merge(appDefaults)

		count := c.getTotal(argMerged)

		if count == 0 {
			return c.Render(arg)
		}

		entries = c.getEntries(argMerged)

		offset  := argMerged.Size * (argMerged.Page - 1)
		hasNext := int64(offset + argMerged.Size) < count
		hasPrev := offset > 0

		c.RenderArgs["hasNext"] = hasNext
		c.RenderArgs["hasPrev"] = hasPrev
	}

	if author, ok := c.Session["author"]; ok {
		c.Flash.Data["quote.Author"] = author
	}

	return c.Render(arg, entries)
}

func (c *App) Quote (id int) revel.Result {
	var quote string
	if e := c.getEntry(id); len(e) != 0 {
		quote = e[0].Quote
	}
	return Utf8Result(quote)
}

func (c *App) AdvSearch () revel.Result {
	return c.Render()
}

func (c *App) Random () revel.Result {
	arg := models.Args{Sort: models.SORT_RANDOM, Size:1}

	var id int
	if e := c.getEntries(arg); len(e) != 0 {
		id = e[0].QuoteId
	}
	return c.Redirect(routes.App.Index(models.Args{Id: id}))
}

func (c *App) Insert (quote models.Quote, arg models.Args) revel.Result {
	c.Session["author"] = quote.Author
	c.insertQuote(&quote)
	return c.Redirect(routes.App.Index(arg))
}


func (c *App) Vote (voteId int, voteType string, arg models.Args) revel.Result {
	c.vote(voteId, voteType)
	return c.Redirect(routes.App.Index(arg))
}
