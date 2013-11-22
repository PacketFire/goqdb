package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"
	"net/http"
	"encoding/json"
)
type Api struct {
	Core
}

func (c *Api) Index () revel.Result {

	entries, err := c.getEntries(0, -1, "")

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
	}
	c.Response.ContentType = "application/json; charset=utf-8"

	return c.RenderJson(entries)
}

func (c *Api) Post () revel.Result {

	var post models.QdbEntry

	dec := json.NewDecoder(c.Request.Body)

	err := dec.Decode(&post)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(nil)
	}

	/* validation stuffs */
	if post.Quote == "" || post.Author == "" {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	err = c.insertEntry(&post)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(nil)
	}

	c.Response.Status = http.StatusCreated
	c.Response.ContentType = "application/json; charset=utf-8"
	return c.RenderJson(post)
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
	c.Response.ContentType = "application/json; charset=utf-8"
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
