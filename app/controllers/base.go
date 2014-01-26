package controllers

import (
	"github.com/robfig/revel"
	"github.com/PacketFire/goqdb/app/models"
	"strings"
	"time"
)

type Base struct {
	GorpController
}

const (
	POST_THRESHOLD = 30

	ERR_POST_THRESHOLD_REACHED = "You have reached your maximum post limit"
	ERR_MULTIPLE_VOTE = "You may not vote more than once"
	ERR_ID_NOT_FOUND = "id not found"
)

func (c *Base) getEntries (arg models.Args) []models.Quote {
	var entries []models.Quote

	_, err := c.Txn.Select(&entries,
		`SELECT * FROM Quote` +
		where(arg) + sort(arg) +
		` LIMIT :Size * (:Page - 1), :Size`,
		arg,
	)

	if err != nil {
		panic(err)
	}
	return entries
}

func where (arg models.Args) string {
	var s []string
	if arg.Tag != "" {
		s = append(s, `QuoteId IN (` +
			`SELECT QuoteId FROM TagEntry WHERE Tag = :Tag)`,
		)
	}

	if arg.Search != "" {
		s = append(s, ` Quote LIKE '%' || :Search || '%'`)
		if arg.Tag == "" {
			s[len(s) - 1] += ` OR Tags LIKE '%' || :Search || '%'`
		}
	}

	if arg.From != 0 {
		s = append(s, ` Created BETWEEN :From AND :To`)
	}

	if len(s) > 0 {
		return ` WHERE ` + strings.Join(s, ` AND `)
	} else {
		return ``
	}
}

func sort (arg models.Args) string {

	var q = map[string]string{
		models.SORT_DATE: ` Created`,
		models.SORT_RATING: ` Rating`,
		models.SORT_RANDOM: ` Random()`,
		models.SORT_RELEVANCE: ` CASE` +
			` WHEN Quote LIKE :Search || '%' THEN 0` +
			` WHEN Quote LIKE '%' || :Search || '%' THEN 1` +
			` WHEN Tags LIKE '%' || :Search || '%' THEN 2` +
			` ELSE 3 END`,
	}

	if arg.Sort == models.SORT_RELEVANCE && arg.Search == "" {
		arg.Sort = models.SORT_DATE
	}

	s, ok := q[arg.Sort]

	if !ok {
		if arg.Search != "" {
			arg.Sort = models.SORT_RELEVANCE
		} else {
			arg.Sort = models.SORT_DATE
		}
		s = q[arg.Sort]
	}

	if arg.Desc {
		s += ` DESC`
	}

	return ` ORDER BY` + s
}

func (c *Base) getEntry (id int) []models.Quote {
	var entries []models.Quote
	_, err := c.Txn.Select(&entries,
		`SELECT * FROM Quote WHERE QuoteId == ? LIMIT 1`, id)

	if err != nil {
		panic(err)
	}

	if len(entries) < 1 {
		// not working for some wierd reason
		//c.Flash.Error("id does not exist") 
		c.Flash.Data["error"] = ERR_ID_NOT_FOUND
	}

	return entries
}

func (c *Base) getTotal (arg models.Args) int64 {
	total, err := c.Txn.SelectInt(`SELECT COUNT(*) FROM Quote` + where(arg), arg)

	if err != nil {
		panic(err)
	}
	return total
}

func (c *Base) insertQuote (quote *models.Quote) {

	userId := c.Session.Id()

	if !c.canPost(userId) {
		c.Flash.Error(ERR_POST_THRESHOLD_REACHED)
		return
	}

	quote.Validate(c.Validation)

	if c.Validation.HasErrors() {
		c.FlashParams()
		c.Validation.Keep()
		return
	}

	entry := models.QuoteEntry{
		Author: quote.Author,
		Quote: quote.Quote,
		UserId: userId,
	}

	err := c.Txn.Insert(&entry)

	if err != nil {
		revel.TRACE.Println(err)
		panic(err)
	}

	for i, s := range quote.Tags {
		if n := len(s); n > models.TAG_CHAR_MAX || n == 0 {
			continue
		}

		_, err = c.Txn.Exec(
			`INSERT OR IGNORE INTO TagEntry VALUES(?,?)`,
			entry.QuoteId, s)

		if err != nil {
			revel.TRACE.Println(err)
			panic(err)
		}

		if i >= models.TAG_PER_QUOTE_MAX {
			break
		}
	}

	return
}

func (c *Base) canPost (userId string) bool {

	t := time.Now().Truncate(24 * time.Hour).Unix()

	n, err := c.Txn.SelectInt(`SELECT COUNT(*) FROM QuoteEntry WHERE` +
		` UserId = ? AND Created >= ?`,
		userId, t)

	if err != nil {
		revel.TRACE.Println(err)
		panic(err)
	}

	if n >= POST_THRESHOLD {
		return false
	}

	return true
}

func (c *Base) vote (voteId int, voteType string) {
	c.Validation.Required(voteId)
	c.Validation.Check(voteType,
		revel.Required{},
		revel.Match{models.ValidVoteType},
	)

	if c.Validation.HasErrors() {
		c.FlashParams()
		c.Validation.Keep()
		return
	}

	userId := c.Session.Id()

	if voteType != models.VOTE_DELETE {

		if !c.canVote(voteId, userId) {
			c.Flash.Error(ERR_MULTIPLE_VOTE)
			return
		}

		query := `UPDATE QuoteEntry SET Rating = Rating`
		switch voteType {
			case models.VOTE_UP:
				query += ` +`
			case models.VOTE_DOWN:
				query += ` -`
		}
		query += ` 1 WHERE QuoteId = ?`

		result, _ := c.Txn.Exec(query, voteId)

		if affected, err := result.RowsAffected(); err != nil {
			panic(err)
			return
		} else if affected == 0 {
			c.Flash.Error(ERR_ID_NOT_FOUND)
			return
		}
	} else {
		t := time.Now().Truncate(24 * time.Hour).Unix()
		res, err := c.Txn.Exec(
			`DELETE FROM QuoteEntry ` +
			`WHERE QuoteId = ? AND UserId = ? AND Created >= ?`,
			voteId, userId, t)

		if err != nil {
			panic(err)
		}
		if n, err := res.RowsAffected(); n >= 1 {
			return
		} else if err != nil {
			panic(err)
		}
	}

	err := c.Txn.Insert(&models.VoteEntry{
		QuoteId: voteId, VoteType: voteType, UserId: userId})

	if err != nil {
		panic(err)
	}

	return
}

func (c *Base) canVote (voteId int, userId string) bool {

	t := time.Now().Truncate(24 * time.Hour).Unix()

	n, err := c.Txn.SelectInt(
		`SELECT COUNT(*) FROM VoteEntry WHERE` +
			` QuoteId = ? AND UserId = ? AND Created >= ?`,
		voteId, userId, t,
	)

	if err != nil {
		panic(err)
	}

	if n > 0 {
		return false
	}

	return true
}

