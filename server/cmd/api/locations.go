package main

import (
	"net/http"
)

type locationListItem struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Note string `json:"note"`
}

func (a *app) handleLocationsList(w http.ResponseWriter, r *http.Request) {
	rows, err := a.st.ListLocations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]locationListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, locationListItem{Id: row.ID, Name: row.Name, Note: row.Note})
	}

	writeJSON(w, http.StatusOK, items)
}
