package controllers

import (
	"github.com/PacketFire/goqdb/app/models"
//	"github.com/PacketFire/goqdb/app/routes"
	"github.com/robfig/revel"
	"net/http"
	"time"
)

/* Not sure wether to use Api or App here */
type Api struct {
	GorpController
}

func (c Api) Index () revel.Result {

	entries := loadEntries(c.Txn.Select(models.QdbEntry{}, `SELECT * FROM QdbEntry`))

	if len(entries) == 0 {
		c.Response.Status = http.StatusNoContent
		return c.RenderJson(nil)
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(entries)
}

func (c Api) One (id int) revel.Result {

	entries := loadEntries(c.Txn.Select(models.QdbEntry{}, `SELECT * FROM QdbEntry WHERE QuoteId = ?`, id))

	if len(entries) == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	return c.RenderJson(entries[0])
}

func (c *Api) Post (entry models.QdbEntry) revel.Result {

	entry.Created = time.Now().Unix()
	entry.Rating = 0

	c.Validation.Required(entry.Quote)
	c.Validation.Required(entry.Author)

	if c.Validation.HasErrors() {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	err := c.Txn.Insert(&entry)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(nil)
	}

	if err != nil {
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJson(entry)
}

func (c Api) RatingUp (id int) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating + 1 WHERE QuoteId = ?", id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(nil)
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(nil)
}

func (c Api) RatingDown (id int) revel.Result {
	_, err := c.Txn.Exec("UPDATE QdbEntry SET Rating = Rating - 1 WHERE QuoteId = ?", id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(nil)
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(nil)
}

