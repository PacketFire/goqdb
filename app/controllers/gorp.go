package controllers

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	r "github.com/robfig/revel"
	"github.com/robfig/revel/modules/db/app"

	"github.com/PacketFire/goqdb/app/models"
)

var (
	Dbm *gorp.DbMap
)

func Init() {
	db.Init()
	Dbm = &gorp.DbMap{Db: db.Db, Dialect: gorp.SqliteDialect{}}
	Dbm.TraceOn("[gorp]", r.INFO)

	Dbm.AddTable(models.QuoteEntry{}).SetKeys(true, "QuoteId")

	Dbm.AddTable(models.TagEntry{}).SetKeys(false, "QuoteId", "Tag")
	Dbm.AddTable(models.VoteEntry{}).SetKeys(false, "UserId", "QuoteId")

	Dbm.CreateTables()

	Dbm.Exec(`
		CREATE VIEW IF NOT EXISTS Quote AS
			SELECT QuoteEntry.*, IFNULL(G.Tags, "") AS Tags
			FROM QuoteEntry
			LEFT JOIN (
				SELECT TagEntry.QuoteId,
				       GROUP_CONCAT(TagEntry.Tag, ',') AS Tags
				FROM TagEntry
				GROUP BY TagEntry.QuoteId
			) AS G
			ON G.QuoteId = QuoteEntry.QuoteId`,
	)
	Dbm.TypeConverter = models.QuoteTypeConverter{}
}

type GorpController struct {
	*r.Controller
	Txn *gorp.Transaction
}

func (c *GorpController) Begin() r.Result {
	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

func (c *GorpController) Commit() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
