package controllers

import (
	"github.com/robfig/revel"

	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"

	"net/http"
	"encoding/json"

	"reflect"
	"time"
	"fmt"
)

type Api struct {
	GorpController
}

var (
// binders

	// date range binder
	RangeBinder = revel.Binder{
		Bind: func (params *revel.Params, name string, typ reflect.Type) reflect.Value {
			var Y, m, d int

			params.Bind(&Y, "Y")
			params.Bind(&m, "m")
			params.Bind(&d, "d")

			var r models.DateRange

			if Y == 0 {
				Y = time.Now().Year()
			}

			toUnix := func (Y, m, d int) int64 {

				t, _ := time.Parse("2006-Jan-2",
					fmt.Sprintf("%d-%s-%d",
					Y, time.Month(m).String()[:3], d))

				return t.Unix()
			}

			if d != 0 {
				r.Upper = toUnix(Y, m, d + 1)
				r.Lower = toUnix(Y, m, d)
			} else if m != 0 {
				r.Upper = toUnix(Y, m + 1, 1)
				r.Lower = toUnix(Y, m, 1)
			} else {
				r.Upper = toUnix(Y + 1, 1, 1)
				r.Lower = toUnix(Y, 1, 1)
			}


			return reflect.ValueOf(r)
		},
		Unbind: nil,
	}
)

func init () {
	revel.TypeBinders[reflect.TypeOf(models.DateRange{})] = RangeBinder
}

// index
func (c *Api) Index (R models.DateRange) revel.Result {

	params := make(map[string]interface{})

	params["lower"] = R.Lower
	params["upper"] = R.Upper
	params["max"] = VIEW_SIZE_MAX

	var entries []models.QdbView

	_, err := c.Txn.Select(&entries, `
		SELECT 
			* 
		FROM 
			QdbView
		WHERE
			Created BETWEEN :lower AND :upper
		LIMIT :max`,
		params,
	)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
	}

	return c.RenderJson(entries)
}

// post
func (c *Api) Post () revel.Result {
	var quote models.QdbView

	dec := json.NewDecoder(c.Request.Body)
	err := dec.Decode(&quote)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(err)
	}

	if quote.Quote == "" || quote.Author == "" {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(quote)
	}

	err = c.insertView(&quote)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(err)
	}

	c.Response.Status = http.StatusCreated

//	return c.RenderJson(quote)
	return c.Redirect(routes.Api.One(quote.QuoteId))
}

func (c *Api) One (id int) revel.Result {
	obj, err := c.Txn.Get(models.QdbEntry{}, id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(err)
	}

	if obj == nil {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	entry := obj.(*models.QdbEntry)
	return c.RenderJson(entry)
}

func (c *Api) UpVote (id int) revel.Result {
	found, err := c.upVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(err)
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	return c.Redirect(routes.Api.One(id))
}

func (c *Api) DownVote (id int) revel.Result {
	found, err := c.downVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(err)
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(err)
	}

	return c.Redirect(routes.Api.One(id))
}
