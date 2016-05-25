package webserver

import (
	"fmt"
	"github.com/gocraft/web"
	"net/http"
	"strconv"
	"strings"
	//"os"
	//"path"
)

type Context struct {
	HelloCount int
	Thing      string
}

func (c *Context) SetHelloCount(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	var thing string = req.PathParams["count"]

	c.Thing = thing
	next(rw, req)
}

func (c *Context) SayHello(rw web.ResponseWriter, req *web.Request) {
	var numHellos, err = strconv.Atoi(req.PathParams["count"])
	if err == nil {
		c.HelloCount = numHellos
	}
	fmt.Fprint(rw, strings.Repeat("Hello ", c.HelloCount), "There!")
}

func StartServer() {
	//currentRoot, _ := os.Getwd()

	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).     // Use some included middleware
					Middleware(web.ShowErrorsMiddleware). // ...
					Middleware(web.StaticMiddleware("./static", web.StaticOption{IndexFile: "index.html"})).
					Middleware((*Context).SetHelloCount).     // Your own middleware!
					Get("/hello/:count", (*Context).SayHello) // Add a route

	http.ListenAndServe("localhost:8000", router) // Start the server!
}
