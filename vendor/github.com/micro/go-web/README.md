# Go Web [![GoDoc](https://godoc.org/github.com/micro/go-web?status.svg)](https://godoc.org/github.com/micro/go-web) [![Travis CI](https://travis-ci.org/micro/go-web.svg?branch=master)](https://travis-ci.org/micro/go-web) [![Go Report Card](https://goreportcard.com/badge/micro/go-web)](https://goreportcard.com/report/github.com/micro/go-web)

**Go-web** is a tiny HTTP web server library which leverages [go-micro](https://github.com/micro/go-micro) to create 
micro web services as first class citizens in a microservice ecosystem. It's merely a wrapper around registration, 
heartbeating and initialization of the go-micro client. In the future go-platform features may be included.

## Getting Started

### Prerequisites

Go-web uses a similar pattern to go-micro. Look at the go-micro [readme](https://github.com/micro/go-micro) for 
starting up the registry.

### Usage

```go
service := web.NewService(
	web.Name("go.micro.web.example"),
	web.Version("latest"),
)

service.HandleFunc("/foo", fooHandler)

if err := service.Init(); err != nil {
	log.Fatal(err)
}

if err := service.Run(); err != nil {
	log.Fatal(err)
}
```

### Use your own Handler

You might have a preference for a HTTP handler, so use something else. This loses the ability to register endpoints in discovery 
but we'll fix that soon.

```go
import "github.com/gorilla/mux"

r := mux.NewRouter()
r.HandleFunc("/", indexHandler)
r.HandleFunc("/objects/{object}", objectHandler)

service := web.NewService(
	web.Handler(r)
)
```
