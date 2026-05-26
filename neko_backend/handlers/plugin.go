package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"neko_backend/cache"
	"neko_backend/database"
	"neko_backend/models"
)

// PluginHandler handles the public plugin-facing API endpoints.
type PluginHandler struct {
	cache *cache.KeyCache
}

// NewPluginHandler creates a PluginHandler backed by the given cache.
func NewPluginHandler(kc *cache.KeyCache) *PluginHandler {
	return &PluginHandler{cache: kc}
}

// pluginReq is the JSON body shared by the activate and validate endpoints.
type pluginReq struct {
	ShopName string `json:"shop_name"`
	ShopID   string `json:"shop_id"`
}

// extractBearerKey parses the Authorization header and returns the key string.
// It expects "Bearer nk_XXXX".
func extractBearerKey(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// getClientIP returns the best-effort real client IP.
func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may be a comma-separated list; take the first entry.
		return strings.TrimSpace(strings.SplitN(ip, ",", 2)[0])
	}
	if ip := c.GetHeader("X-Real-Ip"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// lookupKey fetches the key from cache or database.
func (h *PluginHandler) lookupKey(keyStr string) (*models.Key, error) {
	if k, ok := h.cache.Get(keyStr); ok {
		return k, nil
	}
	var k models.Key
	if err := database.DB.Where("key = ?", keyStr).First(&k).Error; err != nil {
		return nil, err
	}
	h.cache.Set(keyStr, &k)
	return &k, nil
}

// writeKeyLog inserts a key_log record.
func writeKeyLog(keyID uint, ip, action string, success bool) {
	log := models.KeyLog{
		KeyID:   keyID,
		IP:      ip,
		Action:  action,
		Success: success,
	}
	database.DB.Create(&log)
}

// checkSharedAlert queries key_logs for distinct IPs in the last hour and
// returns true when more than 2 different IPs used the key.
func checkSharedAlert(keyID uint) bool {
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	database.DB.Model(&models.KeyLog{}).
		Where("key_id = ? AND action = 'activate' AND created_at >= ?", keyID, oneHourAgo).
		Distinct("ip").
		Count(&count)
	return count > 2
}

// Activate godoc
// POST /api/v1/activate
// Authorization: Bearer nk_XXXXXXXXXXXXXXXX
func (h *PluginHandler) Activate(c *gin.Context) {
	keyStr, ok := extractBearerKey(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "缺少授权头"})
		return
	}

	ip := getClientIP(c)

	var req pluginReq
	_ = c.ShouldBindJSON(&req)
	shopName := strings.TrimSpace(req.ShopName)
	if shopName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "缺少店铺名称"})
		return
	}

	k, err := h.lookupKey(keyStr)
	if err != nil {
		writeKeyLog(0, ip, "activate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥不存在"})
		return
	}

	// Validate status before mutating.
	if k.Status == "disabled" {
		writeKeyLog(k.ID, ip, "activate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已禁用"})
		return
	}
	if k.IsExpired() {
		// Mark as expired in DB if not already.
		if k.Status != "expired" {
			database.DB.Model(k).Update("status", "expired")
			h.cache.Delete(keyStr)
		}
		writeKeyLog(k.ID, ip, "activate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已过期"})
		return
	}
	if k.IsUsageLimitReached() {
		writeKeyLog(k.ID, ip, "activate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已达使用上限"})
		return
	}

	// One key = one shop: reject activation from a shop other than the bound one.
	if k.BoundShop != "" && k.BoundShop != shopName {
		writeKeyLog(k.ID, ip, "activate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已绑定其他店铺"})
		return
	}

	// Build the update map.
	updates := map[string]interface{}{
		"use_count": k.UseCount + 1,
		"status":    "active",
	}
	if k.ActivatedAt == nil {
		now := time.Now()
		updates["activated_at"] = now
		updates["activated_ip"] = ip
	}
	// Bind the shop on first activation.
	if k.BoundShop == "" {
		updates["bound_shop"] = shopName
	}

	// Check if this increment will exhaust the key, and mark accordingly.
	newUseCount := k.UseCount + 1
	willExpire := false
	switch k.Type {
	case "one_time":
		willExpire = newUseCount >= 1
	case "multi_use":
		willExpire = k.MaxUses > 0 && newUseCount >= k.MaxUses
	}
	if willExpire {
		updates["status"] = "expired"
	}

	if err := database.DB.Model(k).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "内部错误"})
		return
	}
	// Invalidate cache so subsequent reads get fresh data.
	h.cache.Delete(keyStr)

	writeKeyLog(k.ID, ip, "activate", true)

	// Async shared-IP alert check (non-blocking, best effort).
	go func(kid uint) {
		_ = checkSharedAlert(kid)
	}(k.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"key_id":     k.ID,
		"expires_at": k.ExpiresAt.UTC().Format(time.RFC3339),
		"bound_shop": shopName,
	})
}

// Validate godoc
// POST /api/v1/validate
// Authorization: Bearer nk_XXXXXXXXXXXXXXXX
// Does NOT mutate any state — only checks validity.
func (h *PluginHandler) Validate(c *gin.Context) {
	keyStr, ok := extractBearerKey(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "缺少授权头"})
		return
	}

	ip := getClientIP(c)

	var req pluginReq
	_ = c.ShouldBindJSON(&req)
	shopName := strings.TrimSpace(req.ShopName)
	if shopName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "缺少店铺名称"})
		return
	}

	k, err := h.lookupKey(keyStr)
	if err != nil {
		writeKeyLog(0, ip, "validate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥不存在"})
		return
	}

	if k.Status == "disabled" {
		writeKeyLog(k.ID, ip, "validate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已禁用"})
		return
	}
	if k.IsExpired() {
		writeKeyLog(k.ID, ip, "validate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "密钥已过期"})
		return
	}
	// Ongoing validation intentionally ignores the usage limit: an activated
	// one_time key has use_count>=1 but must keep validating for its bound shop
	// until the calendar expiry. It enforces the one-key-one-shop binding instead.
	if k.BoundShop == "" || k.BoundShop != shopName {
		writeKeyLog(k.ID, ip, "validate", false)
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "shop_mismatch"})
		return
	}

	writeKeyLog(k.ID, ip, "validate", true)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"key_id":     k.ID,
		"expires_at": k.ExpiresAt.UTC().Format(time.RFC3339),
		"bound_shop": k.BoundShop,
	})
}
