package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// Shared validator instance for all handlers in this package.
var validate = validator.New()

// decodeAndValidate decodes the JSON request body into dst and validates it.
// If it fails, it writes the appropriate HTTP error and returns false.
// This is a package-level helper that any handler can reuse.
func decodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	if err := validate.Struct(dst); err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
		return false
	}
	return true
}
