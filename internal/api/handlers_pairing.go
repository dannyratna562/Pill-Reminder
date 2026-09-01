package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/pillreminder/backend/internal/domain"
)

// PairingStore is what PairingHandler needs from persistence. Satisfied by
// *store.PairingStore in production.
type PairingStore interface {
	CreatePairingCode(ctx context.Context, parentID uuid.UUID) (domain.PairingCode, error)
	RedeemAndLink(ctx context.Context, code string, childID uuid.UUID) (domain.FamilyLink, error)
}

type PairingHandler struct {
	store  PairingStore
	logger *slog.Logger
}

func NewPairingHandler(store PairingStore, logger *slog.Logger) *PairingHandler {
	return &PairingHandler{store: store, logger: logger}
}

type requestPairingCodeRequest struct {
	ParentID string `json:"parent_id"`
}

type pairingCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RequestCode handles POST /v1/pairing-codes — the Parent app requesting a
// short-lived code for the Child app to redeem.
func (h *PairingHandler) RequestCode(w http.ResponseWriter, r *http.Request) {
	var req requestPairingCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid request body")
		return
	}

	parentID, err := domain.ParseID(req.ParentID, "parent_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	code, err := h.store.CreatePairingCode(r.Context(), parentID)
	if err != nil {
		h.logger.Error("create pairing code failed", "err", err)
		respondError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	respondData(w, http.StatusCreated, pairingCodeResponse{
		Code:      code.Code,
		ExpiresAt: code.ExpiresAt,
	})
}

type redeemPairingCodeRequest struct {
	Code    string `json:"code"`
	ChildID string `json:"child_id"`
}

type familyLinkResponse struct {
	ID        string    `json:"id"`
	ChildID   string    `json:"child_id"`
	ParentID  string    `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Redeem handles POST /v1/family-links — the Child app redeeming a pairing
// code and creating the resulting family_link.
func (h *PairingHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	var req redeemPairingCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "invalid request body")
		return
	}

	childID, err := domain.ParseID(req.ChildID, "child_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := domain.ValidateCodeFormat(req.Code); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	link, err := h.store.RedeemAndLink(r.Context(), req.Code, childID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPairingCodeNotFound):
			respondError(w, http.StatusNotFound, "pairing_code_not_found", "pairing code not found")
		case errors.Is(err, domain.ErrPairingCodeExpired):
			respondError(w, http.StatusBadRequest, "pairing_code_expired", "pairing code has expired")
		case errors.Is(err, domain.ErrPairingCodeUsed):
			respondError(w, http.StatusBadRequest, "pairing_code_used", "pairing code has already been used")
		default:
			h.logger.Error("redeem pairing code failed", "err", err)
			respondError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		}
		return
	}

	respondData(w, http.StatusCreated, familyLinkResponse{
		ID:        link.ID.String(),
		ChildID:   link.ChildID.String(),
		ParentID:  link.ParentID.String(),
		CreatedAt: link.CreatedAt,
	})
}
