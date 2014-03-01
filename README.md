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

API
---

### Quote ###

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
		<tr>
			<td>UserId</td> <td>string</td> <td>Author's user id hash</td>
		</tr>
	</tbody>
</table>

All resources return *200* on success or *500* with an undefined body 
if fatal errors were encountered. Resources requiring an id return a 
*404* with undefined body if the id does not exist in the database. 

### Index

	GET /api/v1
	GET /api/v1/:id

#### Parameters
<table>
	<tr>
		<td>id</td> <td>return single entry</td>
	</tr>
	<tr>
		<td>tag</td> <td>return quotes with tag</td>
	</tr>
	<tr>
		<td>search</td> <td>search quotes</td>
	</tr>
	<tr>
		<td>from</td> <td>date string in the form mm/dd/yyyy</td>
	</tr>
	<tr>
		<td>to</td> <td>same as above</td>
	</tr>
	<tr>
		<td>sort</td> <td>one of relevence, rating, date, random</td>
	</tr>
	<tr>
		<td>desc</td> <td>boolean, sort in descending order</td>
	</tr>
	<tr>
		<td>size</td> <td>maximum amount of entries to return, capped at 4096 by default</td>
	</tr>
</table>

### Insert a new entry

	POST /api/v1

Accepts *Quote*, *Author* and *Tags* fields of a *Quote*

Request:

	POST /api/v0/ HTTP/1.1
	Content-Type: application/json
	Content-Length: 58

	{"Quote": "test", "Author": "jgr", "Tags": ["foo", "bar"]}

Response:

	HTTP/1.1 201 Created
	Content-Type: application/json	
	...

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

### Vote

	PATCH /api/v1/:id/:typ

#### Parameters

<table>
	<tr>
		<td>id</td> <td>target id</td>
	</tr>
	<tr>
		<td>typ</td> <td>one of up, down, delete</td>
	</tr>
</table>
