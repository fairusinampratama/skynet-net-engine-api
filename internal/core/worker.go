package core

import (
	"fmt"
	"time"
	"sync"
	"skynet-net-engine-api/internal/mikrotik"
	"skynet-net-engine-api/internal/models"
	"skynet-net-engine-api/internal/database"
	"skynet-net-engine-api/pkg/logger"
	"go.uber.org/zap"
)

type Worker struct {
	Router   models.Router
	CmdChan  chan Command
	Client        *mikrotik.Client
	MetricsClient *mikrotik.Client // dedicated connection for heavy reads
	IsOnline      bool
	
	// Synchronization
	once sync.Once
	wg   *sync.WaitGroup

	// Cache
	ActiveUsers    []models.ActiveUser
	SystemResource *models.SystemResource
	Lock           sync.RWMutex
	IsScraping     bool // Simple guard for background jobs
}

func (w *Worker) safeRefreshMetrics() {
	w.Lock.Lock()
	if w.IsScraping {
		w.Lock.Unlock()
		return
	}
	w.IsScraping = true
	w.Lock.Unlock()

	defer func() {
		w.Lock.Lock()
		w.IsScraping = false
		w.Lock.Unlock()
	}()

	w.refreshMetrics()
}

func NewWorker(r models.Router, wg *sync.WaitGroup) *Worker {
	return &Worker{
		Router:  r,
		CmdChan: make(chan Command, 1000), // Increased buffer for stability
		wg:      wg,
	}
}

// Start begins the persistent loop
func (w *Worker) Start() {
	// Ensure we always mark as done eventually
	signalReady := func() {
		w.once.Do(func() {
			if w.wg != nil {
				w.wg.Done()
			}
		})
	}

	go w.metricsLoop() // Start background metrics/keepalive

	for {
		// 1. Try to Connect (Primary)
		logger.Info("Dialing router (Primary)...", zap.String("host", w.Router.Host), zap.Int("port", w.Router.Port), zap.String("user", w.Router.Username))
		client, err := mikrotik.NewClient(w.Router)
		
		if err != nil {
			logger.Error("Primary Connection failed, retrying in 5s...", zap.String("host", w.Router.Host), zap.Error(err))
			w.IsOnline = false
			signalReady() 
			time.Sleep(5 * time.Second)
			continue
		}

		// 2. Try to Connect (Metrics)
		// We allow this to fail initially, but we should try to reconnect it if possible
		logger.Info("Dialing router (Metrics)...", zap.String("host", w.Router.Host))
		metricsClient, errM := mikrotik.NewClient(w.Router)
		if errM != nil {
			logger.Warn("Metrics Connection failed - Metrics will be disabled until restart/reconnect", zap.Error(errM))
		}
		
		// 3. Connected!
		w.Client = client
		w.MetricsClient = metricsClient
		w.IsOnline = true
		logger.Info("Router Connected!", zap.String("host", w.Router.Host))
		SendWebhook("router.up", w.Router.ID, w.Router.Host, nil)

		// 4. WARMUP
		logger.Info("Warming up cache...", zap.String("host", w.Router.Host))
		// We can now run this in background immediately because it uses a separate client!
		go w.safeRefreshMetrics()
		
		w.CmdChan <- Command{Type: CmdSync} // Initial Sync
		
		signalReady() // Signal ready

		// 5. Command Loop (Blocks until connection dies)
		w.handleCommands()

		// 6. Cleanup
		logger.Warn("Router Disconnected. Cleaning up...", zap.String("host", w.Router.Host))
		if w.IsOnline {
			SendWebhook("router.down", w.Router.ID, w.Router.Host, "Connection lost")
		}
		w.IsOnline = false
		if w.Client != nil {
			w.Client.Close()
		}
		if w.MetricsClient != nil {
			w.MetricsClient.Close()
		}
		
		// 7. Backoff
		time.Sleep(3 * time.Second)
	}
}

func (w *Worker) handleCommands() {
	for cmd := range w.CmdChan {
		logger.Info("Received command", zap.String("type", string(cmd.Type)))

		var err error
		
		// Retry Logic
		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				logger.Warn("Retrying command...", zap.String("type", string(cmd.Type)), zap.Int("attempt", attempt+1))
				time.Sleep(2 * time.Second)
			}

			err = nil
			switch cmd.Type {
			case CmdSync:
				// Run in background on Metrics connection to avoid blocking command loop
				go w.syncSecrets()
				// We don't block here. Command returns immediately.
				if cmd.Result != nil {
					cmd.Result <- "Scheduled"
				}
			
			case CmdCreateSecret:
				payload := cmd.Payload.(map[string]string)
				err = w.Client.AddSecret(payload["user"], payload["password"], payload["profile"], payload["local_ip"], payload["remote_ip"], payload["comment"])
			
			case CmdUpdateSecret:
				payload := cmd.Payload.(map[string]string)
				err = w.Client.SetSecretProfile(payload["user"], payload["profile"])

			case CmdIsolate:
				payload := cmd.Payload.(map[string]string)
				isIsolated := payload["action"] == "add"
				ip := payload["ip"]
				
				// 1. DETERMINE TARGET PROFILE
				var targetProfile string
				if isIsolated {
					targetProfile = "isolirebilling" // Found existing profile!
				} else {
					targetProfile = "10M" // Fallback
				}

				// 2. FIND USERNAME BY IP (Since SetSecretProfile needs Username)
				// We need to look this up from our Cache or DB because the request might only have IP
				// Ideally the request SHOULD have username, but let's look it up safest way
				username := ""
				// Try Cache first
				w.Lock.RLock()
				for _, u := range w.ActiveUsers {
					if u.Address == ip {
						username = u.Name
						break
					}
				}
				w.Lock.RUnlock()
				
				// If not in cache (offline), query DB? Or we require Username in Payload?
				// Let's assume we need to query DB if not in active cache.
				if username == "" {
					// Quick DB lookup
					dbUser, _ := database.GetUserByIP(ip) // Helper we might need to add or just query
					username = dbUser.Username
				}

				if username == "" {
					logger.Error("Cannot Isolate: User not found for IP", zap.String("ip", ip))
					err = fmt.Errorf("user not found")
					break
				}

				// 3. EXECUTE: CHANGE PROFILE
				if isIsolated {
					err = w.Client.SetSecretProfile(username, targetProfile)
				} else {
					// RESTORE
					// Try to restore previous profile from DB
					dbUser, _ := database.GetUserByIP(ip)
					if dbUser.PreviousProfile != "" {
						err = w.Client.SetSecretProfile(username, dbUser.PreviousProfile)
					} else {
						err = w.Client.SetSecretProfile(username, targetProfile) // Use targetProfile (10M fallback)
					}
				}

				// 4. EXECUTE: KICK USER (To force profile reload)
				if err == nil {
					_ = w.Client.RemoveActiveUser(username) // Ignore error if they are already offline
					
					// 5. UPDATE DB STATE
					targetForDB := "isolirebilling"
					if !isIsolated { targetForDB = "10M" } // Just for fallback argument
					database.UpdateUserIsolationStatus(ip, isIsolated, targetForDB)
					
					logger.Info("Isolation (Profile) Updated", zap.String("user", username), zap.Bool("isolated", isIsolated))
				}

			case CmdGetTraffic:
				target := cmd.Payload.(string)
				stats, errT := w.Client.GetQueueTraffic(target)
				if errT != nil {
					err = errT
				} else {
					if cmd.Result != nil { cmd.Result <- stats }
					return
				}

			case CmdBackup:
				err = w.Client.RunBackup(cmd.Payload.(string))

			case CmdRefreshMetrics:
				// Run in background to avoid blocking critical commands (Sync/Isolate)
				// Use atomic/bool check to prevent stacking
				go w.safeRefreshMetrics()
			}
			
                if err == nil {
				break
			}
		}

		if cmd.Error != nil {
			cmd.Error <- err
		}
		if err == nil {
			if cmd.Type != CmdSync && cmd.Type != CmdGetTraffic {
				if cmd.Result != nil {
					cmd.Result <- "Success"
				}
			}
		}
	}
}

func (w *Worker) metricsLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !w.IsOnline {
			continue
		}
		// Non-blocking send to avoid clogging queue during high load/instability
		select {
		case w.CmdChan <- Command{Type: CmdRefreshMetrics}:
		default:
			// Drop metrics update if queue is full
		}
	}
}

func (w *Worker) syncSecrets() {
	if w.MetricsClient == nil {
		return
	}

	// Use Sync Guard if needed? Secrets sync is rare (startup), so maybe not strict.
	// But let's use the scraping guard to be safe
	w.Lock.Lock()
	if w.IsScraping {
		w.Lock.Unlock()
		return
	}
	w.IsScraping = true
	w.Lock.Unlock()
	
	defer func() {
		w.Lock.Lock()
		w.IsScraping = false
		w.Lock.Unlock()
	}()

	secrets, err := w.MetricsClient.GetAllSecrets()
	if err != nil {
		logger.Error("Failed to sync secrets", zap.Error(err))
		return
	}

	count := 0
	for _, s := range secrets {
		if dbErr := database.UpsertUser(s.Name, w.Router.ID, s.Profile, s.RemoteAddress, !s.Disabled); dbErr == nil {
			count++
		}
	}
	logger.Info("Synced Secrets to DB", zap.Int("synced", count))
}

func (w *Worker) refreshMetrics() {
	if w.MetricsClient == nil {
		// Try to reconnect lazily
		logger.Info("Metrics Client offline, attempting reconnect...", zap.String("host", w.Router.Host))
		mc, err := mikrotik.NewClient(w.Router)
		if err == nil {
			w.MetricsClient = mc
			logger.Info("Metrics Client Reconnected!", zap.String("host", w.Router.Host))
		} else {
			logger.Warn("Metrics Reconnect Failed", zap.Error(err))
			return
		}
	}

	users, err := w.MetricsClient.GetActiveUsers()
	if err != nil {
		logger.Error("Failed to fetch active users", zap.String("host", w.Router.Host), zap.Error(err))
		// If metrics connection dies, we should probably reconnect? 
		// For now, logging error is enough.
	} 
	
	res, errRes := w.MetricsClient.GetSystemResource()
	if errRes != nil {
		// logging error optional
	}

	// Update Cache
	w.Lock.Lock()
	if err == nil {
		w.ActiveUsers = users
		logger.Info("Worker Cache Updated", zap.String("router", w.Router.Name), zap.Int("active_users", len(w.ActiveUsers)))
	}
	if errRes == nil {
		w.SystemResource = res
	}
	w.Lock.Unlock()
	
	// logger.Info("Metrics refreshed", zap.String("host", w.Router.Host), zap.Int("users", len(users)))
}
