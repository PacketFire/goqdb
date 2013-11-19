package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"
	"net/http"
)
type Api struct {
	Core
}

func (c *Api) Index () revel.Result {

	entries, err := c.getEntries(0, -1, "")

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	}

	return c.RenderJson(entries)
}

func (c *Api) Post (quote, author string) revel.Result {

	if quote == "" || author == "" {
		c.Response.Status = http.StatusBadRequest
	} else {
		err := c.insertEntry(models.QdbEntry{Quote: quote, Author: author})

		if err != nil {
			c.Response.Status = http.StatusInternalServerError
		} else {
			c.Response.Status = http.StatusCreated
		}
	}
	return c.RenderJson(nil)
}

func (c *Api) One (id int) revel.Result {

	entries, err := c.getEntryById(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	} else {
		if len(entries) == 0 {
			c.Response.Status = http.StatusNotFound
		}
	}
	return c.RenderJson(entries)
}

func (c *Api) UpVote (id int) revel.Result {

	changes, err := c.upVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	} else {
		if changes == 0 {
			c.Response.Status = http.StatusNotFound
		}
	}
	return c.RenderJson(nil)
}

func (c *Api) DownVote (id int) revel.Result {

	changes, err := c.downVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	} else {
		if changes == 0 {
			c.Response.Status = http.StatusNotFound
		}
	}
	return c.RenderJson(nil)
}
