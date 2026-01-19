package mikrotik

import (
	"fmt"
	"net"
	"time"
	"strconv"
	"strings"
	"skynet-net-engine-api/pkg/logger"
	"skynet-net-engine-api/internal/models"
	
	"go.uber.org/zap" // Explicit import provided
	"github.com/go-routeros/routeros"
)

type Client struct {
	Conn *routeros.Client
	Router models.Router
}

func NewClient(r models.Router) (*Client, error) {
	address := fmt.Sprintf("%s:%d", r.Host, r.Port)
	
	// Custom Dialer with 15s Timeout (Fix for slow tunnels)
	dialer := net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	// Create wrapper
	c, err := routeros.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Login manually
	if err := c.Login(r.Username, r.Password); err != nil {
		c.Close()
		return nil, err
	}
	
	return &Client{
		Conn: c,
		Router: r,
	}, nil
}

func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// KeepAlive sends a lightweight command to check if connection is alive
func (c *Client) KeepAlive() error {
	_, err := c.Conn.Run("/system/identity/print")
	return err
}

func (c *Client) AddSecret(user, password, profile, localIP, remoteIP, comment string) error {
	// Construct the command
	cmd := []string{
		"/ppp/secret/add",
		"=name=" + user,
		"=password=" + password,
		"=profile=" + profile,
		"=comment=" + comment,
	}

	if localIP != "" {
		cmd = append(cmd, "=local-address="+localIP)
	}
	if remoteIP != "" {
		cmd = append(cmd, "=remote-address="+remoteIP)
	}

	_, err := c.Conn.RunArgs(cmd)
	return err
}

func (c *Client) SetSecretProfile(user, newProfile string) error {
	// Find the secret first to get ID? Or use specific query
	// RouterOS > 6.45 supports 'set [ find name=X ] profile=Y'
	
	// Doing it reliably via ID lookup is safer usually, but 'find where name=X' works
	// We will try the direct set command with a query
	
	// 1. Find ID
	res, err := c.Conn.Run("/ppp/secret/print", "?name="+user, "=.proplist=.id")
	if err != nil {
		return err
	}
	if len(res.Re) == 0 {
		return fmt.Errorf("user not found")
	}
	id := res.Re[0].Map[".id"]

	// 2. Set Profile
	_, err = c.Conn.Run("/ppp/secret/set", "=.id="+id, "=profile="+newProfile)
	return err
}

// GetAllSecrets fetches all PPPoE secrets from the router
func (c *Client) GetAllSecrets() ([]models.PPPoESecret, error) {
	res, err := c.Conn.Run("/ppp/secret/print", "=.proplist=name,profile,remote-address,disabled")
	if err != nil {
		return nil, err
	}

	var secrets []models.PPPoESecret
	for _, re := range res.Re {
		disabled := re.Map["disabled"] == "true"
		secrets = append(secrets, models.PPPoESecret{
			Name:          re.Map["name"],
			Profile:       re.Map["profile"],
			RemoteAddress: re.Map["remote-address"],
			Disabled:      disabled,
		})
	}

	return secrets, nil
}

func (c *Client) AddAddressList(ip, list, comment string) error {
	_, err := c.Conn.Run(
		"/ip/firewall/address-list/add",
		"=address="+ip,
		"=list="+list,
		"=comment="+comment,
	)
	return err
}

func (c *Client) RemoveAddressList(ip, list string) error {
	// Find ID first
	res, err := c.Conn.Run("/ip/firewall/address-list/print", "?address="+ip, "?list="+list, "=.proplist=.id")
	if err != nil {
		return err
	}
	
	for _, re := range res.Re {
		id := re.Map[".id"]
		c.Conn.Run("/ip/firewall/address-list/remove", "=.id="+id)
	}
	// Ignore if not found, idempotent
	return nil
}

func (c *Client) GetActiveUsers() ([]models.ActiveUser, error) {
	// Optimizing query - include bytes for traffic data
	res, err := c.Conn.Run("/ppp/active/print", "=.proplist=name,address,caller-id,uptime,limit-bytes-in,limit-bytes-out")
	if err != nil {
		logger.Error("Mikrotik Query Failed", zap.Error(err))
		return nil, err
	}

	logger.Info("Mikrotik Raw Response", zap.Int("count", len(res.Re)), zap.String("router", c.Router.Name))

	users := make([]models.ActiveUser, 0)
	for _, re := range res.Re {
		// Parse bytes (default to 0 if empty)
		bytesIn, _ := strconv.ParseInt(re.Map["limit-bytes-in"], 10, 64)
		bytesOut, _ := strconv.ParseInt(re.Map["limit-bytes-out"], 10, 64)
		
		users = append(users, models.ActiveUser{
			Name:     re.Map["name"],
			Address:  re.Map["address"],
			CallerID: re.Map["caller-id"],
			Uptime:   re.Map["uptime"],
			RouterID: c.Router.ID,
			BytesIn:  bytesIn,
			BytesOut: bytesOut,
		})
	}
	return users, nil
}

func (c *Client) GetSystemResource() (*models.SystemResource, error) {
	res, err := c.Conn.Run("/system/resource/print")
	if err != nil {
		return nil, err
	}
	if len(res.Re) == 0 {
		return nil, fmt.Errorf("no data")
	}
	
	m := res.Re[0].Map
	// Parsing integers in production should use strconv, simplified here for MVP
	// RouterOS usually returns numbers for memory
	
	// Helper for parsing
	parseInt := func(s string) int64 {
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}

	return &models.SystemResource{
		Uptime:      m["uptime"],
		CPU:         m["cpu-load"],
		BoardName:   m["board-name"],
		Version:     m["version"],
		TotalMemory: parseInt(m["total-memory"]),
		FreeMemory:  parseInt(m["free-memory"]),
	}, nil
}

func (c *Client) GetQueueTraffic(target string) (*models.TrafficStats, error) {
	// Try exact match first
	res, err := c.Conn.Run("/queue/simple/print", "?name="+target, "=.proplist=rate,name")
	if err != nil {
		return nil, err
	}
	
	var queueName, rawRate string
	
	// If exact match fails, scan all queues for substring match
	if len(res.Re) == 0 {
		logger.Warn("Exact match failed, scanning all queues...", zap.String("target", target))
		res, err = c.Conn.Run("/queue/simple/print", "=.proplist=rate,name")
		if err != nil {
			return nil, err
		}

		// Find first queue containing target username
		found := false
		for _, re := range res.Re {
			if strings.Contains(re.Map["name"], target) {
				queueName = re.Map["name"]
				rawRate = re.Map["rate"]
				found = true
				logger.Info("Found queue via scan", zap.String("target", target), zap.String("found", queueName))
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("queue not found")
		}
	} else {
		// Exact match found
		queueName = res.Re[0].Map["name"]
		rawRate = res.Re[0].Map["rate"]
	}
	
	// Rate comes like "rx/tx" (e.g. "1500/5000" in bits per second)
	// Parse it
	
	// Basic parsing
	var rx, tx int64
	fmt.Sscanf(rawRate, "%d/%d", &rx, &tx)
	
	return &models.TrafficStats{
		Name: target,
		RX:   rx,
		TX:   tx,
	}, nil
}

func (c *Client) RunBackup(name string) error {
	_, err := c.Conn.Run("/system/backup/save", "=name="+name)
	return err
}

func (c *Client) RemoveActiveUser(username string) error {
	// 1. Find Session ID
	res, err := c.Conn.Run("/ppp/active/print", "?name="+username, "=.proplist=.id")
	if err != nil {
		return err
	}
	
	// 2. Remove all matching sessions (usually just one)
	for _, re := range res.Re {
		id := re.Map[".id"]
		c.Conn.Run("/ppp/active/remove", "=.id="+id)
	}
	return nil
}
