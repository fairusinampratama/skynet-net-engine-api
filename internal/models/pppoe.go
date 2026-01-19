package models

// PPPoESecret represents a PPPoE account from MikroTik
type PPPoESecret struct {
	Name          string
	Profile       string
	RemoteAddress string
	Disabled      bool
}

// UserWithStatus represents a user with their profile and status
type UserWithStatus struct {
	Username string `json:"username"`
	Status   string `json:"status"` // "online" or "offline"
	IP       string `json:"ip,omitempty"`
	Uptime   string `json:"uptime,omitempty"`
	Profile  string `json:"profile,omitempty"`
	Enabled  bool   `json:"enabled"`
	BytesIn  int64  `json:"bytes_in"`
	BytesOut int64  `json:"bytes_out"`
}
