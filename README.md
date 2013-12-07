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

### QdbView ###

<table>
	<thead>
		<tr>
			<th>Name</th> <th>Type</th> <th>Description</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>QuoteId</td> <td>int</td> <td>The quote id</td>
		</tr>
		<tr>
			<td>Quote</td> <td>string</td> <td>The quote body</td>
		</tr>
		<tr>
			<td>Author</td> <td>string</td> <td>The author of the quote</td>
		</tr>
		<tr>
			<td>Created</td> <td>int64</td> <td>unix time in seconds</td>
		</tr>
		<tr>
			<td>Rating</td> <td>int</td> <td>The quote's rating</td>
		</tr>
		<tr>
			<td>Tags</td> <td>[]string</td> <td>An array of tag strings, white space is trimmed from either side</td>
		</tr>
	</tbody>
</table>

All resources return *200* on success or *500* with an undefined body 
if fatal errors were encountered. Resources requiring an id return a 
*404* with undefined body if the id does not exist in the database. 

### Authentication

Authentication uses HMAC-SHA256 of the concatenation of
the request URL and, request body (if there is one), 
signed by the private key associated with the user's api key.
This is sent using the Authorization header as so:

<code>
	Authorization: HMAC <em>apikey</em>:<em>digest</em>
</code>

Note: All API actions require authentication

### Main listing, sorted by entry id

	GET /api/v0

### Retrieve entries by date

Year/Month/Day, Eg.

	GET /api/v0/2010/7/1

### Insert a new entry

	POST /api/v0

Accepts *Quote*, *Author* and *Tags* fields of a *QdbView*

Note: POST returns 201 Created on success and 400 Bad Request
if the post data did not pass validation

Request:

	POST /api/v0/ HTTP/1.1
	Authorization: HMAC jgr:ee64125d7a28bda807a0276aa6705b607181f1cb6cd65470c2def0b5f160d9ee
	Content-Type: application/json
	Content-Length: 58

	{"Quote": "test", "Author": "jgr", "Tags": ["foo", "bar"]}

Response:

	HTTP/1.1 201 Created
	Content-Length: 135
	Content-Type: application/json
	Date: Mon, 25 Nov 2013 16:11:28 GMT
	Set-Cookie: REVEL_FLASH=; Path=/
	Set-Cookie: REVEL_SESSION=64172d7dab5d922c6cdc2ca993e72647e3585d75-%00_TS%3A1387987888%00; Path=/; Expires=Wed, 25 Dec 2013 16:11:28 UTC

	{
	  "QuoteId": 20,
	  "Quote": "test",
	  "Created": 1385395888,
	  "Rating": 0,
	  "Author": "jgr",
	  "Tags": [
	    "foo",
	    "bar"
	 ]
	}

### Retrieve quote entry

*:id* is used here in place of the quote id for the target entry

	GET /api/v0/:id/view

### Upvote a quote

	PUT /api/v0/:id/rating

### Downvote a quote

	DELETE /api/v0/:id/rating

