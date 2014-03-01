package controllers

import (
	"github.com/robfig/revel"
//	"net/http"
)

// Helper for sending utf-8 encoded plain text
type Utf8Result string

func (u Utf8Result) Apply(req *revel.Request, resp *revel.Response) {
	resp.WriteHeader(resp.Status, "text/plain; charset=utf-8")
	resp.Out.Write([]byte(u))
}
