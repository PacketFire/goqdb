package models

import (
	"github.com/coopernurse/gorp"
	"regexp"
	"strings"
	"time"
)

const (
	VOTE_UP = "up"
	VOTE_DOWN = "down"
	VOTE_DELETE = "delete"
)

type VoteEntry struct {
	UserId   string
	QuoteId  int
	VoteType string
	Created  int64
}

var (
	VoteTypeStrings = []string{
		VOTE_UP,
		VOTE_DOWN,
		VOTE_DELETE,
	}

	ValidVoteType = regexp.MustCompile(
		strings.Join(VoteTypeStrings, `|`),
	)

)

func (v *VoteEntry) PreInsert (s gorp.SqlExecutor) error {
	v.Created = time.Now().Unix()
	return nil
}
