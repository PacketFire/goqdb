package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/routes"
	"github.com/PacketFire/goqdb/app/models"
	"net/http"
	"encoding/json"
	"reflect"
	"time"
)

var DateRangeBinder = revel.Binder{
	Bind: func (params *revel.Params, name string, typ reflect.Type) reflect.Value {
		var Y, m, d int

		params.Bind(&Y, "Y")
		params.Bind(&m, "m")
		params.Bind(&d, "d")

		var R models.DateRange

		if d != 0 {
			R.Upper = time.Date(Y, time.Month(m), d + 1, 0, 0, 0, 0, time.UTC).Unix()
		} else if m != 0 {
			R.Upper = time.Date(Y, time.Month(m + 1), d, 0, 0, 0, 0, time.UTC).Unix()
		} else if Y != 0 {
			R.Upper = time.Date(Y + 1, time.Month(m), d, 0, 0, 0, 0, time.UTC).Unix()
		} else {
			R.Upper = time.Date(time.Now().Year() + 1, 0, 0, 0, 0, 0, 0, time.UTC).Unix()
			R.Lower = time.Date(time.Now().Year(), 0, 0, 0, 0, 0, 0, time.UTC).Unix()
			return reflect.ValueOf(R)
		}

		R.Lower = time.Date(Y, time.Month(m), d, 0, 0, 0, 0, time.UTC).Unix()

		return reflect.ValueOf(R)
	},
	Unbind: nil,
}

func init () {
	revel.TypeBinders[reflect.TypeOf(models.DateRange{})] = DateRangeBinder
}

type Api struct {
	Core
}

func (c *Api) Index (R models.DateRange) revel.Result {

	entries, err := c.getEntries(models.PageState{
			Page: 0 , Size: -1,
			Tag:"", Search:"",
		},
		R,
	)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
	} else if len(entries) == 0 {
		c.Response.Status = http.StatusNotFound
	}
	return c.RenderJson(entries)
}

func (c *Api) Post () revel.Result {

	var post models.QdbView

	dec := json.NewDecoder(c.Request.Body)

	err := dec.Decode(&post)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(err)
	}

	/* validation stuffs */
	if post.Quote == "" || post.Author == "" {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(err)
	}

	err = c.insertView(&post)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(err)
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJson(post)
}

func (c *Api) One (id int) revel.Result {

	entry, err := c.getEntry(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(err)
	}
	if entry.QuoteId == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.RenderJson(entry)
}

func (c *Api) UpVote (id int) revel.Result {

	changes, err := c.upVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(err)
	}
	if changes == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.Redirect(routes.Api.One(id))
}

func (c *Api) DownVote (id int) revel.Result {

	changes, err := c.downVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.TRACE.Print(err)
		return c.RenderJson(err)
	}
	if changes == 0 {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.Redirect(routes.Api.One(id))
}
