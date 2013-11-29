package controllers

import (
	"database/sql"
	"github.com/PacketFire/goqdb/app/models"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	r "github.com/robfig/revel"
	"github.com/robfig/revel/modules/db/app"

	"errors"
	"strings"
)

var (
	Dbm *gorp.DbMap
)

type QdbTypeConverter struct {}

func (me QdbTypeConverter) ToDb (val interface{}) (interface{}, error) {
	return val, nil
}

func (me QdbTypeConverter) FromDb (target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {

		// split csv values from QdbView Tags column
		case *models.TagArray:
			binder := func (holder, target interface{}) error {
				s, ok := holder.(*string)

				if !ok {
					return errors.New("FromDb: Unable to convert holder")
				}

				sl, ok := target.(*models.TagArray)

				if !ok {
					return errors.New("FromDb: Unable to convert target")
				}

				if *s == "" {
					*sl = models.TagArray{}
				} else {
					*sl = strings.Split(*s, ",")
				}
				return nil
			}
			return gorp.CustomScanner{new(string), target, binder}, true
		}

	return gorp.CustomScanner{}, false
}

func Init() {
	db.Init()
	Dbm = &gorp.DbMap{Db: db.Db, Dialect: gorp.SqliteDialect{}}
	Dbm.TraceOn("[gorp]", r.INFO)

	Dbm.AddTable(models.QdbEntry{}).SetKeys(true, "QuoteId")

	Dbm.AddTable(models.TagEntry{}).SetKeys(false, "QuoteId", "Tag")

	Dbm.CreateTables()

	Dbm.TypeConverter = QdbTypeConverter{}

	Dbm.Exec(`
		CREATE VIEW QdbView AS
			SELECT QdbEntry.*, IFNULL(G.Tags, "") AS Tags
			FROM QdbEntry
			LEFT JOIN (
				SELECT TagEntry.QuoteId,
				       GROUP_CONCAT(TagEntry.Tag, ',') AS Tags
				FROM TagEntry
				GROUP BY TagEntry.QuoteId
			) AS G
			ON G.QuoteId = QdbEntry.QuoteId`,
	)

	// TagCloud is a representation of the most common tags of the most recent entries
	Dbm.Exec(`
		CREATE VIEW TagCloud AS
			SELECT TagEntry.Tag
			FROM TagEntry
			LEFT JOIN (
				SELECT QuoteId
				FROM QdbEntry
				ORDER BY Created DESC
				LIMIT 50
			) AS G
			ON G.QuoteId = TagEntry.QuoteId
			GROUP BY TagEntry.Tag
			ORDER BY count(TagEntry.Tag) DESC`,
	)
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
