package middleware

import (
	"errors"
	"net/http"
	"strings"

	principal "github.com/example/scholarship-os/apps/api/pkg/auth"
	"github.com/example/scholarship-os/apps/api/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EnforceWorkspaceOwnership is the HTTP boundary guard for existing Phase 1-4
// resource routes. Domain services additionally constrain create/list operations.
func EnforceWorkspaceOwnership(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := principal.PrincipalFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required")
				return
			}
			if actor.IsSystem() {
				next.ServeHTTP(w, r)
				return
			}
			if actor.UserID == nil {
				response.Error(w, http.StatusForbidden, "FORBIDDEN", "credential is not associated with a user workspace")
				return
			}
			if !authorizedRouteOwner(r, db, *actor.UserID) {
				response.Error(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "resource not found")
				return
			}
			query := r.URL.Query()
			query.Set("userId", actor.UserID.String())
			r.URL.RawQuery = query.Encode()
			next.ServeHTTP(w, r)
		})
	}
}

func authorizedRouteOwner(r *http.Request, db *gorm.DB, owner uuid.UUID) bool {
	if raw := chi.URLParam(r, "userID"); raw != "" {
		id, err := uuid.Parse(raw)
		return err == nil && id == owner
	}
	checks := []struct {
		param string
		query string
	}{
		{"profileID", "SELECT user_id FROM applicant_profiles WHERE id = ? AND deleted_at IS NULL"},
		{"otherProfileID", "SELECT user_id FROM applicant_profiles WHERE id = ? AND deleted_at IS NULL"},
		{"applicationID", "SELECT user_id FROM applications WHERE id = ? AND deleted_at IS NULL"},
		{"proposalID", "SELECT user_id FROM application_proposals WHERE id = ?"},
		{"questionnaireID", "SELECT a.user_id FROM application_questionnaires q JOIN applications a ON a.id = q.application_id WHERE q.id = ? AND q.deleted_at IS NULL AND a.deleted_at IS NULL"},
		{"questionID", "SELECT a.user_id FROM application_questions q JOIN application_questionnaires aq ON aq.id = q.questionnaire_id JOIN applications a ON a.id = aq.application_id WHERE q.id = ? AND q.deleted_at IS NULL AND aq.deleted_at IS NULL AND a.deleted_at IS NULL"},
		{"requestID", "SELECT user_id FROM information_requests WHERE id = ?"},
	}
	for _, check := range checks {
		raw := chi.URLParam(r, check.param)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil || !queryOwner(r, db, check.query, id, owner) {
			return false
		}
	}
	if raw := chi.URLParam(r, "taskID"); raw != "" && strings.Contains(r.URL.Path, "/research-tasks/") {
		id, err := uuid.Parse(raw)
		if err != nil || !queryOwner(r, db, "SELECT user_id FROM research_tasks WHERE id = ? AND deleted_at IS NULL", id, owner) {
			return false
		}
	}
	return true
}

func queryOwner(r *http.Request, db *gorm.DB, query string, id, expected uuid.UUID) bool {
	var actual uuid.UUID
	err := db.WithContext(r.Context()).Raw(query, id).Scan(&actual).Error
	return err == nil && actual != uuid.Nil && actual == expected
}

func ownershipLookupError(err error) bool {
	return err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
}
