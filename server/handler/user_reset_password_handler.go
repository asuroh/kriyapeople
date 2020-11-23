package handler

import (
	"kriyapeople/helper"
	"kriyapeople/server/request"
	"kriyapeople/usecase"
	"net/http"

	"github.com/go-chi/chi"
	validator "gopkg.in/go-playground/validator.v9"
)

// UserResetPasswordHandler ...
type UserResetPasswordHandler struct {
	Handler
}

// ForgotPasswordHandler ...
func (h *UserResetPasswordHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	req := request.ForgotPasswordRequest{}
	if err := h.Handler.Bind(r, &req); err != nil {
		SendBadRequest(w, err.Error())
		return
	}
	if err := h.Handler.Validate.Struct(req); err != nil {
		SendBadRequest(w, helper.InvalidBody)
		return
	}

	userResetPasswordUC := usecase.UserResetPasswordUC{ContractUC: h.ContractUC}
	err := userResetPasswordUC.ReqResetPassword(req)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}

// GetTokenByKeyHandler ...
func (h *UserResetPasswordHandler) GetTokenByKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		SendBadRequest(w, "Parameter must be filled")
		return
	}

	userResetPasswordUC := usecase.UserResetPasswordUC{ContractUC: h.ContractUC}
	res, err := userResetPasswordUC.GetTokenByKey(key)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, res, nil)
	return
}

// NewPasswordSubmitHandler ...
func (h *UserResetPasswordHandler) NewPasswordSubmitHandler(w http.ResponseWriter, r *http.Request) {
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

	userResetPasswordUC := usecase.UserResetPasswordUC{ContractUC: h.ContractUC}
	err := userResetPasswordUC.NewPasswordSubmit(userID, req.Password)
	if err != nil {
		SendBadRequest(w, err.Error())
		return
	}

	SendSuccess(w, nil, nil)
	return
}
