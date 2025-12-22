package store

import (
	"context"
	"time"
)

// Users

type CreateUserAutoAdminRow struct {
	Role     string
	Approved bool
}

type GetUserAuthByUsernameRow struct {
	PasswordHash string
	Role         string
	Approved     bool
}

type GetUserRoleApprovedByUsernameRow struct {
	Role     string
	Approved bool
}

type ListPendingUsersRow struct {
	ID        int64
	Username  string
	Role      string
	CreatedAt time.Time
}

type ListUsersRow struct {
	ID        int64
	Username  string
	Role      string
	Approved  bool
	CreatedAt time.Time
}

type GetUserRoleByIDRow struct {
	Role string
}

type GetUserUsernameRoleByIDRow struct {
	Username string
	Role     string
}

func (q *Queries) CreateUserAutoAdmin(ctx context.Context, username string, passwordHash string) (CreateUserAutoAdminRow, error) {
	row := q.db.QueryRow(ctx, sql("CreateUserAutoAdmin"), username, passwordHash)
	var out CreateUserAutoAdminRow
	err := row.Scan(&out.Role, &out.Approved)
	return out, err
}

func (q *Queries) GetUserAuthByUsername(ctx context.Context, username string) (GetUserAuthByUsernameRow, error) {
	row := q.db.QueryRow(ctx, sql("GetUserAuthByUsername"), username)
	var out GetUserAuthByUsernameRow
	err := row.Scan(&out.PasswordHash, &out.Role, &out.Approved)
	return out, err
}

func (q *Queries) GetUserRoleApprovedByUsername(ctx context.Context, username string) (GetUserRoleApprovedByUsernameRow, error) {
	row := q.db.QueryRow(ctx, sql("GetUserRoleApprovedByUsername"), username)
	var out GetUserRoleApprovedByUsernameRow
	err := row.Scan(&out.Role, &out.Approved)
	return out, err
}

func (q *Queries) ListPendingUsers(ctx context.Context) ([]ListPendingUsersRow, error) {
	rows, err := q.db.Query(ctx, sql("ListPendingUsers"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListPendingUsersRow
	for rows.Next() {
		var it ListPendingUsersRow
		if err := rows.Scan(&it.ID, &it.Username, &it.Role, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) ApproveUserByID(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("ApproveUserByID"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) ListUsers(ctx context.Context) ([]ListUsersRow, error) {
	rows, err := q.db.Query(ctx, sql("ListUsers"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListUsersRow
	for rows.Next() {
		var it ListUsersRow
		if err := rows.Scan(&it.ID, &it.Username, &it.Role, &it.Approved, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) GetUserRoleByID(ctx context.Context, id int64) (string, error) {
	row := q.db.QueryRow(ctx, sql("GetUserRoleByID"), id)
	var role string
	err := row.Scan(&role)
	return role, err
}

func (q *Queries) SetUserApprovedByID(ctx context.Context, id int64, approved bool) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("SetUserApprovedByID"), id, approved)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) GetUserUsernameRoleByID(ctx context.Context, id int64) (GetUserUsernameRoleByIDRow, error) {
	row := q.db.QueryRow(ctx, sql("GetUserUsernameRoleByID"), id)
	var out GetUserUsernameRoleByIDRow
	err := row.Scan(&out.Username, &out.Role)
	return out, err
}

func (q *Queries) DeleteUserByID(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("DeleteUserByID"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

// Vendors

type ListVendorsRow struct {
	ID      int64
	Name    string
	Country string
}

func (q *Queries) ListVendors(ctx context.Context) ([]ListVendorsRow, error) {
	rows, err := q.db.Query(ctx, sql("ListVendors"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListVendorsRow
	for rows.Next() {
		var it ListVendorsRow
		if err := rows.Scan(&it.ID, &it.Name, &it.Country); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) CreateVendor(ctx context.Context, name string, country any) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CreateVendor"), name, country)
	var id int64
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) UpdateVendor(ctx context.Context, id int64, name string, country any) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("UpdateVendor"), name, country, id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) DeleteVendor(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("DeleteVendor"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) CountVendors(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CountVendors"))
	var cnt int64
	err := row.Scan(&cnt)
	return cnt, err
}

func (q *Queries) GetFirstVendorID(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, sql("GetFirstVendorID"))
	var id int64
	err := row.Scan(&id)
	return id, err
}

// Models

type ListModelsRow struct {
	ID         int64
	VendorID   int64
	VendorName string
	Name       string
	DeviceType string
}

func (q *Queries) ListModels(ctx context.Context) ([]ListModelsRow, error) {
	rows, err := q.db.Query(ctx, sql("ListModels"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListModelsRow
	for rows.Next() {
		var it ListModelsRow
		if err := rows.Scan(&it.ID, &it.VendorID, &it.VendorName, &it.Name, &it.DeviceType); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) CreateModel(ctx context.Context, vendorID int64, name string, deviceType any) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CreateModel"), vendorID, name, deviceType)
	var id int64
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) UpdateModel(ctx context.Context, id int64, vendorID int64, name string, deviceType any) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("UpdateModel"), vendorID, name, deviceType, id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) DeleteModel(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("DeleteModel"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) CountModels(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CountModels"))
	var cnt int64
	err := row.Scan(&cnt)
	return cnt, err
}

// Locations

type ListLocationsRow struct {
	ID   int64
	Name string
	Note string
}

func (q *Queries) ListLocations(ctx context.Context) ([]ListLocationsRow, error) {
	rows, err := q.db.Query(ctx, sql("ListLocations"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListLocationsRow
	for rows.Next() {
		var it ListLocationsRow
		if err := rows.Scan(&it.ID, &it.Name, &it.Note); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) CreateLocation(ctx context.Context, name string, note any) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CreateLocation"), name, note)
	var id int64
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) UpdateLocation(ctx context.Context, id int64, name string, note any) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("UpdateLocation"), name, note, id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) DeleteLocation(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("DeleteLocation"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) CountLocations(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CountLocations"))
	var cnt int64
	err := row.Scan(&cnt)
	return cnt, err
}

// Devices

type ListDevicesRow struct {
	ID              int64
	VendorName      string
	ModelName       string
	LocationName    string
	SerialNumber    string
	InventoryNumber string
	Status          string
	InstalledAt     string
}

type GetDeviceByIDRow struct {
	ID              int64
	ModelID         int64
	LocationID      *int64
	SerialNumber    string
	InventoryNumber string
	Status          string
	InstalledAt     string
	Description     string
}

func (q *Queries) ListDevices(ctx context.Context, query string) ([]ListDevicesRow, error) {
	rows, err := q.db.Query(ctx, sql("ListDevices"), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListDevicesRow
	for rows.Next() {
		var it ListDevicesRow
		if err := rows.Scan(&it.ID, &it.VendorName, &it.ModelName, &it.LocationName, &it.SerialNumber, &it.InventoryNumber, &it.Status, &it.InstalledAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (q *Queries) CreateDevice(ctx context.Context, modelID int64, locationID *int64, serialNumber any, inventoryNumber any, status string, installedAt any, description any) (int64, error) {
	row := q.db.QueryRow(ctx, sql("CreateDevice"), modelID, locationID, serialNumber, inventoryNumber, status, installedAt, description)
	var id int64
	err := row.Scan(&id)
	return id, err
}

func (q *Queries) GetDeviceByID(ctx context.Context, id int64) (GetDeviceByIDRow, error) {
	row := q.db.QueryRow(ctx, sql("GetDeviceByID"), id)
	var out GetDeviceByIDRow
	err := row.Scan(&out.ID, &out.ModelID, &out.LocationID, &out.SerialNumber, &out.InventoryNumber, &out.Status, &out.InstalledAt, &out.Description)
	return out, err
}

func (q *Queries) UpdateDevice(ctx context.Context, id int64, modelID int64, locationID *int64, serialNumber any, inventoryNumber any, status string, installedAt any, description any) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("UpdateDevice"), modelID, locationID, serialNumber, inventoryNumber, status, installedAt, description, id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (q *Queries) DeleteDevice(ctx context.Context, id int64) (int64, error) {
	cmd, err := q.db.Exec(ctx, sql("DeleteDevice"), id)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}
