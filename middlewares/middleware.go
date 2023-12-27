package middlewares

import (
	"github.com/prateekjoshi2013/scotch"
	"github.com/prateekjoshi2013/scotch-primer/data"
)

type Middleware struct {
	App    *scotch.Scotch
	Models data.Models
}


