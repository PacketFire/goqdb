package controllers

import (
	"github.com/PacketFire/goqdb/app/models"
)

func loadEntries (results []interface{}, err error) []*models.QdbEntry {

	/* note: excise panic() */
	if err != nil {
		panic(err)
	}

	var entries []*models.QdbEntry

	for _, r := range results {
		entries = append(entries, r.(*models.QdbEntry))
	}

	return entries
}

