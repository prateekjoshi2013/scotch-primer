package main

import "github.com/prateekjoshi2013/scotch"

type application struct {
	App *scotch.Scotch
}

func main() {
	s := initApplication()
	s.App.ListenAndServe()
}
