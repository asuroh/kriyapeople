package bootstrap

import (
	"kriyapeople/pkg/logruslogger"
	api "kriyapeople/server/handler"
	"kriyapeople/server/middleware"

	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/go-chi/chi"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RegisterRoutes ...
func (boot *Bootup) RegisterRoutes() {
	handlerType := api.Handler{
		DB:         boot.DB,
		EnvConfig:  boot.EnvConfig,
		Validate:   boot.Validator,
		Translator: boot.Translator,
		ContractUC: &boot.ContractUC,
		Jwe:        boot.Jwe,
		Jwt:        boot.Jwt,
	}
	mJwt := middleware.VerifyMiddlewareInit{
		ContractUC: &boot.ContractUC,
	}

	boot.R.Route("/v1", func(r chi.Router) {
		// Define a limit rate to 1000 requests per IP per request.
		rate, _ := limiter.NewRateFromFormatted("1000-S")
		store, _ := sredis.NewStoreWithOptions(boot.ContractUC.Redis, limiter.StoreOptions{
			Prefix:   "limiter_global",
			MaxRetry: 3,
		})
		rateMiddleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))
		r.Use(rateMiddleware.Handler)

		// Logging setup
		r.Use(chimiddleware.RequestID)
		r.Use(logruslogger.NewStructuredLogger(boot.EnvConfig["LOG_FILE_PATH"], boot.EnvConfig["LOG_DEFAULT"], boot.ContractUC.ReqID))
		r.Use(chimiddleware.Recoverer)

		r.Route("/api", func(r chi.Router) {
			userHandler := api.UserHandler{Handler: handlerType}
			r.Route("/user", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Post("/register", userHandler.RegisterHandler)
					r.Post("/login", userHandler.LoginHandler)
					r.Get("/activationMail/{key}", userHandler.VerifyEmailHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyUserToken)
					r.Post("/logout", userHandler.LogoutHandler)
					r.Get("/token", userHandler.GetByTokenHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyRefreshTokenCredential)
					r.Get("/generateToken", userHandler.GenerateTokenHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyUserEmailValidToken)
					r.Put("/password", userHandler.UpdatePasswordHandler)
					r.Put("/profile", userHandler.UpdateProfileHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyUserEmailInvalidToken)
					r.Post("/resendActivationMail", userHandler.ResendActivationMailHandler)
				})
			})

			userResetPasswordHandler := api.UserResetPasswordHandler{Handler: handlerType}
			r.Route("/userResetPassword", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					limitInit := middleware.LimitInit{
						ContractUC: &boot.ContractUC,
						MaxLimit:   5,
						Duration:   "24h",
					}
					r.Use(limitInit.LimitForgotPassword)
					r.Post("/", userResetPasswordHandler.ForgotPasswordHandler)
				})
				r.Group(func(r chi.Router) {
					r.Get("/token/key/{key}", userResetPasswordHandler.GetTokenByKeyHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyUserForgotPasswordTokenCredential)
					r.Post("/newPassword", userResetPasswordHandler.NewPasswordSubmitHandler)
				})
			})
		})

		// API ADMIN
		r.Route("/api-admin", func(r chi.Router) {
			adminHandler := api.AdminHandler{Handler: handlerType}
			r.Route("/admin", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Post("/login", adminHandler.LoginHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifySuperadminTokenCredential)
					r.Get("/", adminHandler.GetAllHandler)
					r.Get("/id/{id}", adminHandler.GetByIDHandler)
					r.Get("/code/{code}", adminHandler.GetByCodeHandler)
					r.Post("/", adminHandler.CreateHandler)
					r.Put("/id/{id}", adminHandler.UpdateHandler)
					r.Delete("/id/{id}", adminHandler.DeleteHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyAdminTokenCredential)
					r.Get("/token", adminHandler.GetByTokenHandler)
					r.Put("/token", adminHandler.UpdateByTokenHandler)
				})
			})

			adminResetPasswordHandler := api.AdminResetPasswordHandler{Handler: handlerType}
			r.Route("/adminResetPassword", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					limitInit := middleware.LimitInit{
						ContractUC: &boot.ContractUC,
						MaxLimit:   5,
						Duration:   "24h",
					}
					r.Use(limitInit.LimitForgotPassword)
					r.Post("/", adminResetPasswordHandler.ForgotPasswordHandler)
				})
				r.Group(func(r chi.Router) {
					r.Get("/token/key/{key}", adminResetPasswordHandler.GetTokenByKeyHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(mJwt.VerifyAdminForgotPasswordTokenCredential)
					r.Post("/newPassword", adminResetPasswordHandler.NewPasswordSubmitHandler)
				})
			})

			roleHandler := api.RoleHandler{Handler: handlerType}
			r.Route("/role", func(r chi.Router) {
				r.Use(mJwt.VerifyAdminTokenCredential)
				r.Get("/select", roleHandler.SelectAllHandler)
			})
		})
	})
}
