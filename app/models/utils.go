package models

import (
	"github.com/robfig/revel"
	"reflect"
	_"fmt"
	_"errors"
	"time"
	"strings"
)

func init () {
	revel.TemplateFuncs["epochdate"] = func (v int64) string {
		return time.Unix(v, 0).Format(revel.DateFormat)
	}

	revel.TemplateFuncs["epochdatetime"] = func (v int64) string {
		return time.Unix(v, 0).Format(revel.DateTimeFormat)
	}

	revel.TemplateFuncs["iszero"] = func (v interface{}) bool {
		return IsZero(reflect.ValueOf(v))
	}

	revel.TemplateFuncs["notzero"] = func (v interface{}) bool {
		return !IsZero(reflect.ValueOf(v))
	}
}

func IsZero (v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func toKey (name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}

func toField (name string) string {
	return strings.ToUpper(name[:1]) + name[1:]
}
