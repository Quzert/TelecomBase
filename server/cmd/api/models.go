package main

import (
	"net/http"
)

type modelListItem struct {
	Id         int64  `json:"id"`
	VendorId   int64  `json:"vendorId"`
	VendorName string `json:"vendorName"`
	Name       string `json:"name"`
	DeviceType string `json:"deviceType"`
}

func (a *app) handleModelsList(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(
		r.Context(),
		"SELECT m.id, m.vendor_id, v.name, m.name, COALESCE(m.device_type, '') FROM models m JOIN vendors v ON v.id = m.vendor_id ORDER BY v.name, m.name",
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}
	defer rows.Close()

	items := make([]modelListItem, 0)
	for rows.Next() {
		var it modelListItem
		if err := rows.Scan(&it.Id, &it.VendorId, &it.VendorName, &it.Name, &it.DeviceType); err != nil {
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
