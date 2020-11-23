package model

import (
	"database/sql"
	"kriyapeople/usecase/viewmodel"
	"strings"
	"time"
)

var (
	// DefaultAdminBy ...
	DefaultAdminBy = "def.updated_at"
	// AdminBy ...
	AdminBy = []string{
		"def.created_at", "def.updated_at",
	}

	adminSelectString = `SELECT def.id, def."data" ->> 'email' as email, def."data" ->> 'password' as password, def."data" ->> 'username' as username, def."data" -> 'status' ->> 'is_active' as status, r."id" as role_id, r."data" ->> 'role_name' as role_name, def.created_at, def.updated_at, def.deleted_at FROM "users" def LEFT JOIN "roles" r ON r."id" = def."role_id"`
)

func (model adminModel) scanRows(rows *sql.Rows) (d AdminEntity, err error) {
	err = rows.Scan(
		&d.ID, &d.Email, &d.Password, &d.UserName, &d.Status, &d.RoleID, &d.Role.Name, &d.CreatedAt,
		&d.UpdatedAt, &d.DeletedAt,
	)

	return d, err
}

func (model adminModel) scanRow(row *sql.Row) (d AdminEntity, err error) {
	err = row.Scan(
		&d.ID, &d.Email, &d.Password, &d.UserName, &d.Status, &d.RoleID, &d.Role.Name, &d.CreatedAt,
		&d.UpdatedAt, &d.DeletedAt,
	)

	return d, err
}

// adminModel ...
type adminModel struct {
	DB *sql.DB
}

// IAdmin ...
type IAdmin interface {
	SelectAll(search, by, sort string) ([]AdminEntity, error)
	FindAll(search string, offset, limit int, by, sort string) ([]AdminEntity, int, error)
	FindByID(id string) (AdminEntity, error)
	FindByCode(code string) (AdminEntity, error)
	FindByEmail(email string) (AdminEntity, error)
	Store(body viewmodel.AdminVM, changedAt time.Time) (string, error)
	Update(id string, body viewmodel.AdminVM, changedAt time.Time) (string, error)
	UpdatePassword(id, password string, changedAt time.Time) (string, error)
	Destroy(id string, changedAt time.Time) (string, error)
}

// AdminEntity ....
type AdminEntity struct {
	ID        string         `db:"id"`
	Email     sql.NullString `db:"email"`
	Password  sql.NullString `db:"password"`
	UserName  sql.NullString `db:"user_name"`
	RoleID    sql.NullString `db:"role_id"`
	Role      RoleEntity     `db:"role"`
	Status    sql.NullBool   `db:"status"`
	CreatedAt string         `db:"created_at"`
	UpdatedAt string         `db:"updated_at"`
	DeletedAt sql.NullString `db:"deleted_at"`
}

// NewAdminModel ...
func NewAdminModel(db *sql.DB) IAdmin {
	return &adminModel{DB: db}
}

// SelectAll ...
func (model adminModel) SelectAll(search, by, sort string) (res []AdminEntity, err error) {
	query := adminSelectString + ` WHERE def."deleted_at" IS NULL AND (
		LOWER(def."code") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."email") LIKE $1
		OR LOWER(r."code") LIKE $1 OR LOWER(r."name") LIKE $1
	) ORDER BY ` + by + ` ` + sort
	rows, err := model.DB.Query(query, `%`+strings.ToLower(search)+`%`)
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

// FindAll ...
func (model adminModel) FindAll(search string, offset, limit int, by, sort string) (res []AdminEntity, count int, err error) {
	query := adminSelectString + ` WHERE def."deleted_at" IS NULL AND (
		LOWER(def."code") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."email") LIKE $1
		OR LOWER(r."code") LIKE $1 OR LOWER(r."name") LIKE $1
	) ORDER BY ` + by + ` ` + sort + ` OFFSET $2 LIMIT $3`
	rows, err := model.DB.Query(query, `%`+strings.ToLower(search)+`%`, offset, limit)
	if err != nil {
		return res, count, err
	}
	defer rows.Close()

	for rows.Next() {
		d, err := model.scanRows(rows)
		if err != nil {
			return res, count, err
		}
		res = append(res, d)
	}
	err = rows.Err()
	if err != nil {
		return res, count, err
	}

	query = `SELECT COUNT(def."id") FROM "admins" def
		LEFT JOIN "roles" r ON r."id" = def."role_id"
		WHERE def."deleted_at" IS NULL AND (
			LOWER(def."code") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."email") LIKE $1
			OR LOWER(r."code") LIKE $1 OR LOWER(r."name") LIKE $1
		)`
	err = model.DB.QueryRow(query, `%`+strings.ToLower(search)+`%`).Scan(&count)

	return res, count, err
}

// FindByID ...
func (model adminModel) FindByID(id string) (res AdminEntity, err error) {
	query := adminSelectString + ` WHERE def."deleted_at" IS NULL AND def."id" = $1
		ORDER BY def."created_at" DESC LIMIT 1`
	row := model.DB.QueryRow(query, id)
	res, err = model.scanRow(row)

	return res, err
}

// FindByCode ...
func (model adminModel) FindByCode(code string) (res AdminEntity, err error) {
	query := adminSelectString + ` WHERE def."deleted_at" IS NULL AND def."code" = $1
		ORDER BY def."created_at" DESC LIMIT 1`
	row := model.DB.QueryRow(query, code)
	res, err = model.scanRow(row)

	return res, err
}

// FindByEmail ...
func (model adminModel) FindByEmail(email string) (res AdminEntity, err error) {
	query := adminSelectString + ` WHERE def."deleted_at" IS NULL  AND LOWER (def."data" ->> 'email' ) = $1 ORDER BY def."created_at" DESC  LIMIT 1`
	row := model.DB.QueryRow(query, strings.ToLower(email))
	res, err = model.scanRow(row)

	return res, err
}

// Store ...
func (model adminModel) Store(body viewmodel.AdminVM, changedAt time.Time) (res string, err error) {
	sql := `INSERT INTO "admins" (
			"code", "name", "email", "password", "role_id", "profile_image_id", "status", "created_at", "updated_at"
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $8) RETURNING "id"`
	err = model.DB.QueryRow(sql, body.UserName, body.Email, body.Password, body.RoleID,
		body.Status, changedAt,
	).Scan(&res)

	return res, err
}

// Update ...
func (model adminModel) Update(id string, body viewmodel.AdminVM, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "admins" SET "name" = $1, "email" = $2, "password" = $3, "role_id" = $4,
		"profile_image_id" = $5, "status" = $6, "updated_at" = $7 WHERE "deleted_at" IS NULL
		AND "id" = $8 RETURNING "id"`
	err = model.DB.QueryRow(sql,
		body.UserName, body.Email, body.Password, body.RoleID,
		body.Status, changedAt, id,
	).Scan(&res)

	return res, err
}

// UpdatePassword ...
func (model adminModel) UpdatePassword(id, password string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "admins" SET "password" = $1, "updated_at" = $2 WHERE "deleted_at" IS NULL
		AND "id" = $3 RETURNING "id"`
	err = model.DB.QueryRow(sql, password, changedAt, id).Scan(&res)

	return res, err
}

// Destroy ...
func (model adminModel) Destroy(id string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "admins" SET "updated_at" = $1, "deleted_at" = $1
		WHERE "deleted_at" IS NULL AND "id" = $2 RETURNING "id"`
	err = model.DB.QueryRow(sql, changedAt, id).Scan(&res)

	return res, err
}
