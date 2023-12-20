package main

import (
	"log"
	"os"

	"github.com/prateekjoshi2013/scotch"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	scotch := &scotch.Scotch{}
	err = scotch.New(path)
	if err != nil {
		log.Fatal(err)
	}

	scotch.AppName = "myapp"
	scotch.InfoLog.Println("Debug is set to ", scotch.Debug)
	app := &application{
		App: scotch,
	}
	return app
}
