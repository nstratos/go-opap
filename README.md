go-opap
==============

go-opap is a Go client library for accessing the [OPAP REST web services](https://www.opap.gr/en/web-services).

[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/nstratos/go-opap/opap?status.svg)](https://godoc.org/github.com/nstratos/go-opap/opap)
[![Go Report Card](https://goreportcard.com/badge/github.com/nstratos/go-opap)](https://goreportcard.com/report/github.com/nstratos/go-opap)
[![Coverage Status](https://coveralls.io/repos/github/nstratos/go-opap/badge.svg?branch=master)](https://coveralls.io/github/nstratos/go-opap?branch=master)
[![Build Status](https://travis-ci.org/nstratos/go-opap.svg?branch=master)](https://travis-ci.org/nstratos/go-opap)

Installation
------------

This package can be installed using:

	go get github.com/nstratos/go-opap/opap

Usage
-----

Import the package using:

```go
import "github.com/nstratos/go-opap/opap"
```

First construct a new opap client:

```go
c := opap.NewClient(nil)
```

The client supports three methods of returning draws for each game. It can return a specific draw by number or by date as well as the latest draw.
For example, to get the latest draw of the game Joker:

```go
c := opap.NewClient(nil)

draw, _, err := c.Draws.Latest(opap.Lotto)
// ...

draw, _, err := c.Draws.ByNumber(opap.Joker, 1873)
// ...

draws, _, err := c.Draws.ByDate(opap.Kino, 27, 12, 2017)
// ...
```

The Latest and ByNumber methods return one draw. The ByDate method returns a
slice of Draw objects. Each Draw object contains the draw time, the draw number
and the results as a slice of integers. It looks like this:

```
opap.Draw{
	DrawTime: "24-12-2017T22:00:00",
	DrawNo:   1873,
	Results:  []int{40, 13, 1, 24, 15, 8},
}
```

The number of results differs depending on the game. For example the draw of a
Joker game (shown above) will return 6 results, the last result is always the
joker number.

There are also three equivalent methods for the Propo games which return
results as a slice of strings.

```go
c := opap.NewClient(nil)

draw, _, err := c.Draws.PropoLatest(opap.PropoWed)
// ...

draw, _, err := c.Draws.PropoByNumber(opap.PropoSat, 201751)
// ...

draws, _, err := c.Draws.PropoByDate(opap.PropoSun, 17, 12, 2017)
// ...
```

If you need more control, when creating a new client you can pass an
http.Client as an argument.

For example this http.Client passed to the opap client will make sure to cancel
any request that takes longer than 1 second:

```go
httpcl := &http.Client{
	Timeout: 1 * time.Second,
}
c := opap.NewClient(httpcl)
// ...
```

Unit Testing
------------

To run all unit tests:

	cd $GOPATH/src/github.com/nstratos/go-opap/opap
	go test -cover

To see test coverage in your browser:

	go test -covermode=count -coverprofile=count.out && go tool cover -html count.out

License
-------

MIT
