// Package middleware provides Gin middleware for the IMSI bot API.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/whynot00/imsi-bot/internal/repo"
)

type contextKey string

const telegramUserIDKey contextKey = "telegram_user_id"

// TelegramAuth validates the Telegram Mini App initData and checks that
// the user exists in the database.
//
// The client must send the raw initData string in the
// X-Telegram-Init-Data header.
func TelegramAuth(botToken string, users *repo.UserRepo) gin.HandlerFunc {
	// Derive the secret key once: HMAC-SHA256("WebAppData", botToken).
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	secretKey := mac.Sum(nil)

	return func(c *gin.Context) {
		initData := c.GetHeader("X-Telegram-Init-Data")
		if initData == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing init data"})
			return
		}

		userID, err := validateInitData(initData, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid init data"})
			return
		}

		ok, err := users.Exists(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth check failed"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.Set(string(telegramUserIDKey), userID)
		c.Next()
	}
}

// UserIDFromContext extracts the Telegram user_id set by TelegramAuth.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(telegramUserIDKey).(int64)
	return v, ok
}

// validateInitData verifies the Telegram initData HMAC and returns the user_id.
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func validateInitData(initData string, secretKey []byte) (int64, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, fmt.Errorf("parse init data: %w", err)
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return 0, fmt.Errorf("missing hash")
	}

	// Build the data-check string: sorted key=value pairs excluding "hash".
	pairs := make([]string, 0, len(values))
	for k, vs := range values {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+vs[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Compute expected HMAC-SHA256.
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(receivedHash)) {
		return 0, fmt.Errorf("hash mismatch")
	}

	// Extract user_id from the "user" field (JSON object).
	userJSON := values.Get("user")
	userID, err := extractUserID(userJSON)
	if err != nil {
		return 0, fmt.Errorf("extract user id: %w", err)
	}

	return userID, nil
}

// extractUserID pulls the numeric id from a JSON string like {"id":123456,...}
// without importing encoding/json to keep the dependency minimal.
func extractUserID(userJSON string) (int64, error) {
	// Find "id": <number>
	const key = `"id":`
	idx := strings.Index(userJSON, key)
	if idx == -1 {
		return 0, fmt.Errorf("id field not found in user json")
	}
	rest := strings.TrimSpace(userJSON[idx+len(key):])
	end := strings.IndexAny(rest, ",}")
	if end == -1 {
		end = len(rest)
	}
	idStr := strings.TrimSpace(rest[:end])
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse user id %q: %w", idStr, err)
	}
	return id, nil
}
