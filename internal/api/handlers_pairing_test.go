package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/pillreminder/backend/internal/domain"
)

type fakePairingStore struct {
	createCodeFn func(ctx context.Context, parentID uuid.UUID) (domain.PairingCode, error)
	redeemFn     func(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error)
	createCalled bool
	redeemCalled bool
}

func (f *fakePairingStore) CreatePairingCode(ctx context.Context, parentID uuid.UUID) (domain.PairingCode, error) {
	f.createCalled = true
	return f.createCodeFn(ctx, parentID)
}

func (f *fakePairingStore) RedeemAndLink(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error) {
	f.redeemCalled = true
	return f.redeemFn(ctx, code, childID)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func doRequest(t *testing.T, handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestRequestCode_Success(t *testing.T) {
	parentID := uuid.New()
	expiresAt := time.Now().Add(domain.PairingCodeTTL)
	fake := &fakePairingStore{
		createCodeFn: func(ctx context.Context, gotParentID uuid.UUID) (domain.PairingCode, error) {
			if gotParentID != parentID {
				t.Errorf("CreatePairingCode() parentID = %v, want %v", gotParentID, parentID)
			}
			return domain.PairingCode{Code: "042817", ParentID: parentID, ExpiresAt: expiresAt}, nil
		},
	}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.RequestCode, requestPairingCodeRequest{ParentID: parentID.String()})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var got dataEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data := got.Data.(map[string]any)
	if data["code"] != "042817" {
		t.Errorf("data.code = %v, want 042817", data["code"])
	}
}

func TestRequestCode_InvalidParentID(t *testing.T) {
	fake := &fakePairingStore{}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.RequestCode, requestPairingCodeRequest{ParentID: "not-a-uuid"})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if fake.createCalled {
		t.Error("CreatePairingCode() should not be called for invalid input")
	}
}

func TestRedeem_Success(t *testing.T) {
	childID := uuid.New()
	parentID := uuid.New()
	linkID := uuid.New()
	createdAt := time.Now()
	fake := &fakePairingStore{
		redeemFn: func(ctx context.Context, code string, gotChildID uuid.UUID) (domain.FamilyLink, error) {
			if code != "042817" {
				t.Errorf("RedeemAndLink() code = %q, want 042817", code)
			}
			if gotChildID != childID {
				t.Errorf("RedeemAndLink() childID = %v, want %v", gotChildID, childID)
			}
			return domain.FamilyLink{ID: linkID, ChildID: childID, ParentID: parentID, CreatedAt: createdAt}, nil
		},
	}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.Redeem, redeemPairingCodeRequest{Code: "042817", ChildID: childID.String()})

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var got dataEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data := got.Data.(map[string]any)
	if data["id"] != linkID.String() || data["child_id"] != childID.String() || data["parent_id"] != parentID.String() {
		t.Errorf("data = %v, want family_link for %v/%v/%v", data, linkID, childID, parentID)
	}
}

func TestRedeem_ExpiredCode(t *testing.T) {
	fake := &fakePairingStore{
		redeemFn: func(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error) {
			return domain.FamilyLink{}, domain.ErrPairingCodeExpired
		},
	}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.Redeem, redeemPairingCodeRequest{Code: "042817", ChildID: uuid.New().String()})

	assertErrorResponse(t, w, http.StatusBadRequest, "pairing_code_expired")
}

func TestRedeem_UsedCode(t *testing.T) {
	fake := &fakePairingStore{
		redeemFn: func(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error) {
			return domain.FamilyLink{}, domain.ErrPairingCodeUsed
		},
	}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.Redeem, redeemPairingCodeRequest{Code: "042817", ChildID: uuid.New().String()})

	assertErrorResponse(t, w, http.StatusBadRequest, "pairing_code_used")
}

func TestRedeem_NotFoundCode(t *testing.T) {
	fake := &fakePairingStore{
		redeemFn: func(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error) {
			return domain.FamilyLink{}, domain.ErrPairingCodeNotFound
		},
	}
	h := NewPairingHandler(fake, testLogger())

	w := doRequest(t, h.Redeem, redeemPairingCodeRequest{Code: "042817", ChildID: uuid.New().String()})

	assertErrorResponse(t, w, http.StatusNotFound, "pairing_code_not_found")
}

func TestRedeem_InvalidBody(t *testing.T) {
	fake := &fakePairingStore{}
	h := NewPairingHandler(fake, testLogger())

	tests := []struct {
		name string
		req  redeemPairingCodeRequest
	}{
		{"bad child_id", redeemPairingCodeRequest{Code: "042817", ChildID: "not-a-uuid"}},
		{"bad code format", redeemPairingCodeRequest{Code: "abc", ChildID: uuid.New().String()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake.redeemCalled = false
			w := doRequest(t, h.Redeem, tt.req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if fake.redeemCalled {
				t.Error("RedeemAndLink() should not be called for invalid input")
			}
		})
	}
}

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, wantStatus, w.Body.String())
	}
	var got errorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", got.Error.Code, wantCode)
	}
}
