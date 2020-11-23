package handler

import (
	"kriyapeople/helper"
	"kriyapeople/server/request"
	"kriyapeople/usecase"
	"net/http"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"
)

// AdminResetPasswordHandler ...
type AdminResetPasswordHandler struct {
	Handler
}

// ForgotPasswordHandler ...
func (h *AdminResetPasswordHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	req := request.ForgotPasswordRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		SendBadRequest(w, helper.InvalidBody)
		return
	}

	adminResetPasswordUC := usecase.AdminResetPasswordUC{ContractUC: h.ContractUC}
	err := adminResetPasswordUC.ReqResetPassword(req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}

// GetTokenByKeyHandler ...
func (h *AdminResetPasswordHandler) GetTokenByKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		SendBadRequest(w, "Parameter must be filled")
		return
	}

	adminResetPasswordUC := usecase.AdminResetPasswordUC{ContractUC: h.ContractUC}
	res, err := adminResetPasswordUC.GetTokenByKey(key)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// NewPasswordSubmitHandler ...
func (h *AdminResetPasswordHandler) NewPasswordSubmitHandler(w http.ResponseWriter, r *http.Request) {
	userID := requestKeyFromContextInterface(r.Context(), "user", "id")

	req := request.NewPasswordSubmitRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	adminResetPasswordUC := usecase.AdminResetPasswordUC{ContractUC: h.ContractUC}
	err := adminResetPasswordUC.NewPasswordSubmit(userID, req.Password)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}
