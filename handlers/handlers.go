package handlers

import (
	"net/http"

	"github.com/prateekjoshi2013/scotch"
)

type Handlers struct {
	App *scotch.Scotch
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "home", nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
