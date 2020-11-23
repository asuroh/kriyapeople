package usecase

import (
	"errors"
	"kriyapeople/helper"
	"kriyapeople/pkg/amqp"
	"kriyapeople/pkg/logruslogger"
	"kriyapeople/server/request"
	"kriyapeople/usecase/viewmodel"
)

var (
	// VerifyKey ...
	VerifyKey = "verify_key"
	// UserLocked ...
	UserLocked = "user_locked"
)

// AdminResetPasswordUC ...
type AdminResetPasswordUC struct {
	*ContractUC
}

// ReqResetPassword ...
func (uc AdminResetPasswordUC) ReqResetPassword(data request.ForgotPasswordRequest) (err error) {
	ctx := "AdminResetPasswordUC.ReqResetPassword"

	adminUc := AdminUC{ContractUC: uc.ContractUC}
	admin, err := adminUc.FindByEmail(data.Email, false)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return err
	}

	// Push data to mqueue
	mqueue := amqp.NewQueue(AmqpConnection, AmqpChannel)
	queueBody := map[string]interface{}{
		"qid":       uc.ContractUC.ReqID,
		"id":        admin.ID,
		"name":      admin.UserName,
		"email":     admin.Email,
		"user_type": "admin",
	}
	AmqpConnection, AmqpChannel, err = mqueue.PushQueueReconnect(uc.ContractUC.EnvConfig["AMQP_URL"], queueBody, amqp.ResetPasswordMail, amqp.ResetPasswordMailDeadLetter)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_cash_request_amqp", uc.ReqID)
		return errors.New(helper.InternalServer)
	}

	return err
}

// GetTokenByKey ...
func (uc AdminResetPasswordUC) GetTokenByKey(key string) (res viewmodel.JwtVM, err error) {
	ctx := "AdminResetPasswordUC.GetTokenByKey"
	var (
		id       string
		validKey string
	)

	// Get key
	err = uc.GetFromRedis("resetPassword"+key, &id)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "get_key", uc.ReqID)
		return res, err
	}

	// Check valid key
	err = uc.GetFromRedis("resetPasswordKey"+id, &validKey)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "get_valid_key", uc.ReqID)
		return res, err
	}
	if validKey != key {
		logruslogger.Log(logruslogger.WarnLevel, validKey+"!="+key, ctx, "invalid_key", uc.ReqID)
		return res, errors.New(helper.ExpKey)
	}

	adminUc := AdminUC{ContractUC: uc.ContractUC}
	_, err = adminUc.FindByID(id, false)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return res, err
	}

	payload := map[string]interface{}{
		"id":   id,
		"role": "admin",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "jwt", uc.ReqID)
		return res, err
	}

	uc.StoreToRedisExp("latestAction"+id, VerifyKey, "24h")
	res.LatestAction = VerifyKey

	return res, err
}

// NewPasswordSubmit ...
func (uc AdminResetPasswordUC) NewPasswordSubmit(userID, password string) (err error) {
	ctx := "AdminResetPasswordUC.NewPasswordSubmit"

	// Decrypt input
	password, err = uc.AesFront.Decrypt(password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt", uc.ReqID)
		return errors.New(helper.InternalServer)
	}

	// Check password format
	err = helper.CheckPassword(password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "check_password", uc.ReqID)
		return err
	}

	// Update um password
	adminUc := AdminUC{ContractUC: uc.ContractUC}
	_, err = adminUc.UpdatePassword(userID, password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "um_reset_password", uc.ReqID)
		return err
	}

	uc.StoreToRedisExp("latestAction"+userID, UserLocked, "24h")
	uc.RemoveFromRedis("resetPasswordKey" + userID)

	return err
}
