package account

import (
	"encoding/json"
	"net/http"
	"strings"
)

// SignUpRoute creates a tenant and its first user.
func SignUpRoute(w http.ResponseWriter, r *http.Request) {
	var payload SignUpPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload.TenantName = strings.TrimSpace(payload.TenantName)
	payload.UserName = strings.TrimSpace(payload.UserName)
	payload.Email = strings.TrimSpace(strings.ToLower(payload.Email))
	if payload.TenantName == "" || payload.UserName == "" || payload.Email == "" || payload.Password == "" {
		http.Error(w, "tenant_name, user_name, email, and password are required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(payload.Email, "@") {
		http.Error(w, "valid email is required", http.StatusBadRequest)
		return
	}
	if len(payload.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	account, err := SignUpServices(r.Context(), payload.TenantName, payload.UserName, payload.Email, payload.Password)
	if err != nil {
		http.Error(w, "error signing up tenant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

// LoginRoute restores an existing account by email.
func LoginRoute(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload.Email = strings.TrimSpace(strings.ToLower(payload.Email))
	if payload.Email == "" || payload.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(payload.Email, "@") {
		http.Error(w, "valid email is required", http.StatusBadRequest)
		return
	}

	account, err := LoginServices(r.Context(), payload.Email, payload.Password)
	if err != nil {
		http.Error(w, "account not found", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

// CreateTeamRoute creates a new team for the signed-up account.
func CreateTeamRoute(w http.ResponseWriter, r *http.Request) {
	var payload CreateTeamPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload.TeamName = strings.TrimSpace(payload.TeamName)
	if payload.TeamName == "" {
		http.Error(w, "team_name is required", http.StatusBadRequest)
		return
	}

	session, err := SessionContextFromRequest(r.Context(), r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	team, err := CreateTeamServices(r.Context(), session.TenantPublicId, session.UserPublicId, payload.TeamName)
	if err != nil {
		http.Error(w, "error creating team", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(team)
}

// ListTeamsRoute lists teams for the signed-up account.
func ListTeamsRoute(w http.ResponseWriter, r *http.Request) {
	session, err := SessionContextFromRequest(r.Context(), r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	teams, err := ListTeamsServices(r.Context(), session.TenantPublicId, session.UserPublicId)
	if err != nil {
		http.Error(w, "error listing teams", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teams)
}
