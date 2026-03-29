package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) error {
	return decodeJSONWithLimit(w, r, out, 1<<20)
}

func decodeJSONWithLimit(w http.ResponseWriter, r *http.Request, out any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
				"error": fmt.Sprintf("request body too large (max %d MiB)", maxBytes>>20),
			})
			return err
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
