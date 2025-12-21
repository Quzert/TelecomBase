package main

import (
	"context"
	"log"
)

func (a *app) seedIfEmpty(ctx context.Context) {
	// Seed minimal reference data so the demo works out-of-the-box.
	// Safe to run multiple times: only seeds if tables are empty.

	var vendorsCount int
	if err := a.db.QueryRow(ctx, "SELECT COUNT(*) FROM vendors").Scan(&vendorsCount); err != nil {
		log.Printf("seed vendors count: %v", err)
		return
	}
	if vendorsCount == 0 {
		_, err := a.db.Exec(ctx, "INSERT INTO vendors(name, country) VALUES($1, $2)", "Cisco", "US")
		if err != nil {
			log.Printf("seed vendors: %v", err)
			return
		}
	}

	var locationsCount int
	if err := a.db.QueryRow(ctx, "SELECT COUNT(*) FROM locations").Scan(&locationsCount); err != nil {
		log.Printf("seed locations count: %v", err)
		return
	}
	if locationsCount == 0 {
		_, err := a.db.Exec(ctx, "INSERT INTO locations(name, note) VALUES($1, $2)", "Main office", "Default location")
		if err != nil {
			log.Printf("seed locations: %v", err)
			return
		}
	}

	var modelsCount int
	if err := a.db.QueryRow(ctx, "SELECT COUNT(*) FROM models").Scan(&modelsCount); err != nil {
		log.Printf("seed models count: %v", err)
		return
	}
	if modelsCount == 0 {
		var vendorId int64
		if err := a.db.QueryRow(ctx, "SELECT id FROM vendors ORDER BY id LIMIT 1").Scan(&vendorId); err != nil {
			log.Printf("seed models vendor: %v", err)
			return
		}
		_, err := a.db.Exec(ctx, "INSERT INTO models(vendor_id, name, device_type) VALUES($1, $2, $3)", vendorId, "ISR 4321", "router")
		if err != nil {
			log.Printf("seed models: %v", err)
			return
		}
	}
}
