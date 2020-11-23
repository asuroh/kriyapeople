package usecase

import (
	"errors"
	"kriyapeople/helper"
	"kriyapeople/model"
	"kriyapeople/pkg/amqp"
	"kriyapeople/pkg/logruslogger"
	"kriyapeople/server/request"
	"kriyapeople/usecase/viewmodel"
)

// UserResetPasswordUC ...
type UserResetPasswordUC struct {
	*ContractUC
}

// ReqResetPassword ...
func (uc UserResetPasswordUC) ReqResetPassword(data request.ForgotPasswordRequest) (err error) {
	ctx := "UserResetPasswordUC.ReqResetPassword"

	userUc := UserUC{ContractUC: uc.ContractUC}
	user, err := userUc.FindByEmail(data.Email, false)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return err
	}
	if user.RegisterType != model.UserRegisterTypeEmail {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "register_type", uc.ReqID)
		return errors.New(helper.InvalidRegisterType)
	}

	// Push data to mqueue
	mqueue := amqp.NewQueue(AmqpConnection, AmqpChannel)
	queueBody := map[string]interface{}{
		"qid":       uc.ContractUC.ReqID,
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"user_type": "user",
	}
	AmqpConnection, AmqpChannel, err = mqueue.PushQueueReconnect(uc.ContractUC.EnvConfig["AMQP_URL"], queueBody, amqp.ResetPasswordMail, amqp.ResetPasswordMailDeadLetter)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_cash_request_amqp", uc.ReqID)
		return errors.New(helper.InternalServer)
	}

	return err
}

// GetTokenByKey ...
func (uc UserResetPasswordUC) GetTokenByKey(key string) (res viewmodel.JwtVM, err error) {
	ctx := "UserResetPasswordUC.GetTokenByKey"

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

	userUc := UserUC{ContractUC: uc.ContractUC}
	_, err = userUc.FindByID(id, false)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return res, err
	}

	payload := map[string]interface{}{
		"id":   id,
		"role": "user",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "generate_token", uc.ReqID)
		return res, err
	}

	uc.StoreToRedisExp("latestAction"+id, VerifyKey, "24h")
	res.LatestAction = VerifyKey

	return res, err
}

// NewPasswordSubmit ...
func (uc UserResetPasswordUC) NewPasswordSubmit(userID, password string) (err error) {
	ctx := "UserResetPasswordUC.NewPasswordSubmit"

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

	// Update password
	userUc := UserUC{ContractUC: uc.ContractUC}
	err = userUc.UpdateForgotPassword(userID, password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "um_reset_password", uc.ReqID)
		return err
	}

	uc.StoreToRedisExp("latestAction"+userID, UserLocked, "24h")
	uc.RemoveFromRedis("resetPasswordKey" + userID)

	return err
}
