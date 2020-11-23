package request

// UserDeviceTokenRequest ....
type UserDeviceTokenRequest struct {
	Type  string `json:"type" validate:"required,oneof=fcm"`
	Token string `json:"token" validate:"required"`
}
