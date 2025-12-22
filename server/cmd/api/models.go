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
	rows, err := a.st.ListModels(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "db_error"})
		return
	}

	items := make([]modelListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, modelListItem{
			Id:         row.ID,
			VendorId:   row.VendorID,
			VendorName: row.VendorName,
			Name:       row.Name,
			DeviceType: row.DeviceType,
		})
	}

	writeJSON(w, http.StatusOK, items)
}
