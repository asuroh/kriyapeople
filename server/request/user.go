package request

// UserRequest ...
type UserRequest struct {
	Email          string      `json:"email" validate:"email"`
	EmailValidAt   string      `json:"email_valid_at"`
	Name           string      `json:"name" validate:"required"`
	Phone          string      `json:"phone"`
	ProfileImageID string      `json:"profile_image_id"`
	Password       string      `json:"password"`
	RegisterType   string      `json:"register_type"`
	RegisterDetail interface{} `json:"register_detail"`
	Status         bool        `json:"status"`
}

// UserRegisterRequest ....
type UserRegisterRequest struct {
	Email          string `json:"email" validate:"email"`
	Name           string `json:"name" validate:"required"`
	RegisterType   string `json:"register_type" validate:"required,oneof=email"`
	RegisterDetail string `json:"register_detail"`
	Password       string `json:"password" validate:"required"`
	FcmDeviceToken string `json:"fcm_device_token"`
}

// UserLoginRequest ....
type UserLoginRequest struct {
	Email          string `json:"email" validate:"omitempty,email"`
	RegisterType   string `json:"register_type" validate:"required,oneof=email facebook gmail apple"`
	Password       string `json:"password" validate:"required"`
	FcmDeviceToken string `json:"fcm_device_token"`
}

// UserLogoutRequest ....
type UserLogoutRequest struct {
	FcmDeviceToken string `json:"fcm_device_token"`
}

// UserUpdatePasswordRequest ....
type UserUpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

// UserUpdateProfileRequest ...
type UserUpdateProfileRequest struct {
	Name           string `json:"name" validate:"required"`
	Phone          string `json:"phone"`
	ProfileImageID string `json:"profile_image_id"`
}
