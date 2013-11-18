goqdb
=====

Quote Database in Go

usage
=====

requirements
------------
* http://robfig.github.io/revel/
* https://github.com/mattn/go-sqlite3
* working sqlite3 library

running
-------

```
revel run github.com/PacketFire/goqdb
```

api
---

The API results are displayed in JSON.

Returned field names:

* QuoteId: int
* Quote: string
* Author: string
* Created: 64 bit int (unix time in seconds)
* Rating: int

Resources:

* /api/v0

	+ GET: Retrieve the entire database

	+ POST: Insert a new entry. The accepted fields are:

			- "entry.Author"
			- "entry.Quote"


* /api/v0/:id
	+ GET: Retrieve the entry of the id

* /api/v0/:id/rating

	+ PUT: upvote

	+ DELETE: downvote

