package usecase

import (
	"errors"
	"kriyapeople/helper"
	"kriyapeople/model"
	"kriyapeople/pkg/logruslogger"
	"kriyapeople/pkg/str"
	"kriyapeople/server/request"
	"kriyapeople/usecase/viewmodel"
	"strings"
	"time"
)

// AdminUC ...
type AdminUC struct {
	*ContractUC
}

// GenerateCode randomize code & check uniqueness from DB
func (uc AdminUC) GenerateCode() (res string, err error) {
	m := model.NewAdminModel(uc.DB)
	res = str.RandAlphanumericString(8)
	for {
		data, _ := m.FindByCode(res)
		if data.ID == "" {
			break
		}
		res = str.RandAlphanumericString(8)
	}

	return res, err
}

// BuildBody ...
func (uc AdminUC) BuildBody(data *model.AdminEntity, res *viewmodel.AdminVM, isShowPassword bool) {

	res.ID = data.ID
	res.UserName = data.UserName.String
	res.Email = data.Email.String
	res.Password = str.ShowString(isShowPassword, uc.Aes.DecryptNoErr(data.Password.String))
	res.RoleID = data.RoleID.String
	res.RoleName = data.Role.Name.String
	res.Status = data.Status.Bool
	res.CreatedAt = data.CreatedAt
	res.UpdatedAt = data.UpdatedAt
	res.DeletedAt = data.DeletedAt.String
}

// Login ...
func (uc AdminUC) Login(data request.AdminLoginRequest) (res viewmodel.JwtVM, err error) {
	ctx := "AdminUC.Login"

	if len(data.Password) < 8 {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "password_length", uc.ReqID)
		return res, errors.New(helper.InvalidCredentials)
	}

	admin, err := uc.FindByEmail(data.Email, true)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_by_email", uc.ReqID)
		return res, errors.New(helper.InvalidCredentials)
	}
	if admin.Password != data.Password {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "invalid_password", uc.ReqID)
		return res, errors.New(helper.InvalidCredentials)
	}
	if !admin.Status {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "inactive_admin", uc.ReqID)
		return res, errors.New(helper.InactiveAdmin)
	}

	// Jwe the payload & Generate jwt token
	payload := map[string]interface{}{
		"id":   admin.ID,
		"role": "admin",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "jwt", uc.ReqID)
		return res, errors.New(helper.InternalServer)
	}

	return res, err
}

// SelectAll ...
func (uc AdminUC) SelectAll(search, by, sort string) (res []viewmodel.AdminVM, err error) {
	ctx := "AdminUC.SelectAll"

	if !str.Contains(model.AdminBy, by) {
		by = model.DefaultAdminBy
	}
	if !str.Contains(SortWhitelist, strings.ToLower(sort)) {
		sort = DescSort
	}

	m := model.NewAdminModel(uc.DB)
	data, err := m.SelectAll(search, by, sort)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	for _, r := range data {
		temp := viewmodel.AdminVM{}
		uc.BuildBody(&r, &temp, false)
		res = append(res, temp)
	}

	return res, err
}

// FindAll ...
func (uc AdminUC) FindAll(search string, page, limit int, by, sort string) (res []viewmodel.AdminVM, pagination viewmodel.PaginationVM, err error) {
	ctx := "AdminUC.FindAll"

	if !str.Contains(model.AdminBy, by) {
		by = model.DefaultAdminBy
	}
	if !str.Contains(SortWhitelist, strings.ToLower(sort)) {
		sort = DescSort
	}

	limit = uc.LimitMax(limit)
	limit, offset := uc.PaginationPageOffset(page, limit)

	m := model.NewAdminModel(uc.DB)
	data, count, err := m.FindAll(search, offset, limit, by, sort)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, pagination, err
	}
	pagination = PaginationRes(page, count, limit)

	for _, r := range data {
		temp := viewmodel.AdminVM{}
		uc.BuildBody(&r, &temp, false)
		res = append(res, temp)
	}

	return res, pagination, err
}

// FindByID ...
func (uc AdminUC) FindByID(id string, isShowPassword bool) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.FindByID"

	m := model.NewAdminModel(uc.DB)
	data, err := m.FindByID(id)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}
	uc.BuildBody(&data, &res, isShowPassword)

	return res, err
}

// FindByCode ...
func (uc AdminUC) FindByCode(code string, isShowPassword bool) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.FindByCode"

	m := model.NewAdminModel(uc.DB)
	data, err := m.FindByCode(code)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}
	uc.BuildBody(&data, &res, isShowPassword)

	return res, err
}

// FindByEmail ...
func (uc AdminUC) FindByEmail(email string, isShowPassword bool) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.FindByEmail"

	m := model.NewAdminModel(uc.DB)
	data, err := m.FindByEmail(email)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	uc.BuildBody(&data, &res, isShowPassword)

	return res, err
}

// CheckDetails ...
func (uc AdminUC) CheckDetails(userID string, data *request.AdminRequest, oldData *viewmodel.AdminVM) (err error) {
	ctx := "AdminUC.CheckDetails"

	admin, _ := uc.FindByEmail(data.Email, false)
	if admin.ID != "" && admin.ID != oldData.ID {
		logruslogger.Log(logruslogger.WarnLevel, data.Email, ctx, "duplicate_email", uc.ReqID)
		return errors.New(helper.DuplicateEmail)
	}

	if data.Password == "" && oldData.Password == "" {
		logruslogger.Log(logruslogger.WarnLevel, data.Email, ctx, "empty_password", uc.ReqID)
		return errors.New(helper.InvalidPassword)
	}

	// Decrypt password input
	if data.Password != "" {
		data.Password, err = uc.AesFront.Decrypt(data.Password)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt", uc.ReqID)
			return err
		}
	} else {
		data.Password = oldData.Password
	}

	// Encrypt password
	data.Password, err = uc.Aes.Encrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return err
	}

	// Generate code
	data.Code, err = uc.GenerateCode()
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "generate_code", uc.ReqID)
		return err
	}

	return err
}

// Create ...
func (uc AdminUC) Create(userID string, data *request.AdminRequest) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.Create"

	err = uc.CheckDetails(userID, data, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "check_details", uc.ReqID)
		return res, err
	}

	now := time.Now().UTC()
	res = viewmodel.AdminVM{
		// UserName:  data.UserName,
		Email:     data.Email,
		Password:  data.Password,
		RoleID:    data.RoleID,
		Status:    data.Status,
		CreatedAt: now.Format(time.RFC3339),
		UpdatedAt: now.Format(time.RFC3339),
	}
	m := model.NewAdminModel(uc.DB)
	res.ID, err = m.Store(res, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}

// Update ...
func (uc AdminUC) Update(userID, id string, data *request.AdminRequest) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.Update"

	oldData, err := uc.FindByID(id, true)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return res, err
	}

	err = uc.CheckDetails(userID, data, &oldData)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "check_details", uc.ReqID)
		return res, err
	}

	now := time.Now().UTC()
	res = viewmodel.AdminVM{
		// Name:           data.Name,
		Email:     data.Email,
		Password:  data.Password,
		RoleID:    data.RoleID,
		Status:    data.Status,
		UpdatedAt: now.Format(time.RFC3339),
	}
	m := model.NewAdminModel(uc.DB)
	res.ID, err = m.Update(id, res, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}

// UpdatePassword ...
func (uc AdminUC) UpdatePassword(id, password string) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.UpdatePassword"

	// Encrypt password
	password, err = uc.Aes.Encrypt(password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return res, err
	}

	now := time.Now().UTC()
	adminModel := model.NewAdminModel(uc.DB)
	res.ID, err = adminModel.UpdatePassword(id, password, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}

// Delete ...
func (uc AdminUC) Delete(id string) (res viewmodel.AdminVM, err error) {
	ctx := "AdminUC.Delete"

	now := time.Now().UTC()
	m := model.NewAdminModel(uc.DB)
	res.ID, err = m.Destroy(id, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}
