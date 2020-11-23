package viewmodel

// UserVM ...
type UserVM struct {
	ID              string      `json:"id"`
	Email           string      `json:"email"`
	EmailValidAt    string      `json:"email_valid_at"`
	Name            string      `json:"name"`
	Phone           string      `json:"phone"`
	ProfileImageID  string      `json:"profile_image_id"`
	ProfileImageURL string      `json:"profile_image"`
	Password        string      `json:"password"`
	RegisterType    string      `json:"register_type"`
	RegisterDetail  interface{} `json:"register_detail"`
	Status          bool        `json:"status"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
	DeletedAt       string      `json:"deleted_at"`
}
