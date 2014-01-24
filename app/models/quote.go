package models

import (
	"github.com/robfig/revel"
	"github.com/coopernurse/gorp"
	"reflect"
	"strings"
	"time"
	"errors"
)

const (
	TAG_DELIM = ","

	QUOTE_CHAR_MAX = 1024
	QUOTE_AUTHOR_MAX = 32

	TAG_PER_QUOTE_MAX = 32
	TAG_CHAR_MAX = 32
)

type (
	Quote struct {
		QuoteId int
		Quote   string
		Author  string
		Created int64
		Rating  int
		Tags    TagsArray
		UserId	string
	}

	TagsArray []string

	QuoteEntry struct {
		QuoteId int
		Quote   string
		Author  string
		Created int64
		Rating  int
		UserId  string
	}

	TagEntry struct {
		QuoteId int
		Tag     string
	}
)

var TagsBinder = revel.Binder{
	Bind: revel.ValueBinder(func(v string, t reflect.Type) reflect.Value {
		if len(v) == 0 {
			return reflect.Zero(t)
		}
		s := strings.Split(v, TAG_DELIM)
		for i := range s {
			s[i] = strings.TrimSpace(s[i])
		}
		return reflect.ValueOf(s)
	}),
	Unbind: nil,
}

func init () {
	revel.TypeBinders[reflect.TypeOf(TagsArray{})] = TagsBinder
}

func (quote *Quote) Validate (v *revel.Validation) {
	// TODO: add messages

	v.Check(quote.Author,
		revel.Required{},
		revel.MaxSize{QUOTE_AUTHOR_MAX},
	)

	v.Check(quote.Quote,
		revel.Required{},
		revel.MaxSize{QUOTE_CHAR_MAX},
	)

	v.MaxSize(quote.Tags, TAG_PER_QUOTE_MAX)

	for _, t := range quote.Tags {
		v.MaxSize(t, TAG_CHAR_MAX).Key("quote.Tags")
	}
}

func (q *QuoteEntry) PreInsert (s gorp.SqlExecutor) error {
	q.Created = time.Now().Unix()
	q.Rating  = 0
	return nil
}

type QuoteTypeConverter struct {}

func (me QuoteTypeConverter) ToDb (val interface{}) (interface{}, error) {
	return val, nil
}

func (me QuoteTypeConverter) FromDb (target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {

		// split csv values from Quote Tags column
		case *TagsArray:
			binder := func (holder, target interface{}) error {
				s, ok := holder.(*string)

				if !ok {
					return errors.New("FromDb: Unable to convert holder")
				}

				sl, ok := target.(*TagsArray)

				if !ok {
					return errors.New("FromDb: Unable to convert target")
				}

				if *s == "" {
					*sl = TagsArray{}
				} else {
					*sl = strings.Split(*s, ",")
				}
				return nil
			}
			return gorp.CustomScanner{new(string), target, binder}, true
	}

	return gorp.CustomScanner{}, false
}
