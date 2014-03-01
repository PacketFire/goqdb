package models

import (
	"github.com/robfig/revel"
	"time"
	"reflect"
	"regexp"
	"strings"
)

const (
	SIZE_MAX = 4096
	SIZE_DEFAULT = 5

	SORT_DATE = "date"
	SORT_RELEVANCE = "relevance"
	SORT_RATING = "rating"
	SORT_RANDOM = "random"
)

type (
	UnixEpoch int64

	Args struct {
		Id int
		Tag, Search string
		From, To UnixEpoch

		Sort string
		Desc bool

		Page, Size int
	}
)

var (
	SortStrings = []string{
		SORT_DATE, SORT_RELEVANCE,
		SORT_RATING, SORT_RANDOM,
	}

	validSort = regexp.MustCompile(strings.Join(SortStrings, `|`))

	UnixEpochBinder = revel.Binder{
		Bind: func (params *revel.Params, key string, typ reflect.Type) reflect.Value {
			var t time.Time

			params.Bind(&t, key)

			if t.IsZero() {
				return reflect.Zero(typ)
			}

			return reflect.ValueOf(UnixEpoch(t.Unix()))
		},
		Unbind: func (output map[string]string, key string, value interface{}) {
			e := value.(UnixEpoch).Time()
			output[key] = e.Format(revel.DateFormat)
		},
	}

	ArgsBinder = revel.Binder{
		Bind: func (params *revel.Params, key string, t reflect.Type) reflect.Value {
			var arg Args

			val := reflect.ValueOf(&arg).Elem()
			typ := val.Type()

			for i := 0; i < typ.NumField(); i++ {
				dest := val.Field(i)
				dest.Set(
					revel.Bind(params,
						toKey(typ.Field(i).Name),
						dest.Type(),
					),
				)
			}
			return reflect.ValueOf(arg)
		},
		Unbind: func (output map[string]string, key string, value interface{}) {
			arg := value.(Args)
			val := reflect.ValueOf(&arg).Elem()
			typ := val.Type()

			if IsZero(val) {
				return
			}

			for i := 0; i < typ.NumField(); i++ {
				if field := val.Field(i); !IsZero(field) {
					revel.Unbind(output,
						toKey(typ.Field(i).Name),
						field.Interface(),
					)
				}
			}
		},
	}
)

func init () {

	revel.TypeBinders[reflect.TypeOf(Args{})] = ArgsBinder
	revel.TypeBinders[reflect.TypeOf(UnixEpoch(0))] = UnixEpochBinder

	revel.TemplateFuncs["tagarg"] = func (tag string) Args {
		return Args{Tag: tag}
	}

	revel.TemplateFuncs["idarg"] = func (id int) Args {
		return Args{Id: id}
	}

	revel.TemplateFuncs["sorts"] = func () []string {
		return SortStrings
	}

	revel.TemplateFuncs["sortby"] = func (sort string) Args {
		return Args{Sort: sort}
	}

	revel.TemplateFuncs["prev"] = func (hasPrev bool, arg Args) (string, error) {
		if !hasPrev {
			return "", nil
		}

		if arg.Page > 0 {
			arg.Page -= 1
		}

		return revel.ReverseUrl("App.Index", arg)
	}

	revel.TemplateFuncs["next"] = func (hasNext bool, arg Args) (string, error) {
		if !hasNext {
			return "", nil
		}

		if arg.Page == 0 {
			arg.Page = 2
		} else {
			arg.Page += 1
		}

		return revel.ReverseUrl("App.Index", arg)
	}
}

func (u UnixEpoch) Int () int64 {
	return int64(u)
}

func (u UnixEpoch) Time () time.Time {
	return time.Unix(u.Int(), 0)
}

// Returns a new Args struct containing all non-zero values from dest and
// non-zero values from src for which there is a zero value in dest
// note: reverse merge for templating: Args{Sort:"relevence"}.Merge(arg)
func (dest Args) Merge (src Args) Args {

// more horrible reflection

	destVal := reflect.ValueOf(&dest).Elem()
	srcVal := reflect.ValueOf(&src).Elem()

	typ := destVal.Type()

	for i := 0; i < typ.NumField(); i++ {
		destField := destVal.Field(i)
		srcField := srcVal.Field(i)
		if IsZero(destField) && !IsZero(srcField) {
			destField.Set(srcField)
		}
	}

	return dest
}

