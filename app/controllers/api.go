package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"

	"encoding/json"
)

type Api struct {
	Base
}

var apiDefaults = models.Args{
	Size: 1024,
}.Merge(appDefaults)

func (c *Api) Index (arg models.Args) revel.Result {
	entries := c.getEntries(arg.Merge(apiDefaults))
	return c.RenderJson(entries)
}

func (c *Api) Single (id int) revel.Result {
	entry := c.getEntry(id)
	return c.RenderJson(entry)
}

func (c *Api) Total (arg models.Args) revel.Result {
	total := c.getTotal(arg.Merge(apiDefaults))
	return c.RenderJson(total)
}

func (c *Api) Insert () revel.Result {
	var quote models.Quote

	dec := json.NewDecoder(c.Request.Body)
	err := dec.Decode(&quote)

	if err != nil {
		panic(err)
	}

	c.insertQuote(&quote)

	return c.RenderJson(quote)
}

func (c *Api) Vote (id int, typ string) revel.Result {
	c.vote(id, typ)
	return c.RenderJson(nil)
}
