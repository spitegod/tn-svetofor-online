package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tn/backend/internal/apperror"
)

func TestOrderIDFromRequestRejectsInvalidValue(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/system-catalog?orderId=wrong", nil)
	if _, err := orderIDFromRequest(request); err == nil {
		t.Fatal("expected invalid orderId to be rejected")
	}
}

func TestWriteActionErrorUsesTypedErrorKind(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeActionError(recorder, apperror.New(apperror.Validation, "invalid value"))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"validation_error"`) {
		t.Fatalf("unexpected body %s", recorder.Body.String())
	}
}

func TestOrderIDFromRequestDefaultsOnlyWhenMissing(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/system-catalog", nil)
	id, err := orderIDFromRequest(request)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("order id = %d, want 1", id)
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	request := httptest.NewRequest("PATCH", "/api/orders/1", strings.NewReader(`{"name":"test","unexpected":true}`))
	recorder := httptest.NewRecorder()
	var payload struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(recorder, request, &payload); err == nil {
		t.Fatal("expected unknown JSON field to be rejected")
	}
}
