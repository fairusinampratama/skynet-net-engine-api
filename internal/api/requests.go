package api

type IsolateRequest struct {
	IP       string `json:"ip" binding:"required"`
	Action   string `json:"action" binding:"required,oneof=add remove"` // add or remove
	List     string `json:"list"`                                       // Default to "ISOLATED" if empty
	Comment  string `json:"comment"`
	RouterID int    `json:"router_id"` // Optional, defaults to finding user or hardcoded 1
}
