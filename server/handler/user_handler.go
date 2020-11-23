package handler

import (
	"github.com/go-chi/chi"
	"kriyapeople/server/request"
	"kriyapeople/usecase"
	"kriyapeople/usecase/viewmodel"
	"net/http"
	"strconv"

	validator "gopkg.in/go-playground/validator.v9"
)

// UserHandler ...
type UserHandler struct {
	Handler
}

// GetAllHandler ...
func (h *UserHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		SendBadRequest(w, "Invalid page value")
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		SendBadRequest(w, "Invalid limit value")
		return
	}

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	by := r.URL.Query().Get("by")
	sort := r.URL.Query().Get("sort")

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	res, p, err := userUc.FindAll(search, status, page, limit, by, sort)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, p)
	return
}

// GetByTokenHandler ...
func (h *UserHandler) GetByTokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	res, err := userUc.FindByID(userID, false)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// RegisterHandler ...
func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	req := request.UserRegisterRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	res, err := userUc.Register(req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// ResendActivationMailHandler ...
func (h *UserHandler) ResendActivationMailHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	err := userUc.ResendActivationMail(userID)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}

// VerifyEmailHandler ...
func (h *UserHandler) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		SendBadRequest(w, "Parameter must be filled")
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	res, err := userUc.VerifyEmail(key)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// LoginHandler ...
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	req := request.UserLoginRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	res, err := userUc.Login(req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// LogoutHandler ...
func (h *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	req := request.UserLogoutRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	err := userUc.Logout(userID, req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}

// GenerateTokenHandler ...
func (h *UserHandler) GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	payload := map[string]interface{}{
		"id":   userID,
		"role": "user",
	}
	var res viewmodel.JwtVM
	jwtUc := usecase.JwtUC{ContractUC: h.ContractUC}
	err := jwtUc.GenerateToken(payload, &res)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// UpdatePasswordHandler ...
func (h *UserHandler) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	req := request.UserUpdatePasswordRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	err := userUc.UpdatePassword(userID, req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}

// UpdateProfileHandler ...
func (h *UserHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	req := request.UserUpdateProfileRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	userUc := usecase.UserUC{ContractUC: h.ContractUC}
	err := userUc.UpdateProfile(userID, req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}
