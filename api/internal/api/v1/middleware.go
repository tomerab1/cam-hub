package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"tomerab.com/cam-hub/internal/api"
)

type contextKey string

const validatedBodyKey contextKey = "validatedBody"

var validate = validator.New()

func validationMiddleware[T any](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body T

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid JSON format",
			})
			_ = r.Body.Close()
			return
		}
		defer r.Body.Close()

		if err := validate.Struct(body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			var validationErrors []ValidationError
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, ValidationError{
					Field:   err.Field(),
					Message: getErrorMsg(err),
				})
			}

			json.NewEncoder(w).Encode(api.ErrorEnvp{"error": validationErrors})
			return
		}

		ctx := context.WithValue(r.Context(), validatedBodyKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getErrorMsg(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	default:
		return "Invalid value"
	}
}

func getValidatedBody[T any](r *http.Request) T {
	body, _ := r.Context().Value(validatedBodyKey).(T)
	return body
}
