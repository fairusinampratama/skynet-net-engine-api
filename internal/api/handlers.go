package api

import (
	"net/http"
	"time"
	"strconv"
	
	"skynet-net-engine-api/internal/core"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/internal/models"
	
	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary      Get API Health
// @Description  Checks if the NetEngine Muscle is alive
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "muscle": "alive"})
}

// SyncRouter godoc
// @Summary      Force Sync Router
// @Description  Triggers an immediate sync of active users
// @Tags         Control
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Router ID"
// @Success      200  {object}  map[string]string
// @Router       /sync/{id} [post]
func SyncRouter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Router ID"})
		return
	}

	worker := core.GlobalPool.GetWorker(id)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router Not Found"})
		return
	}

	// Send Async Command
	cmd := core.Command{
		Type: core.CmdSync,
		Result: make(chan interface{}),
	}
	
	// Non-blocking send
	select {
	case worker.CmdChan <- cmd:
		// Wait for ack
		<-cmd.Result
		c.JSON(http.StatusOK, gin.H{"status": "Sync command sent"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Worker busy or offline"})
	}
}

func KickUser(c *gin.Context) {
	// TODO: Implement Kick Logic parsing body
	c.JSON(http.StatusNotImplemented, gin.H{"status": "TODO"})
}

// CreateSecret godoc
// @Summary      Create PPP Secret
// @Description  Adds a new PPPoE secret to a router
// @Tags         Bridge
// @Accept       json
// @Produce      json
// @Param        request body CreateSecretRequest true "Secret Data"
// @Success      201  {object}  map[string]string
// @Router       /secret [post]
func CreateSecret(c *gin.Context) {
	var req CreateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Implementation pending (Will require CommandType expansion)
	// c.JSON(http.StatusAccepted, gin.H{"status": "Queued", "user": req.User})

	workerID := 1 // TODO: Logic to determine which router (Round robin or specified)
	// For creating a secret, we usually need to know WHICH router to put it on.
	// Assuming for now we put it on all? Or getting RouterID from request context? 
	// USE CASE: "Bridge". The Laravel app should probably specify target router.
	// I'll add RouterID to the request struct in a future iteration, for now hardcoding to 1 or loop.
	
	// Assuming we just send to Worker #1 for MVP
	worker := core.GlobalPool.GetWorker(workerID)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router 1 Not Found (MVP Hardcode)"})
		return
	}

	cmd := core.Command{
		Type: core.CmdCreateSecret,
		Payload: map[string]string{
			"user": req.User, "password": req.Password, "profile": req.Profile,
			"local_ip": req.LocalIP, "remote_ip": req.RemoteIP, "comment": req.Comment,
		},
		Error: make(chan error),
		Result: make(chan interface{}),
	}

	worker.CmdChan <- cmd
	
	err := <-cmd.Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"status": "Secret Created", "user": req.User})
}

// UpdatePlan godoc
// @Summary      Change Plan
// @Description  Updates the profile of an existing secret
// @Tags         Bridge
// @Accept       json
// @Produce      json
// @Param        user  path  string  true  "Username"
// @Param        request body UpdatePlanRequest true "Plan Data"
// @Success      200  {object}  map[string]string
// @Router       /secret/{user} [put]
func UpdatePlan(c *gin.Context) {
	user := c.Param("user") // Using username as ID
	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workerID := 1 // MVP Hardcode
	worker := core.GlobalPool.GetWorker(workerID)
	
	cmd := core.Command{
		Type: core.CmdUpdateSecret,
		Payload: map[string]string{
			"user": user,
			"profile": req.Profile,
		},
		Error: make(chan error),
		Result: make(chan interface{}),
	}
	
	worker.CmdChan <- cmd
	err := <-cmd.Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Plan Updated", "user": user, "profile": req.Profile})
}

// IsolateUser godoc
// @Summary      Isolate User
// @Description  Adds or removes a user from a Firewall Address List
// @Tags         Advanced
// @Accept       json
// @Produce      json
// @Param        request body IsolateRequest true "Isolation Data"
// @Success      200  {object}  map[string]string
// @Router       /isolate [post]
func IsolateUser(c *gin.Context) {
	var req IsolateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if req.List == "" {
		req.List = "ISOLATED"
	}

	workerID := 1 // Default MVP
	if req.RouterID != 0 {
		workerID = req.RouterID
	}
	worker := core.GlobalPool.GetWorker(workerID)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}
	
	isSync := c.Query("sync") == "true"
	
	cmd := core.Command{
		Type: core.CmdIsolate,
		Payload: map[string]string{
			"ip": req.IP,
			"list": req.List,
			"action": req.Action,
			"comment": req.Comment,
		},
		Error: nil,
		Result: nil,
	}

	if isSync {
		cmd.Error = make(chan error)
		cmd.Result = make(chan interface{})
	}
	
	// Send Command
	select {
	case worker.CmdChan <- cmd:
		if isSync {
			// SYNC MODE: Block and Wait
			select {
			case err := <-cmd.Error:
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "status": "Router Connection Failed"})
			case <-cmd.Result:
				c.JSON(http.StatusOK, gin.H{
					"status": "Success", 
					"message": "User isolation updated successfully.",
					"isolated": req.Action == "add",
				})
			case <-time.After(30 * time.Second):
				c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Timeout waiting for router response"})
			}
		} else {
			// ASYNC MODE: Return immediately
			c.JSON(http.StatusAccepted, gin.H{
				"status": "Isolate Request Queued", 
				"message": "Command queued for execution.",
				"mode": "async",
			})
		}
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Worker queue full, please try again later"})
	}
}

func GetTargets(c *gin.Context) {
	requestID := c.Query("request_id") // Optional tracking
	targets := core.GlobalPool.GetAllTargets()
	
	// Add metadata if needed, but array is efficient
	c.JSON(http.StatusOK, targets)
	
	// Logging heavy usage
	if requestID != "" {
		// logger.Debug("Targets fetched", zap.String("rid", requestID), zap.Int("count", len(targets)))
	}
}

// GetAllUsers godoc
// @Summary      Get All Users with Status
// @Description  Returns all PPPoE users from database with real-time connection status (connected, isolated, offline)
// @Tags         Monitoring
// @Accept       json
// @Produce      json
// @Param        id path int true "Router ID"
// @Success      200  {array}  models.UserWithStatus
// @Router       /router/{id}/users [get]
func GetAllUsers(c *gin.Context) {
	idStr := c.Param("id")
	routerID, _ := strconv.Atoi(idStr)
	showAll := c.Query("show_all") == "true"
	
	// 1. Get all secrets from DB (synced from /ppp/secret for THIS router)
	dbUsers, err := database.GetUsersByRouter(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	
	// 2. Get active sessions from THIS ROUTER ONLY
	worker := core.GlobalPool.GetWorker(routerID)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}
	
	worker.Lock.RLock()
	activeUsers := worker.ActiveUsers
	worker.Lock.RUnlock()
	
	// 3. Build a lookup map for quick online check (THIS router only)
	activeMap := make(map[string]models.ActiveUser)
	for _, au := range activeUsers {
		activeMap[au.Name] = au
	}
	
	// 4. Build response with online/offline status based on uptime
	result := []models.UserWithStatus{}
	
	for username, dbUser := range dbUsers {
		// Check if user is in THIS router's active list
		activeSession, isOnline := activeMap[username]
		
		status := "offline"
		uptime := ""
		ip := dbUser.RemoteAddress
		var bytesIn, bytesOut int64
		
		if isOnline {
			status = "online"
			uptime = activeSession.Uptime
			ip = activeSession.Address
			bytesIn = activeSession.BytesIn
			bytesOut = activeSession.BytesOut
		}
		
		// Filter: Hide offline users by default AND hide ghosts
		// Strategy: If user is offline AND router is Randuagung (ID=1), hide them?
		// Better: Client side toggle.
		// User Request: "can we filter the ghost to not shown?"
		// Let's implement strict filtering:
		
		if !showAll && !isOnline && status != "isolated" {
			// For Router 1 (Hub), the ghosts are offline customer profiles
			// Showing 560 offline users is useless.
			continue
		}
		
		result = append(result, models.UserWithStatus{
			Username: username,
			Status:   status,
			Profile:  dbUser.Profile,
			IP:       ip,
			Uptime:   uptime,
			Enabled:  dbUser.IsEnabled,
			BytesIn:  bytesIn,
			BytesOut: bytesOut,
		})
	}
	
	c.JSON(http.StatusOK, result)
}


// GetRouterHealth godoc
// @Summary      Get Router Health
// @Description  Returns CPU, Memory, and Uptime
// @Tags         Monitoring
// @Accept       json
// @Produce      json
// @Param        id   path   int  true  "Router ID"
// @Success      200  {object}  models.SystemResource
// @Router       /router/{id}/health [get]
func GetRouterHealth(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	worker := core.GlobalPool.GetWorker(id)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}
	
	worker.Lock.RLock()
	defer worker.Lock.RUnlock()
	
	if worker.SystemResource == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "No data yet", "next_retry": "10s"})
		return
	}
	c.JSON(http.StatusOK, worker.SystemResource)
}

func GetUserTraffic(c *gin.Context) {
	routerIDStr := c.Param("id")
	routerID, _ := strconv.Atoi(routerIDStr)
	user := c.Query("user")
	
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query param 'user' required"})
		return
	}

	worker := core.GlobalPool.GetWorker(routerID)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}

	cmd := core.Command{
		Type: core.CmdGetTraffic,
		Payload: user,
		Error: make(chan error),
		Result: make(chan interface{}),
	}
	
	select {
	case worker.CmdChan <- cmd:
		select {
		case res := <-cmd.Result:
			c.JSON(http.StatusOK, res)
		case err := <-cmd.Error:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case <-time.After(10 * time.Second):
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Timeout waiting for router"})
		}
	case <-time.After(2 * time.Second):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Router queue full or offline"})
	}
}

func TriggerBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	worker := core.GlobalPool.GetWorker(id)
	if worker == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Router not found"})
		return
	}

	filename := "netengine_backup_" + time.Now().Format("20060102_150405")
	cmd := core.Command{
		Type: core.CmdBackup,
		Payload: filename,
		Error: make(chan error),
		Result: make(chan interface{}),
	}
	
	worker.CmdChan <- cmd
	err := <-cmd.Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "Backup created", "file": filename + ".backup"})
}

func GetRouters(c *gin.Context) {
	routers, err := database.GetAllRouters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch routers"})
		return
	}
	
	// Sanitize passwords
	for i := range routers {
		routers[i].Password = ""
	}
	
	c.JSON(http.StatusOK, routers)
}

