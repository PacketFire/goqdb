package controllers

import (
	"github.com/robfig/revel"

	"github.com/PacketFire/goqdb/app/models"
	"github.com/PacketFire/goqdb/app/routes"

	"net/http"
	"encoding/json"
	_"encoding/base64"
	"encoding/hex"

	"reflect"
	"time"
	"fmt"
	"strings"

	"io/ioutil"
	"bytes"

	"crypto/hmac"
	"crypto/sha256"

	"errors"
)

type Api struct {
	GorpController
}

type ApiAuth struct {
	ApiAuthId int
	ApiKey string
	PrivKey []byte
}

var (

	// date range binder
	RangeBinder = revel.Binder{
		Bind: func (params *revel.Params, name string, typ reflect.Type) reflect.Value {
			var Y, m, d int

			params.Bind(&Y, "Y")
			params.Bind(&m, "m")
			params.Bind(&d, "d")

			var r models.DateRange

			if Y == 0 {
				return reflect.Zero(typ)
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

func (c *Api) Authenticate () revel.Result {

	r := c.Request.Header.Get("Authorization")

	if r == "" {
		c.Response.Status = http.StatusUnauthorized
		revel.TRACE.Print("Authorization header not supplied")
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	s := strings.Split(r, " ")

	if s[0] != "HMAC" {
		c.Response.Status = http.StatusUnauthorized
		revel.TRACE.Printf("no 'HMAC' marker in field: %q", r)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	s = strings.Split(s[1], ":")

	t, err := c.Txn.SelectStr(`SELECT IFNULL(PrivKey, "") FROM ApiAuth WHERE ApiKey = ? LIMIT 1`, s[0])

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Printf("error selecting private key from db: %q", s[0])
		revel.ERROR.Print(err)
		panic(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	if t == "" {
		c.Response.Status = http.StatusUnauthorized
		revel.TRACE.Printf("key does not exist in db: %q", s[0])
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	key := []byte(t)

	d, err := hex.DecodeString(s[1])

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Printf("error decoding hex string: %q", s[1])
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print("error reading request body")
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	c.Request.Body.Close()
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	blob := []byte(c.Request.URL.String() + string(body))

//	revel.TRACE.Printf("blob: %s\n", string(blob))

	mac := hmac.New(sha256.New, key)
	mac.Write(blob)

	if !hmac.Equal(mac.Sum(nil), d) {
		c.Response.Status = http.StatusUnauthorized
		revel.TRACE.Printf("authentication failure:\ngot: %x\nexpected: %x", mac.Sum(nil), d)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	return nil
}

// index
func (c *Api) Index (R models.DateRange) revel.Result {

	params := make(map[string]interface{})
	params["max"] = VIEW_SIZE_MAX

	query := `SELECT * FROM QdbView `

	var entries []models.QdbView
	if R.Lower == 0 {
		query += ` ORDER BY QuoteId ASC `
	} else {

		params["lower"] = R.Lower
		params["upper"] = R.Upper

		query += ` WHERE Created BETWEEN :lower AND :upper `
	}

	query += ` LIMIT :max `

	_, err := c.Txn.Select(&entries, query, params)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
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
		revel.ERROR.Print("error decoding request body")
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	if quote.Quote == "" || quote.Author == "" {
		c.Response.Status = http.StatusBadRequest
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	err = c.insertView(&quote)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print("error inserting quote")
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	c.Response.Status = http.StatusCreated

//	return c.RenderJson(quote)
	return c.Redirect(routes.Api.One(quote.QuoteId))
}

func (c *Api) One (id int) revel.Result {
	obj, err := c.Txn.Get(models.QdbEntry{}, id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Printf("error retreiving from db: %d", id)
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	if obj == nil {
		c.Response.Status = http.StatusNotFound
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	entry := obj.(*models.QdbEntry)
	return c.RenderJson(entry)
}

func (c *Api) Total () revel.Result {
	total, err := c.Txn.SelectInt(`SELECT COUNT(*) FROM QdbView`)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print("error retreiving total from db")
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	return c.RenderJson(total)
}

func (c *Api) UpVote (id int) revel.Result {
	found, err := c.upVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Printf("error upvoting: %d", id)
		revel.ERROR.Print(err)

		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	return c.Redirect(routes.Api.One(id))
}

func (c *Api) DownVote (id int) revel.Result {
	found, err := c.downVote(id)

	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Printf("error downvoting: %d", id)
		revel.ERROR.Print(err)
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	if !found {
		c.Response.Status = http.StatusNotFound
		return c.RenderError(errors.New(http.StatusText(c.Response.Status)))
	}

	return c.Redirect(routes.Api.One(id))
}
