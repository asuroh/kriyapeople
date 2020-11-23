package usecase

import (
	"errors"
	"kriyapeople/helper"
	"kriyapeople/model"
	"kriyapeople/pkg/logruslogger"
	"kriyapeople/server/request"
	"kriyapeople/usecase/viewmodel"
	"time"
)

// UserDeviceTokenUC ...
type UserDeviceTokenUC struct {
	*ContractUC
}

// BuildBody ...
func (uc UserDeviceTokenUC) BuildBody(data *model.UserDeviceTokenEntity, res *viewmodel.UserDeviceTokenVM) {
	res.ID = data.ID
	res.UserID = data.UserID.String
	res.Type = data.Type.String
	res.Token = data.Token.String
	res.CreatedAt = data.CreatedAt.String
	res.UpdatedAt = data.UpdatedAt.String
	res.DeletedAt = data.DeletedAt.String
}

// SelectAll ...
func (uc UserDeviceTokenUC) SelectAll(userID, types string) (res []viewmodel.UserDeviceTokenVM, err error) {
	ctx := "UserDeviceTokenUC.SelectAll"

	m := model.NewUserDeviceTokenModel(uc.DB)
	data, err := m.SelectAll(userID, types)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	for _, r := range data {
		temp := viewmodel.UserDeviceTokenVM{}
		uc.BuildBody(&r, &temp)
		res = append(res, temp)
	}

	return res, err
}

// FindByUserToken ...
func (uc UserDeviceTokenUC) FindByUserToken(userID, types, token string) (res viewmodel.UserDeviceTokenVM, err error) {
	ctx := "UserDeviceTokenUC.FindByUserToken"

	m := model.NewUserDeviceTokenModel(uc.DB)
	data, err := m.FindByUserToken(userID, types, token)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, errors.New(helper.RecordNotExist)
	}
	uc.BuildBody(&data, &res)

	return res, err
}

// Create ...
func (uc UserDeviceTokenUC) Create(userID string, data request.UserDeviceTokenRequest) (res viewmodel.UserDeviceTokenVM, err error) {
	ctx := "UserDeviceTokenUC.Create"

	oldData, _ := uc.FindByUserToken(userID, data.Type, data.Token)
	if oldData.ID != "" {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "exist", uc.ReqID)
		return res, errors.New(helper.RecordExist)
	}

	now := time.Now().UTC()
	res = viewmodel.UserDeviceTokenVM{
		UserID: userID,
		Type:   data.Type,
		Token:  data.Token,
	}

	m := model.NewUserDeviceTokenModel(uc.DB)
	res.ID, err = m.Store(res, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}

// Delete ...
func (uc UserDeviceTokenUC) Delete(id string) (res viewmodel.UserDeviceTokenVM, err error) {
	ctx := "UserDeviceTokenUC.Delete"

	now := time.Now().UTC()
	m := model.NewUserDeviceTokenModel(uc.DB)
	res.ID, err = m.Destroy(id, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}

// DeleteByUserToken ...
func (uc UserDeviceTokenUC) DeleteByUserToken(userID, types, token string) (res viewmodel.UserDeviceTokenVM, err error) {
	ctx := "UserDeviceTokenUC.DeleteByUserToken"

	now := time.Now().UTC()
	m := model.NewUserDeviceTokenModel(uc.DB)
	res.ID, err = m.DestroyByUserToken(userID, types, token, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}
