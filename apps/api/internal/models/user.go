package models

import "time"

// User represents a user in the system.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`    // never expose password hash
	Role         string    `json:"role"` // "admin" or "user"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidRoles contains the valid user roles.
var ValidRoles = []string{"admin", "user"}

// IsValidRole checks if a role string is valid.
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}

// UserPreferences stores per-user preference settings.
type UserPreferences struct {
	UserID           string    `json:"user_id"`
	Language         string    `json:"language"`
	SidebarCollapsed bool      `json:"sidebar_collapsed"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ChangePasswordRequest is the DTO for password change.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// UpdatePreferencesRequest is the DTO for updating user preferences.
type UpdatePreferencesRequest struct {
	Language         *string `json:"language,omitempty"`
	SidebarCollapsed *bool   `json:"sidebar_collapsed,omitempty"`
}
