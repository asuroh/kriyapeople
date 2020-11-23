package model

import (
	"database/sql"
	"kriyapeople/pkg/interfacepkg"
	"kriyapeople/pkg/str"
	"kriyapeople/usecase/viewmodel"
	"strings"
	"time"
)

var (
	// UserRegisterTypeEmail ...
	UserRegisterTypeEmail = "email"
	// UserRegisterTypeFacebook ...
	UserRegisterTypeFacebook = "facebook"
	// UserRegisterTypeGmail ...
	UserRegisterTypeGmail = "gmail"
	// UserRegisterTypeApple ...
	UserRegisterTypeApple = "apple"

	// DefaultUserBy ...
	DefaultUserBy = "def.updated_at"
	// UserBy ...
	UserBy = []string{
		"def.created_at", "def.updated_at", "def.email", "def.name", "def.phone", "def.register_type", "def.status",
	}

	userSelectString = `SELECT def."id", def."email", def."email_valid_at", def."name", def."phone",
		def."profile_image_id", def."password", def."register_type", def."register_detail", def."status",
		def."created_at", def."updated_at", def."deleted_at", profile_image."url"
		FROM "users" def
		LEFT JOIN "files" profile_image ON profile_image."id" = def."profile_image_id"`
)

func (model userModel) scanRows(rows *sql.Rows) (d UserEntity, err error) {
	err = rows.Scan(
		&d.ID, &d.Email, &d.EmailValidAt, &d.Name, &d.Phone, &d.ProfileImageID, &d.Password,
		&d.RegisterType, &d.RegisterDetail, &d.Status, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		&d.ProfileImage.URL,
	)

	return d, err
}

func (model userModel) scanRow(row *sql.Row) (d UserEntity, err error) {
	err = row.Scan(
		&d.ID, &d.Email, &d.EmailValidAt, &d.Name, &d.Phone, &d.ProfileImageID, &d.Password,
		&d.RegisterType, &d.RegisterDetail, &d.Status, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		&d.ProfileImage.URL,
	)

	return d, err
}

// userModel ...
type userModel struct {
	DB *sql.DB
}

// IUser ...
type IUser interface {
	SelectAll(search, status, by, sort string) ([]UserEntity, error)
	FindAll(search, status string, offset, limit int, by, sort string) ([]UserEntity, int, error)
	FindByID(id string) (UserEntity, error)
	FindByEmail(email string) (UserEntity, error)
	Store(body viewmodel.UserVM, changedAt time.Time) (string, error)
	Update(id string, body viewmodel.UserVM, changedAt time.Time) (string, error)
	UpdateProfile(id string, body viewmodel.UserVM, changedAt time.Time) (string, error)
	UpdatePassword(id, password string, changedAt time.Time) (string, error)
	UpdateEmailValidAt(id string, changedAt time.Time) (string, error)
	Destroy(id string, changedAt time.Time) (string, error)
}

// UserEntity ...
type UserEntity struct {
	ID             string         `db:"id"`
	Email          sql.NullString `db:"email"`
	EmailValidAt   sql.NullString `db:"email_valid_at"`
	Name           sql.NullString `db:"name"`
	Phone          sql.NullString `db:"phone"`
	ProfileImageID sql.NullString `db:"profile_image_id"`
	ProfileImage   FileEntity     `db:"profile_image"`
	Password       sql.NullString `db:"password"`
	RegisterType   sql.NullString `db:"register_type"`
	RegisterDetail sql.NullString `db:"register_detail"`
	Status         sql.NullBool   `db:"status"`
	CreatedAt      string         `db:"created_at"`
	UpdatedAt      string         `db:"updated_at"`
	DeletedAt      sql.NullString `db:"deleted_at"`
}

// NewUserModel ...
func NewUserModel(db *sql.DB) IUser {
	return &userModel{DB: db}
}

// SelectAll ...
func (model userModel) SelectAll(search, status, by, sort string) (res []UserEntity, err error) {
	conditionString := ``
	if status != "" {
		conditionString += ` AND def."status" = ` + str.StringToBoolString(status)
	}

	query := userSelectString + ` WHERE def."deleted_at" IS NULL AND (
		LOWER(def."email") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."phone") LIKE $1
		OR LOWER(def."register_type") LIKE $1
	) ` + conditionString + ` ORDER BY ` + by + ` ` + sort
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
func (model userModel) FindAll(search, status string, offset, limit int, by, sort string) (res []UserEntity, count int, err error) {
	conditionString := ``
	if status != "" {
		conditionString += ` AND def."status" = ` + str.StringToBoolString(status)
	}

	query := userSelectString + ` WHERE def."deleted_at" IS NULL AND (
		LOWER(def."email") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."phone") LIKE $1
		OR LOWER(def."register_type") LIKE $1
	) ` + conditionString + ` ORDER BY ` + by + ` ` + sort + ` OFFSET $2 LIMIT $3`
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

	query = `SELECT COUNT(def."id") FROM "users" def
		WHERE def."deleted_at" IS NULL AND (
			LOWER(def."email") LIKE $1 OR LOWER(def."name") LIKE $1 OR LOWER(def."phone") LIKE $1
			OR LOWER(def."register_type") LIKE $1
		) ` + conditionString
	err = model.DB.QueryRow(query, `%`+strings.ToLower(search)+`%`).Scan(&count)

	return res, count, err
}

// FindByID ...
func (model userModel) FindByID(id string) (res UserEntity, err error) {
	query := userSelectString + ` WHERE def."deleted_at" IS NULL AND def."id" = $1
		ORDER BY def."created_at" DESC LIMIT 1`
	row := model.DB.QueryRow(query, id)
	res, err = model.scanRow(row)

	return res, err
}

// FindByEmail ...
func (model userModel) FindByEmail(email string) (res UserEntity, err error) {
	query := userSelectString + ` WHERE def."deleted_at" IS NULL AND LOWER(def."email") = $1
		ORDER BY def."created_at" DESC LIMIT 1`
	row := model.DB.QueryRow(query, strings.ToLower(email))
	res, err = model.scanRow(row)

	return res, err
}

// Store ...
func (model userModel) Store(body viewmodel.UserVM, changedAt time.Time) (res string, err error) {
	sql := `INSERT INTO "users" (
		"email", "email_valid_at", "name", "phone", "profile_image_id", "password", "register_type", 
		"register_detail", "status", "created_at", "updated_at"
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10) RETURNING "id"`
	err = model.DB.QueryRow(sql,
		body.Email, str.EmptyString(body.EmailValidAt), body.Name, body.Phone,
		str.EmptyString(body.ProfileImageID), str.EmptyString(body.Password), body.RegisterType,
		interfacepkg.Marshall(body.RegisterDetail), body.Status, changedAt,
	).Scan(&res)

	return res, err
}

// Update ...
func (model userModel) Update(id string, body viewmodel.UserVM, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "users" SET "email" = $1, "email_valid_at" = $2, "name" = $3, "phone" = $4,
		"profile_image_id" = $5, "password" = $6, "register_type" = $7, "register_detail" = $8,
		"status" = $9, "updated_at" = $10 WHERE "deleted_at" IS NULL
		AND "id" = $11 RETURNING "id"`
	err = model.DB.QueryRow(sql,
		body.Email, str.EmptyString(body.EmailValidAt), body.Name, body.Phone,
		str.EmptyString(body.ProfileImageID), str.EmptyString(body.Password), body.RegisterType,
		interfacepkg.Marshall(body.RegisterDetail), body.Status, changedAt, id,
	).Scan(&res)

	return res, err
}

// UpdateProfile ...
func (model userModel) UpdateProfile(id string, body viewmodel.UserVM, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "users" SET "name" = $1, "phone" = $2, "profile_image_id" = $3, "status" = $4,
		"updated_at" = $5 WHERE "deleted_at" IS NULL AND "id" = $6 RETURNING "id"`
	err = model.DB.QueryRow(sql,
		body.Name, body.Phone, str.EmptyString(body.ProfileImageID), body.Status, changedAt, id,
	).Scan(&res)

	return res, err
}

// UpdatePassword ...
func (model userModel) UpdatePassword(id, password string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "users" SET "password" = $1, "updated_at" = $2 WHERE "deleted_at" IS NULL
		AND "id" = $3 RETURNING "id"`
	err = model.DB.QueryRow(sql, password, changedAt, id).Scan(&res)

	return res, err
}

// UpdateEmailValidAt ...
func (model userModel) UpdateEmailValidAt(id string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "users" SET "email_valid_at" = $1, "updated_at" = $1 WHERE "deleted_at" IS NULL
		AND "id" = $2 RETURNING "id"`
	err = model.DB.QueryRow(sql, changedAt, id).Scan(&res)

	return res, err
}

// Destroy ...
func (model userModel) Destroy(id string, changedAt time.Time) (res string, err error) {
	sql := `UPDATE "users" SET "updated_at" = $1, "deleted_at" = $1
		WHERE "deleted_at" IS NULL AND "id" = $2 RETURNING "id"`
	err = model.DB.QueryRow(sql, changedAt, id).Scan(&res)

	return res, err
}
