package model

import (
	"database/sql"
	"kriyapeople/pkg/str"
	"kriyapeople/usecase/viewmodel"
	"time"
)

var (
	// UserDeviceTokenTypeFcm ...
	UserDeviceTokenTypeFcm = "fcm"
	// UserDeviceTokenTypeWhitelist ...
	UserDeviceTokenTypeWhitelist = []string{
		UserDeviceTokenTypeFcm,
	}
)

// userDeviceTokenModel ...
type userDeviceTokenModel struct {
	DB *sql.DB
}

// IUserDeviceToken ...
type IUserDeviceToken interface {
	SelectAll(userID, types string) ([]UserDeviceTokenEntity, error)
	FindByUserToken(userID, types, token string) (UserDeviceTokenEntity, error)
	Store(body viewmodel.UserDeviceTokenVM, changedAt time.Time) (string, error)
	Destroy(id string, changedAt time.Time) (string, error)
	DestroyByUserToken(userID, types, token string, changedAt time.Time) (string, error)
}

// UserDeviceTokenEntity ....
type UserDeviceTokenEntity struct {
	ID        string         `db:"id"`
	UserID    sql.NullString `db:"user_id"`
	Type      sql.NullString `db:"type"`
	Token     sql.NullString `db:"token"`
	CreatedAt sql.NullString `db:"created_at"`
	UpdatedAt sql.NullString `db:"updated_at"`
	DeletedAt sql.NullString `db:"deleted_at"`
}

// NewUserDeviceTokenModel ...
func NewUserDeviceTokenModel(db *sql.DB) IUserDeviceToken {
	return &userDeviceTokenModel{DB: db}
}

var (
	userDeviceTokenSelectString = `SELECT def."id", def."user_id", def."type", def."token",
		def."created_at", def."updated_at", def."deleted_at" FROM "user_device_tokens" def `
)

func (model userDeviceTokenModel) scanRows(rows *sql.Rows) (d UserDeviceTokenEntity, err error) {
	err = rows.Scan(
		&d.ID, &d.UserID, &d.Type, &d.Token, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)

	return d, err
}

func (model userDeviceTokenModel) scanRow(row *sql.Row) (d UserDeviceTokenEntity, err error) {
	err = row.Scan(
		&d.ID, &d.UserID, &d.Type, &d.Token, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)

	return d, err
}

// SelectAll ...
func (model userDeviceTokenModel) SelectAll(userID, types string) (res []UserDeviceTokenEntity, err error) {
	sub := ``
	if str.IsValidUUID(userID) {
		sub = sub + ` AND def."user_id" = '` + userID + `'`
	}
	if str.Contains(UserDeviceTokenTypeWhitelist, types) {
		sub = sub + ` AND def."type" = '` + types + `'`
	}

	query := userDeviceTokenSelectString + `WHERE def."deleted_at" IS NULL ` + sub + `
		ORDER BY def."created_at" ASC`
	rows, err := model.DB.Query(query)
	if err != nil {
		return res, err
	}

	defer rows.Close()
	for rows.Next() {
		d, err := model.scanRows(rows)
		if err != nil {
			return res, err
		}
		res = append(res, d)
	}
	err = rows.Err()

	return res, err
}

// FindByUserToken ...
func (model userDeviceTokenModel) FindByUserToken(userID, types, token string) (res UserDeviceTokenEntity, err error) {
	query := userDeviceTokenSelectString + ` WHERE def."deleted_at" IS NULL AND def."user_id" = $1
		AND def."type" = $2 AND def."token" = $3
		ORDER BY def."created_at" DESC LIMIT 1`

	row := model.DB.QueryRow(query, userID, types, token)
	res, err = model.scanRow(row)

	return res, err
}

// Store ...
func (model userDeviceTokenModel) Store(body viewmodel.UserDeviceTokenVM, changedAt time.Time) (res string, err error) {
	sql := `INSERT INTO "user_device_tokens" (
		"user_id", "type", "token", "created_at", "updated_at"
		) VALUES($1, $2, $3, $4, $4) RETURNING "id"`
	err = model.DB.QueryRow(sql, body.UserID, body.Type, body.Token, changedAt).Scan(&res)

	return res, err
}

// Destroy ...
func (model userDeviceTokenModel) Destroy(id string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "user_device_tokens" SET "updated_at" = $1, "deleted_at" = $1
		WHERE "deleted_at" IS NULL AND "id" = $2 RETURNING "id"`
	err = model.DB.QueryRow(sql, changedAt, id).Scan(&res)

	return res, err
}

// DestroyByUserToken ...
func (model userDeviceTokenModel) DestroyByUserToken(userID, types, token string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "user_device_tokens" SET "updated_at" = $1, "deleted_at" = $1
		WHERE "deleted_at" IS NULL AND "user_id" = $2 AND "type" = $3 AND "token" = $4 RETURNING "id"`
	err = model.DB.QueryRow(sql, changedAt, userID, types, token).Scan(&res)

	return res, err
}
