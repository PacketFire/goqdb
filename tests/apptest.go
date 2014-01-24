package tests

import (
	r "github.com/robfig/revel"
	m "github.com/PacketFire/goqdb/app/models"
	c "github.com/PacketFire/goqdb/app/controllers"
	"net/url"
	"os"
)

type AppTest struct {
	r.TestSuite
}

const DBPATH = "goqdb.db"

var (
	entry = url.Values{
		"quote.Quote": []string{"post test quote"},
		"quote.Author": []string{"test suite"},
		"quote.Tags": []string{"test, cruft, post"},
	}
)

func postFormEntry () url.Values {
	return entry
}

func (t *AppTest) Before () {
	r.TRACE.Println("-------------------------- Setup")

// TODO: use a mock database with pre-filled entries

//BUG: if testing the app while it is running you sometimes need to restart it to
// get the original db back.
	os.Rename(DBPATH, DBPATH + ".bak")
	c.Init()
}

func (t *AppTest) After() {
	r.TRACE.Println("-------------------------- Tear down")
	os.Rename(DBPATH + ".bak", DBPATH)
}

func (t AppTest) TestIndex () {
	t.Get("/")
	t.AssertOk()
}

// TODO: arg tests could be better
func (t AppTest) TestIdArg () {
	t.Get("/?id=1")
	t.AssertOk()
}

func (t AppTest) TestTagArg () {
	t.Get("/?tag=foo")
	t.AssertOk()
}

func (t AppTest) TestSearchArg () {
	t.Get("/?search=foo")
	t.AssertOk()
}

func (t AppTest) TestPageArg () {
	t.Get("/?page=1")
	t.AssertOk()
}

func (t AppTest) TestSizeArg () {
	t.Get("/?size=10")
	t.AssertOk()
}

func (t AppTest) TestFromArg () {
	t.Get("/?from=2014-1-1")
	t.AssertOk()
}

func (t AppTest) TestFromToArg () {
	t.Get("/?from=2014-1-1&to=2014-1-29")
	t.AssertOk()
}

func (t AppTest) TestDescArg () {
	t.Get("/?desc=true")
	t.AssertOk()
}

func (t AppTest) TestSortArg () {
	for _, s := range m.SortStrings {
		t.Get("/?sort=" + s)
		t.AssertOk()
	}
}

func (t AppTest) TestPost () {
	t.PostForm("/", postFormEntry())
	t.AssertOk()
}

func (t AppTest) TestPostThreshold () {
	for i := 0; i <= c.POST_THRESHOLD; i++ {
		t.PostForm("/", postFormEntry())
		t.AssertOk()
	}
	t.AssertContains(c.ERR_POST_THRESHOLD_REACHED)
}

func (t AppTest) TestVote () {
	t.PostForm("/", postFormEntry())
	t.AssertOk()

	t.Get("/vote/1/up")
	t.AssertOk()
}

func (t AppTest) TestVoteThreshold () {
	t.PostForm("/", postFormEntry())
	t.AssertOk()

	t.Get("/vote/1/up")
	t.AssertOk()

	t.Get("/vote/1/down")
	t.AssertOk()

	t.AssertContains(c.ERR_MULTIPLE_VOTE)
}
