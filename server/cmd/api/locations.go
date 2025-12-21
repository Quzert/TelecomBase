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
	rows, err := a.db.Query(r.Context(), "SELECT id, name, COALESCE(note, '') FROM locations ORDER BY name")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]locationListItem, 0)
	for rows.Next() {
		var it locationListItem
		if err := rows.Scan(&it.Id, &it.Name, &it.Note); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
			return
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	writeJSON(w, http.StatusOK, items)
}
