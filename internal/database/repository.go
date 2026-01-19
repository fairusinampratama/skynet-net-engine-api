package database

import (
	"skynet-net-engine-api/internal/models"
	"go.uber.org/zap"
	"skynet-net-engine-api/pkg/logger"
	"database/sql"
)

func GetAllRouters() ([]models.Router, error) {
	rows, err := DB.Query("SELECT id, name, host, port, username, password FROM routers")
	if err != nil {
		logger.Error("Failed to fetch routers", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var routers []models.Router
	for rows.Next() {
		var r models.Router
		if err := rows.Scan(&r.ID, &r.Name, &r.Host, &r.Port, &r.Username, &r.Password); err != nil {
			logger.Error("Failed to scan router row", zap.Error(err))
			continue
		}
		routers = append(routers, r)
	}

	return routers, nil
}

// UpsertUser inserts or updates a PPPoE user
func UpsertUser(username string, routerID int, profile string, remoteAddress string, isEnabled bool) error {
	// Derive is_isolated from profile name
	isIsolated := profile == "isolirebilling"
	
	query := `
		INSERT INTO pppoe_users (username, router_id, profile, remote_address, is_enabled, is_isolated)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			profile = VALUES(profile),
			remote_address = VALUES(remote_address),
			is_enabled = VALUES(is_enabled),
			is_isolated = VALUES(is_isolated),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := DB.Exec(query, username, routerID, profile, remoteAddress, isEnabled, isIsolated)
	if err != nil {
		logger.Error("Failed to upsert user", zap.String("user", username), zap.Error(err))
	}
	return err
}

// DBUserWithPrev includes previous profile info
type DBUserWithPrev struct {
	Username        string
	Profile         string
	PreviousProfile string
	RemoteAddress   string
	IsEnabled       bool
	IsIsolated      bool
}

// GetUserByIP fetches a user by their remote address
func GetUserByIP(ip string) (DBUserWithPrev, error) {
	var u DBUserWithPrev
	var prev sql.NullString
	
	query := "SELECT username, profile, previous_profile, remote_address, is_enabled, is_isolated FROM pppoe_users WHERE remote_address = ?"
	err := DB.QueryRow(query, ip).Scan(&u.Username, &u.Profile, &prev, &u.RemoteAddress, &u.IsEnabled, &u.IsIsolated)
	
	if prev.Valid {
		u.PreviousProfile = prev.String
	}
	
	if err != nil {
		logger.Error("Failed to get user by IP", zap.String("ip", ip), zap.Error(err))
	}
	
	return u, err
}

// DBUser represents a user record from the database
type DBUser struct {
	Profile       string
	RemoteAddress string
	IsEnabled     bool
	IsIsolated    bool
}

// GetUsersByRouter fetches all users for a specific router (including disabled ones)
func GetUsersByRouter(routerID int) (map[string]DBUser, error) {
	rows, err := DB.Query("SELECT username, profile, remote_address, is_enabled, is_isolated FROM pppoe_users WHERE router_id = ?", routerID)
	if err != nil {
		logger.Error("Failed to fetch users by router", zap.Int("router_id", routerID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	users := make(map[string]DBUser) // username -> DBUser
	for rows.Next() {
		var username, profile string
		var remoteAddress sql.NullString // Handle potential NULLs
		var isEnabled bool
		var isIsolated sql.NullBool // Handle potential NULL (default false if null)
		
		if err := rows.Scan(&username, &profile, &remoteAddress, &isEnabled, &isIsolated); err != nil {
			logger.Error("Scan error", zap.Error(err))
			continue
		}
		
		isolated := false
		if isIsolated.Valid {
			isolated = isIsolated.Bool
		}

		users[username] = DBUser{
			Profile:       profile,
			RemoteAddress: remoteAddress.String,
			IsEnabled:     isEnabled,
			IsIsolated:    isolated,
		}
	}

	return users, nil
}

// UpdateUserIsolationStatus updates the is_isolated status for a user based on remote_address (IP)
// UpdateUserIsolationStatus switches the profile in the DB and toggles is_isolated
func UpdateUserIsolationStatus(ip string, isIsolated bool, targetProfile string) error {
	var query string
	if isIsolated {
		// Encapsulate: Save current profile as 'previous', set new to 'ISOLATED'
		query = `UPDATE pppoe_users 
				 SET previous_profile = profile, profile = ?, is_isolated = TRUE, updated_at = CURRENT_TIMESTAMP 
				 WHERE remote_address = ?`
		_, err := DB.Exec(query, targetProfile, ip)
		return err
	} else {
		// Restore: Set profile = previous_profile (if exists), set is_isolated = FALSE
		// If previous_profile is null, we might need a fallback, but for now specific query:
		query = `UPDATE pppoe_users 
				 SET profile = COALESCE(previous_profile, ?), is_isolated = FALSE, updated_at = CURRENT_TIMESTAMP 
				 WHERE remote_address = ?`
		// We pass targetProfile as fallback (e.g. "default" or "10M")
		_, err := DB.Exec(query, targetProfile, ip)
		return err
	}
}
