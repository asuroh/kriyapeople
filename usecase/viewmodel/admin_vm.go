package viewmodel

// AdminVM ....
type AdminVM struct {
	ID        string `json:"id"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	RoleID    string `json:"role_id"`
	RoleName  string `json:"role_name"`
	Status    bool   `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at"`
}

// AdminLoginVM ....
type AdminLoginVM struct {
	Token       string `json:"token"`
	ExpiredDate string `json:"expired_date"`
}
