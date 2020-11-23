package usecase

import (
	"errors"
	"kriyapeople/helper"
	"kriyapeople/model"
	"kriyapeople/pkg/amqp"
	"kriyapeople/pkg/apple"
	"kriyapeople/pkg/facebook"
	"kriyapeople/pkg/google"
	"kriyapeople/pkg/interfacepkg"
	"kriyapeople/pkg/logruslogger"
	"kriyapeople/pkg/str"
	"kriyapeople/server/request"
	"kriyapeople/usecase/viewmodel"
	"strings"
	"time"
)

// UserUC ...
type UserUC struct {
	*ContractUC
}

// BuildBody ...
func (uc UserUC) BuildBody(data *model.UserEntity, res *viewmodel.UserVM, isShowPassword bool) {

	res.ID = data.ID
	res.Email = data.Email.String
	res.EmailValidAt = data.EmailValidAt.String
	res.Name = data.Name.String
	res.Phone = data.Phone.String
	res.ProfileImageID = data.ProfileImageID.String
	res.Password = str.ShowString(isShowPassword, uc.Aes.DecryptNoErr(data.Password.String))
	res.RegisterType = data.RegisterType.String
	interfacepkg.UnmarshallCb(data.RegisterDetail.String, &res.RegisterDetail)
	res.Status = data.Status.Bool
	res.CreatedAt = data.CreatedAt
	res.UpdatedAt = data.UpdatedAt
	res.DeletedAt = data.DeletedAt.String
}

// FindAll ...
func (uc UserUC) FindAll(search, status string, page, limit int, by, sort string) (res []viewmodel.UserVM, pagination viewmodel.PaginationVM, err error) {
	ctx := "UserUC.FindAll"

	if !str.Contains(model.UserBy, by) {
		by = model.DefaultUserBy
	}
	if !str.Contains(SortWhitelist, strings.ToLower(sort)) {
		sort = DescSort
	}

	limit = uc.LimitMax(limit)
	limit, offset := uc.PaginationPageOffset(page, limit)

	userModel := model.NewUserModel(uc.DB)
	data, count, err := userModel.FindAll(search, status, offset, limit, by, sort)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, pagination, err
	}
	pagination = PaginationRes(page, count, limit)

	for _, r := range data {
		temp := viewmodel.UserVM{}
		uc.BuildBody(&r, &temp, false)
		res = append(res, temp)
	}

	return res, pagination, err
}

// FindByID ...
func (uc UserUC) FindByID(id string, isShowPassword bool) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.FindByID"

	userModel := model.NewUserModel(uc.DB)
	data, err := userModel.FindByID(id)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}
	uc.BuildBody(&data, &res, isShowPassword)

	return res, err
}

// FindByEmail ...
func (uc UserUC) FindByEmail(email string, isShowPassword bool) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.FindByEmail"

	userModel := model.NewUserModel(uc.DB)
	data, err := userModel.FindByEmail(email)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}
	uc.BuildBody(&data, &res, isShowPassword)

	return res, err
}

// Register ...
func (uc UserUC) Register(data request.UserRegisterRequest) (res viewmodel.JwtVM, err error) {
	ctx := "UserUC.Register"

	// Decrypt password input
	if data.Password != "" {
		data.Password, err = uc.AesFront.Decrypt(data.Password)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt", uc.ReqID)
			return res, errors.New(helper.InvalidCredentials)
		}
	}

	// Find duplicate user
	userModel := model.NewUserModel(uc.DB)
	user, _ := userModel.FindByEmail(data.Email)
	if user.ID != "" {
		logruslogger.Log(logruslogger.WarnLevel, data.Email, ctx, "email_used", uc.ReqID)
		return res, errors.New(helper.EmailExist)
	}

	// Encrypt password
	password, err := uc.Aes.Encrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return res, err
	}

	now := time.Now().UTC()
	userBody := viewmodel.UserVM{
		Email:          data.Email,
		Name:           data.Name,
		RegisterType:   data.RegisterType,
		RegisterDetail: data.RegisterDetail,
		Password:       password,
		Status:         true,
		CreatedAt:      now.Format(time.RFC3339),
		UpdatedAt:      now.Format(time.RFC3339),
	}
	userBody.ID, err = userModel.Store(userBody, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	if data.FcmDeviceToken != "" {
		userDeviceTokenUc := UserDeviceTokenUC{ContractUC: uc.ContractUC}
		userDeviceTokenUc.Create(user.ID, request.UserDeviceTokenRequest{
			Type:  model.UserDeviceTokenTypeFcm,
			Token: data.FcmDeviceToken,
		})
	}

	// Jwe the payload & Generate jwt token
	payload := map[string]interface{}{
		"id":   userBody.ID,
		"role": "user",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "generate_token", uc.ReqID)
		return res, err
	}

	mqueue := amqp.NewQueue(AmqpConnection, AmqpChannel)
	queueBody := map[string]interface{}{
		"qid":     uc.ContractUC.ReqID,
		"user_id": userBody.ID,
		"email":   userBody.Email,
		"name":    userBody.Name,
	}
	AmqpConnection, AmqpChannel, err = mqueue.PushQueueReconnect(uc.ContractUC.EnvConfig["AMQP_URL"], queueBody, amqp.ActivationMail, amqp.ActivationMailDeadLetter)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "email_validation", uc.ReqID)
		return res, err
	}

	return res, err
}

// ResendActivationMail ...
func (uc UserUC) ResendActivationMail(id string) (err error) {
	ctx := "UserUC.ResendActivationMail"

	userUc := UserUC{ContractUC: uc.ContractUC}
	user, err := userUc.FindByID(id, false)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return err
	}

	mqueue := amqp.NewQueue(AmqpConnection, AmqpChannel)
	queueBody := map[string]interface{}{
		"qid":     uc.ContractUC.ReqID,
		"user_id": id,
		"email":   user.Email,
		"name":    user.Name,
	}
	AmqpConnection, AmqpChannel, err = mqueue.PushQueueReconnect(uc.ContractUC.EnvConfig["AMQP_URL"], queueBody, amqp.ActivationMail, amqp.ActivationMailDeadLetter)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "amqp", uc.ReqID)
		return err
	}

	return err
}

// VerifyEmail ...
func (uc UserUC) VerifyEmail(key string) (res viewmodel.JwtVM, err error) {
	ctx := "UserUC.VerifyEmail"

	var userID string
	err = uc.GetFromRedis("emailValidation"+key, &userID)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "get_user_id_redis", uc.ReqID)
		return res, errors.New(helper.ExpKey)
	}

	validKey := ""
	err = uc.GetFromRedis("emailValidationKey"+userID, &validKey)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "get_key_from_redis", uc.ReqID)
		return res, errors.New(helper.ExpKey)
	}
	if validKey != key {
		logruslogger.Log(logruslogger.WarnLevel, validKey+"!="+key, ctx, "invalid_key", uc.ReqID)
		return res, errors.New(helper.ExpKey)
	}

	// Jwe the payload & Generate jwt token
	payload := map[string]interface{}{
		"id":   userID,
		"role": "user",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "generate_token", uc.ReqID)
		return res, err
	}

	err = uc.UpdateEmailValidAt(userID)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_email_valid_at", uc.ReqID)
		return res, err
	}

	return res, err
}

// Login ...
func (uc UserUC) Login(data request.UserLoginRequest) (res viewmodel.JwtVM, err error) {
	ctx := "UserUC.Login"

	// Verify user by login type
	var user viewmodel.UserVM
	if data.RegisterType == model.UserRegisterTypeEmail {
		user, err = uc.loginEmail(&data)
	} else if data.RegisterType == model.UserRegisterTypeFacebook {
		user, err = uc.loginFacebook(&data)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "login", uc.ReqID)
			return res, err
		}
	} else if data.RegisterType == model.UserRegisterTypeGmail {
		user, err = uc.loginGmail(&data)
	} else if data.RegisterType == model.UserRegisterTypeApple {
		user, err = uc.loginApple(&data)
	}
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "login", uc.ReqID)
		return res, errors.New(helper.InvalidCredential)
	}

	// Check user status
	if !user.Status {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "inactive_user", uc.ReqID)
		return res, errors.New(helper.InactiveUser)
	}

	if data.FcmDeviceToken != "" {
		userDeviceTokenUc := UserDeviceTokenUC{ContractUC: uc.ContractUC}
		userDeviceTokenUc.Create(user.ID, request.UserDeviceTokenRequest{
			Type:  model.UserDeviceTokenTypeFcm,
			Token: data.FcmDeviceToken,
		})
	}

	// Jwe the payload & Generate jwt token
	payload := map[string]interface{}{
		"id":   user.ID,
		"role": "user",
	}
	jwtUc := JwtUC{ContractUC: uc.ContractUC}
	err = jwtUc.GenerateToken(payload, &res)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "generate_token", uc.ReqID)
		return res, err
	}

	return res, err
}

// loginEmail ...
func (uc UserUC) loginEmail(data *request.UserLoginRequest) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.loginEmail"

	// Find user
	res, err = uc.FindByEmail(data.Email, true)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return res, err
	}

	// Decrypt password input
	data.Password, err = uc.AesFront.Decrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt", uc.ReqID)
		return res, err
	}

	// Verify password
	if res.Password != data.Password {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "invalid_password", uc.ReqID)
		return res, errors.New(helper.InvalidPassword)
	}

	return res, err
}

// loginFacebook ...
func (uc UserUC) loginFacebook(data *request.UserLoginRequest) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.loginFacebook"

	// Verify token
	facebookUser, err := facebook.GetFacebookProfile(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "verify_token", uc.ReqID)
		return res, err
	}
	if facebookUser["email"] == nil {
		logruslogger.Log(logruslogger.WarnLevel, interfacepkg.Marshall(facebookUser), ctx, "facebook_email", uc.ReqID)
		return res, errors.New(helper.RequiredFacebookEmail)
	}

	if facebookUser["email"].(string) == "" {
		logruslogger.Log(logruslogger.WarnLevel, interfacepkg.Marshall(facebookUser), ctx, "facebook_email", uc.ReqID)
		return res, errors.New(helper.RequiredFacebookEmail)
	}

	// Verify email
	if data.Email != facebookUser["email"].(string) {
		logruslogger.Log(logruslogger.WarnLevel, data.Email+" != "+facebookUser["email"].(string), ctx, "invalid_email", uc.ReqID)
		return res, errors.New(helper.InvalidEmail)
	}

	// Find user
	res, _ = uc.FindByEmail(data.Email, false)
	if res.ID == "" {
		fileID := ""
		// Get name
		name := ""
		if facebookUser["name"] != nil {
			name = facebookUser["name"].(string)
		}

		now := time.Now().UTC()
		res = viewmodel.UserVM{
			Email:          data.Email,
			EmailValidAt:   now.Format(time.RFC3339),
			Name:           name,
			ProfileImageID: fileID,
			RegisterType:   data.RegisterType,
			RegisterDetail: interfacepkg.Marshall(facebookUser),
			Status:         true,
			CreatedAt:      now.Format(time.RFC3339),
			UpdatedAt:      now.Format(time.RFC3339),
		}
		userModel := model.NewUserModel(uc.DB)
		res.ID, err = userModel.Store(res, now)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
			return res, err
		}
	}

	return res, err
}

// loginGmail ...
func (uc UserUC) loginGmail(data *request.UserLoginRequest) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.loginGmail"

	// Verify token
	gmailUser, err := google.GetGoogleProfile(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "verify_token", uc.ReqID)
		return res, err
	}
	if gmailUser["email"] == nil {
		logruslogger.Log(logruslogger.WarnLevel, interfacepkg.Marshall(gmailUser), ctx, "gmail_email", uc.ReqID)
		return res, errors.New(helper.InvalidEmail)
	}

	// Verify email
	if data.Email != gmailUser["email"].(string) {
		logruslogger.Log(logruslogger.WarnLevel, data.Email+" != "+gmailUser["email"].(string), ctx, "invalid_email", uc.ReqID)
		return res, errors.New(helper.InvalidEmail)
	}

	// Find user
	res, _ = uc.FindByEmail(data.Email, false)
	if res.ID == "" {
		fileID := ""

		name := ""
		if gmailUser["name"] != nil {
			name = gmailUser["name"].(string)
		}

		now := time.Now().UTC()
		res = viewmodel.UserVM{
			Email:          data.Email,
			EmailValidAt:   now.Format(time.RFC3339),
			Name:           name,
			ProfileImageID: fileID,
			RegisterType:   data.RegisterType,
			RegisterDetail: interfacepkg.Marshall(gmailUser),
			Status:         true,
			CreatedAt:      now.Format(time.RFC3339),
			UpdatedAt:      now.Format(time.RFC3339),
		}
		userModel := model.NewUserModel(uc.DB)
		res.ID, err = userModel.Store(res, now)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
			return res, err
		}
	}

	return res, err
}

// loginApple ...
func (uc UserUC) loginApple(data *request.UserLoginRequest) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.loginApple"

	// Verify token
	err = apple.VerifyJWT(data.Password, data.Email)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "verify_token", uc.ReqID)
		return res, err
	}

	// Find user
	res, _ = uc.FindByEmail(data.Email, false)
	if res.ID == "" {
		now := time.Now().UTC()
		res = viewmodel.UserVM{
			Email:        data.Email,
			EmailValidAt: now.Format(time.RFC3339),
			Name:         data.Email,
			RegisterType: data.RegisterType,
			Status:       true,
			CreatedAt:    now.Format(time.RFC3339),
			UpdatedAt:    now.Format(time.RFC3339),
		}
		userModel := model.NewUserModel(uc.DB)
		res.ID, err = userModel.Store(res, now)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
			return res, err
		}
	}

	return res, err
}

// Logout ...
func (uc UserUC) Logout(userID string, data request.UserLogoutRequest) (err error) {
	ctx := "UserUC.Logout"

	if data.FcmDeviceToken != "" {
		userDeviceTokenUc := UserDeviceTokenUC{ContractUC: uc.ContractUC}
		_, err = userDeviceTokenUc.DeleteByUserToken(userID, model.UserDeviceTokenTypeFcm, data.FcmDeviceToken)
		if err != nil {
			logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "delete_token", uc.ReqID)
			return err
		}
	}

	return err
}

// UpdatePassword ...
func (uc UserUC) UpdatePassword(id string, data request.UserUpdatePasswordRequest) (err error) {
	ctx := "UserUC.UpdatePassword"

	// Find user
	user, err := uc.FindByID(id, true)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "find_user", uc.ReqID)
		return err
	}
	if user.RegisterType != model.UserRegisterTypeEmail {
		logruslogger.Log(logruslogger.WarnLevel, user.RegisterType, ctx, "register_type", uc.ReqID)
		return errors.New(helper.InvalidBody)
	}

	// Decrypt password input
	data.OldPassword, err = uc.AesFront.Decrypt(data.OldPassword)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt_old_password", uc.ReqID)
		return err
	}
	data.Password, err = uc.AesFront.Decrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt_password", uc.ReqID)
		return err
	}

	// Verify old password
	if data.OldPassword != user.Password {
		logruslogger.Log(logruslogger.WarnLevel, "", ctx, "wrong_password", uc.ReqID)
		return errors.New(helper.InvalidPassword)
	}

	// Encrypt password
	password, err := uc.Aes.Encrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return err
	}

	now := time.Now().UTC()
	userModel := model.NewUserModel(uc.DB)
	_, err = userModel.UpdatePassword(id, password, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_password", uc.ReqID)
		return err
	}

	return err
}

// UpdateForgotPassword ...
func (uc UserUC) UpdateForgotPassword(id, password string) (err error) {
	ctx := "UserUC.UpdateForgotPassword"

	// Encrypt password
	password, err = uc.Aes.Encrypt(password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return err
	}

	now := time.Now().UTC()
	userModel := model.NewUserModel(uc.DB)
	_, err = userModel.UpdatePassword(id, password, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_password", uc.ReqID)
		return err
	}

	return err
}

// Update ...
func (uc UserUC) Update(id string, data request.UserRequest) (err error) {
	ctx := "UserUC.Update"

	oldData, _ := uc.FindByEmail(data.Email, false)
	if oldData.ID != "" && oldData.ID != id {
		logruslogger.Log(logruslogger.WarnLevel, data.Email, ctx, "email_exist", uc.ReqID)
		return errors.New(helper.DuplicateEmail)
	}

	data.Password, err = uc.AesFront.Decrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "decrypt_password", uc.ReqID)
		return err
	}
	data.Password, err = uc.Aes.Encrypt(data.Password)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "encrypt_password", uc.ReqID)
		return err
	}

	now := time.Now().UTC()
	body := viewmodel.UserVM{
		Email:          data.Email,
		EmailValidAt:   data.EmailValidAt,
		Name:           data.Name,
		Phone:          data.Phone,
		ProfileImageID: data.ProfileImageID,
		Password:       data.Password,
		RegisterType:   data.RegisterType,
		RegisterDetail: data.RegisterDetail,
		Status:         data.Status,
	}
	userModel := model.NewUserModel(uc.DB)
	_, err = userModel.Update(id, body, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_profile", uc.ReqID)
		return err
	}

	return err
}

// UpdateProfile ...
func (uc UserUC) UpdateProfile(id string, data request.UserUpdateProfileRequest) (err error) {
	ctx := "UserUC.UpdateProfile"

	now := time.Now().UTC()
	body := viewmodel.UserVM{
		Name:           data.Name,
		Phone:          data.Phone,
		ProfileImageID: data.ProfileImageID,
	}
	userModel := model.NewUserModel(uc.DB)
	_, err = userModel.UpdateProfile(id, body, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update_profile", uc.ReqID)
		return err
	}

	return err
}

// UpdateEmailValidAt ...
func (uc UserUC) UpdateEmailValidAt(id string) (err error) {
	ctx := "UserUC.UpdateEmailValidAt"

	now := time.Now().UTC()
	userModel := model.NewUserModel(uc.DB)
	_, err = userModel.UpdateEmailValidAt(id, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "update", uc.ReqID)
		return err
	}

	return err
}

// Delete ...
func (uc UserUC) Delete(id string) (res viewmodel.UserVM, err error) {
	ctx := "UserUC.Delete"

	now := time.Now().UTC()
	userModel := model.NewUserModel(uc.DB)
	res.ID, err = userModel.Destroy(id, now)
	if err != nil {
		logruslogger.Log(logruslogger.WarnLevel, err.Error(), ctx, "query", uc.ReqID)
		return res, err
	}

	return res, err
}
