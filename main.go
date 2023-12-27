package main

import (
	"github.com/prateekjoshi2013/scotch"
	"github.com/prateekjoshi2013/scotch-primer/data"
	"github.com/prateekjoshi2013/scotch-primer/handlers"
	"github.com/prateekjoshi2013/scotch-primer/middlewares"
)

type application struct {
	App        *scotch.Scotch
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middlewares.Middleware
}

func main() {
	s := initApplication()
	s.App.ListenAndServe()
}
