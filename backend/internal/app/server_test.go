package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteStoreErrorMapsValidationErrorsToBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeStoreError(recorder, fmt.Errorf("%w: enter a valid email address", errValidation))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
