package controllers

import (
	"github.com/robfig/revel"

	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"

	"net/http"

	"reflect"
	"strings"
	"fmt"
)

type App struct {
	GorpController
}


var (
	// form input tag delimiter
	TagDelim = " "

	// order input -> order column
	Orders = map[string]string{
		"created": "Created",
		"rating":  "Rating",
	}

	// form input "c"sv tags binder
	TagsBinder = revel.Binder{

		Bind: revel.ValueBinder(func (val string, typ reflect.Type) reflect.Value {
			if len(val) == 0 {
				return reflect.Zero(typ)
			}
			s := strings.Split(val, TagDelim)

			return reflect.ValueOf(s)
		}),
		Unbind: nil,
	}

	PaginationBinder = revel.Binder{
		Bind: func (params *revel.Params, name string, typ reflect.Type) reflect.Value {
			var p models.Pagination

			params.Bind(&p.Page,   "page")

			if p.Page == 0 {
				p.Page = 1
			}

			params.Bind(&p.Size,   "size")

			if p.Size != 0 && p.Size > VIEW_SIZE_MAX {
				p.Size = VIEW_SIZE_DEFAULT
			}

			params.Bind(&p.Search, "search")
			p.Search = strings.TrimSpace(p.Search)

			params.Bind(&p.Tag, "tag")
			p.Tag = strings.TrimSpace(p.Tag)

			params.Bind(&p.Order, "order")
			p.Order = strings.TrimSpace(p.Order)

			p.HasNext = false
			p.HasPrev = false

			return reflect.ValueOf(p)
		},
		Unbind: func (output map[string]string, key string, val interface{}) {
			p := val.(models.Pagination)

			if p.Page != 0 {
				revel.Unbind(output, "page", p.Page)
			}
			if p.Size != 0 {
				revel.Unbind(output, "size", p.Size)
			}
			if p.Search != "" {
				revel.Unbind(output, "search", p.Search)
			}

			if p.Tag != "" {
				revel.Unbind(output, "tag", p.Tag)
			}

			if p.Order != "" {
				revel.Unbind(output, "order", p.Order)
			}

		},
	}

)

func init () {
	revel.ERROR_CLASS = "has-error"

	revel.TypeBinders[reflect.TypeOf(models.TagArray{})] = TagsBinder
	revel.TypeBinders[reflect.TypeOf(models.Pagination{})] = PaginationBinder
}

func (c App) Index (page models.Pagination) revel.Result {

	var savedAuthor string

	if author, ok := c.Session["author"]; ok {
		savedAuthor = author
	}

	params := make(map[string]interface{})

	params["search"] = "%"+page.Search+"%"
	params["tag"]    = page.Tag
	params["order"]  = page.Order

	var where string

	if page.Tag != "" {
		where = `
		WHERE QuoteId IN (
			SELECT TagEntry.QuoteId FROM TagEntry
			WHERE TagEntry.Tag = :tag
		) `
	} else {
		where = ` WHERE Quote LIKE :search AND Tags LIKE :search `
	}

	where += ` ORDER BY :order DESC `

	count, err := c.Txn.SelectInt(`SELECT COUNT(*) FROM QdbView ` + where, params)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		// TODO: redirect to error page
		panic(err)
	}

	var size int

	if page.Size == 0 {
		size = VIEW_SIZE_DEFAULT
	} else {
		size = page.Size
	}

	offset := size * (page.Page - 1)

	params["offset"] = offset
	params["size"]   = size

	var entries []models.QdbView

	_, err = c.Txn.Select(&entries,
		`SELECT * FROM QdbView ` + where + ` LIMIT :offset, :size`,
		params,
	)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		//TODO: redirect to error page
		panic(err)
	}

	page.HasPrev = offset > 0

	page.HasNext = int64(offset + size) < count

	return c.Render(entries, page, savedAuthor)
}

// post
func (c *App) Post (quote models.QdbView, page models.Pagination) revel.Result {
	quote.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
	} else {
		err := c.insertView(&quote)

		if err != nil {
			c.Response.Status = http.StatusInternalServerError
			revel.ERROR.Print(err)
			//TODO: redirect to error page
			panic(err)
		}
	}

	c.Session["author"] = quote.Author

	return c.Redirect(routes.App.Index(page))
}

func (c *App) One (id int) revel.Result {
	obj, err := c.Txn.Get(models.QdbEntry{}, id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return Utf8Result(fmt.Sprint(err))
	}

	if obj == nil {
		c.Response.Status = http.StatusNotFound
		return Utf8Result(fmt.Sprintf("No such id: %d", id))
	}

	entry := obj.(*models.QdbEntry)

	return Utf8Result(entry.Quote)
}

func (c *App) UpVote (id int, page models.Pagination) revel.Result {
	found, err := c.upVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		//TODO: redirect to error page
		return c.Redirect(routes.App.Index(page))
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		//TODO: what to do?
	}

	return c.Redirect(routes.App.Index(page))
}

func (c *App) DownVote (id int, page models.Pagination) revel.Result {
	found, err := c.downVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		//TODO: redirect to error page
		return c.Redirect(routes.App.Index(page))
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		//TODO: what to do?
	}

	return c.Redirect(routes.App.Index(page))
}
