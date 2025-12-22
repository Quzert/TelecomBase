package main

import (
	"context"
	"log"
)

func (a *app) seedIfEmpty(ctx context.Context) {
	// Минимальные справочные данные, чтобы демо работало «из коробки».
	// Идемпотентно: заполняет только если таблицы пустые.

	vendorsCount, err := a.st.CountVendors(ctx)
	if err != nil {
		log.Printf("seed vendors count: %v", err)
		return
	}
	if vendorsCount == 0 {
		if _, err := a.st.CreateVendor(ctx, "Cisco", "US"); err != nil {
			log.Printf("seed vendors: %v", err)
			return
		}
	}

	locationsCount, err := a.st.CountLocations(ctx)
	if err != nil {
		log.Printf("seed locations count: %v", err)
		return
	}
	if locationsCount == 0 {
		if _, err := a.st.CreateLocation(ctx, "Main office", "Default location"); err != nil {
			log.Printf("seed locations: %v", err)
			return
		}
	}

	modelsCount, err := a.st.CountModels(ctx)
	if err != nil {
		log.Printf("seed models count: %v", err)
		return
	}
	if modelsCount == 0 {
		vendorId, err := a.st.GetFirstVendorID(ctx)
		if err != nil {
			log.Printf("seed models vendor: %v", err)
			return
		}
		if _, err := a.st.CreateModel(ctx, vendorId, "ISR 4321", "router"); err != nil {
			log.Printf("seed models: %v", err)
			return
		}
	}
}
