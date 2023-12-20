package main

import (
	"log"
	"os"

	"github.com/prateekjoshi2013/scotch"
	"github.com/prateekjoshi2013/scotch-primer/handlers"
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

	handlers := &handlers.Handlers{
		App: scotch,
	}

	scotch.InfoLog.Println("Debug is set to ", scotch.Debug)

	app := &application{
		App:      scotch,
		Handlers: handlers,
	}

	app.App.Routes = app.routes()

	return app
}
