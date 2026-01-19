package models

type ActiveUser struct {
	Name      string `json:"name"`
	Address   string `json:"address"`   // IP Address
	CallerID  string `json:"caller_id"` // MAC Address
	Uptime    string `json:"uptime"`
	RouterID  int    `json:"router_id"`
	BytesIn   int64  `json:"bytes_in"`  // Download bytes
	BytesOut  int64  `json:"bytes_out"` // Upload bytes
}
