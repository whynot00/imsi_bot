package middleware

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/repo"
)

const accessDeniedHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Access Denied</title>
<style>
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Unbounded:wght@400;700&display=swap');
*{box-sizing:border-box;margin:0;padding:0}
body{
  background:#0a0a0f;
  color:#c8c8d8;
  font-family:'JetBrains Mono',monospace;
  min-height:100vh;
  display:flex;
  align-items:center;
  justify-content:center;
  flex-direction:column;
  gap:24px;
  padding:40px 24px;
}
body::before{
  content:'';position:fixed;inset:0;
  background:repeating-linear-gradient(0deg,transparent,transparent 2px,rgba(0,0,0,.06) 2px,rgba(0,0,0,.06) 4px);
  pointer-events:none;z-index:1000;
}
.icon{font-size:48px;opacity:.6}
.title{font-family:'Unbounded',sans-serif;font-size:14px;letter-spacing:.15em;color:#ff3e6c}
.code{font-size:64px;font-weight:700;color:rgba(255,62,108,.15);letter-spacing:.2em}
.msg{font-size:12px;color:#555570;text-align:center;line-height:1.8;max-width:400px}
.ip{
  margin-top:8px;
  padding:8px 16px;
  background:#111118;
  border:1px solid #1e1e2e;
  border-radius:6px;
  font-size:11px;
  color:#555570;
}
</style>
</head>
<body>
<div class="code">403</div>
<div class="title">ACCESS DENIED</div>
<div class="msg">
  Доступ запрещён.<br>
  Ваш IP-адрес не входит в список разрешённых.
</div>
</body>
</html>`

// IPWhitelist allows requests only from IPs stored in the allowed_ips table.
// The list is cached and refreshed every 5 seconds to avoid hitting DB on every request.
// Localhost (127.0.0.1, ::1) is always allowed.
func IPWhitelist(ipRepo *repo.IPRepo) gin.HandlerFunc {
	cache := &ipCache{repo: ipRepo}
	// initial load
	cache.refresh()

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		// strip port if present
		host, _, err := net.SplitHostPort(clientIP)
		if err == nil {
			clientIP = host
		}

		if !cache.isAllowed(clientIP) {
			c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(accessDeniedHTML))
			c.Abort()
			return
		}
		c.Next()
	}
}

// ipCache keeps a cached set of allowed IPs, refreshed periodically.
type ipCache struct {
	repo    *repo.IPRepo
	mu      sync.RWMutex
	allowed map[string]struct{}
	updated time.Time
}

const cacheTTL = 5 * time.Second

func (c *ipCache) refresh() {
	ips, err := c.repo.AllIPs(context.Background())
	if err != nil {
		return // keep old cache on error
	}
	m := make(map[string]struct{}, len(ips)+2)
	for _, ip := range ips {
		m[ip] = struct{}{}
	}
	// always allow localhost
	m["127.0.0.1"] = struct{}{}
	m["::1"] = struct{}{}

	c.mu.Lock()
	c.allowed = m
	c.updated = time.Now()
	c.mu.Unlock()
}

func (c *ipCache) isAllowed(ip string) bool {
	c.mu.RLock()
	stale := time.Since(c.updated) > cacheTTL
	_, ok := c.allowed[ip]
	c.mu.RUnlock()

	if stale {
		c.refresh()
		c.mu.RLock()
		_, ok = c.allowed[ip]
		c.mu.RUnlock()
	}
	return ok
}
