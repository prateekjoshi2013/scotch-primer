package main

import (
	"github.com/prateekjoshi2013/scotch"
	"github.com/prateekjoshi2013/scotch-primer/handlers"
)

type application struct {
	App      *scotch.Scotch
	Handlers *handlers.Handlers
}

func main() {
	s := initApplication()
	s.App.ListenAndServe()
}
