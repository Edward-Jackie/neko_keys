package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"neko_backend/cache"
	"neko_backend/database"
	"neko_backend/models"
)

// AdminHandler handles the authenticated management API.
type AdminHandler struct {
	jwtSecret string
	cache     *cache.KeyCache
}

// NewAdminHandler creates an AdminHandler.
func NewAdminHandler(jwtSecret string, kc *cache.KeyCache) *AdminHandler {
	return &AdminHandler{jwtSecret: jwtSecret, cache: kc}
}

// --------------------------------------------------------------------------
// Auth
// --------------------------------------------------------------------------

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login validates credentials and returns a signed JWT.
func (h *AdminHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	var admin models.AdminUser
	if err := database.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": admin.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenStr,
		"username": admin.Username,
	})
}

// --------------------------------------------------------------------------
// Dashboard
// --------------------------------------------------------------------------

// Dashboard returns aggregate statistics.
func (h *AdminHandler) Dashboard(c *gin.Context) {
	type statusCount struct {
		Status string
		Count  int64
	}

	var results []statusCount
	database.DB.Model(&models.Key{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&results)

	counts := map[string]int64{
		"inactive": 0,
		"active":   0,
		"expired":  0,
		"disabled": 0,
	}
	var total int64
	for _, r := range results {
		counts[r.Status] = r.Count
		total += r.Count
	}

	today := time.Now().Truncate(24 * time.Hour)

	var todayActivations int64
	database.DB.Model(&models.KeyLog{}).
		Where("action = 'activate' AND success = true AND created_at >= ?", today).
		Count(&todayActivations)

	var todayValidations int64
	database.DB.Model(&models.KeyLog{}).
		Where("action = 'validate' AND success = true AND created_at >= ?", today).
		Count(&todayValidations)

	alertCount := countAlerts()

	c.JSON(http.StatusOK, gin.H{
		"total":              total,
		"active":             counts["active"],
		"expired":            counts["expired"],
		"disabled":           counts["disabled"],
		"inactive":           counts["inactive"],
		"today_activations":  todayActivations,
		"today_validations":  todayValidations,
		"alert_count":        alertCount,
	})
}

// countAlerts returns the number of distinct keys that had more than 2
// different IPs in the last hour.
func countAlerts() int64 {
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// Find key_ids where distinct IP count > 2 in the last hour.
	type alertRow struct {
		KeyID uint
		IPs   int64
	}
	var rows []alertRow
	database.DB.Model(&models.KeyLog{}).
		Select("key_id, count(distinct ip) as ips").
		Where("action = 'activate' AND created_at >= ?", oneHourAgo).
		Group("key_id").
		Having("count(distinct ip) > 2").
		Scan(&rows)

	return int64(len(rows))
}

// --------------------------------------------------------------------------
// Batches
// --------------------------------------------------------------------------

type batchResponse struct {
	models.KeyBatch
	TotalKeys  int64 `json:"total_keys"`
	ActiveKeys int64 `json:"active_keys"`
}

// ListBatches returns all batches with key counts.
func (h *AdminHandler) ListBatches(c *gin.Context) {
	var batches []models.KeyBatch
	database.DB.Order("id desc").Find(&batches)

	resp := make([]batchResponse, 0, len(batches))
	for _, b := range batches {
		var total, active int64
		database.DB.Model(&models.Key{}).Where("batch_id = ?", b.ID).Count(&total)
		database.DB.Model(&models.Key{}).Where("batch_id = ? AND status = 'active'", b.ID).Count(&active)
		resp = append(resp, batchResponse{
			KeyBatch:   b,
			TotalKeys:  total,
			ActiveKeys: active,
		})
	}

	c.JSON(http.StatusOK, gin.H{"batches": resp})
}

type createBatchRequest struct {
	Note      string    `json:"note"`
	Type      string    `json:"type" binding:"required,oneof=one_time multi_use"`
	Count     int       `json:"count" binding:"required,min=1,max=10000"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
	MaxUses   int       `json:"max_uses"`
}

// CreateBatch creates a batch and bulk-inserts the generated keys.
func (h *AdminHandler) CreateBatch(c *gin.Context) {
	var req createBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	if req.ExpiresAt.Before(now.Add(24*time.Hour)) || req.ExpiresAt.After(now.Add(180*24*time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be between 1 and 180 days from now"})
		return
	}

	batch := models.KeyBatch{
		Note:      req.Note,
		Count:     req.Count,
		Type:      req.Type,
		ExpiresAt: req.ExpiresAt,
		MaxUses:   req.MaxUses,
	}

	if err := database.DB.Create(&batch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create batch"})
		return
	}

	keys := make([]models.Key, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		keyStr, err := generateKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
			return
		}
		keys = append(keys, models.Key{
			BatchID:   batch.ID,
			Key:       keyStr,
			Type:      req.Type,
			Status:    "inactive",
			ExpiresAt: req.ExpiresAt,
			MaxUses:   req.MaxUses,
		})
	}

	// Bulk insert in chunks to avoid hitting MySQL's packet size limit.
	const chunkSize = 500
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		if end > len(keys) {
			end = len(keys)
		}
		if err := database.DB.Create(keys[i:end]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert keys"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"batch": batch, "generated": len(keys)})
}

// generateKey produces a unique key in the format nk_ + 16 uppercase alphanumeric chars.
func generateKey() (string, error) {
	// 10 random bytes → base32 encodes to 16 chars (without padding).
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	// base32 alphabet is A-Z2-7; uppercase alphanumeric subset. Trim to 16.
	return "nk_" + encoded[:16], nil
}

// --------------------------------------------------------------------------
// Keys
// --------------------------------------------------------------------------

// ListKeys returns a paginated list of keys with optional filters.
func (h *AdminHandler) ListKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	query := database.DB.Model(&models.Key{})

	if batchID := c.Query("batch_id"); batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		like := "%" + strings.ToUpper(keyword) + "%"
		query = query.Where("UPPER(`key`) LIKE ? OR note LIKE ?", like, "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var keys []models.Key
	query.Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&keys)

	c.JSON(http.StatusOK, gin.H{
		"keys":      keys,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetKey returns a single key with its last 50 usage logs.
func (h *AdminHandler) GetKey(c *gin.Context) {
	id := c.Param("id")
	var k models.Key
	if err := database.DB.First(&k, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var logs []models.KeyLog
	database.DB.Where("key_id = ?", k.ID).Order("id desc").Limit(50).Find(&logs)

	c.JSON(http.StatusOK, gin.H{"key": k, "logs": logs})
}

type updateKeyRequest struct {
	Note   string `json:"note"`
	Status string `json:"status"`
}

// UpdateKey allows changing the note or forcing the status to disabled/active.
func (h *AdminHandler) UpdateKey(c *gin.Context) {
	id := c.Param("id")
	var k models.Key
	if err := database.DB.First(&k, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var req updateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Note != "" {
		updates["note"] = req.Note
	}
	if req.Status != "" {
		if req.Status != "disabled" && req.Status != "active" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status can only be set to 'disabled' or 'active'"})
			return
		}
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := database.DB.Model(&k).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update key"})
		return
	}

	// Invalidate cache so the next activate/validate reads fresh data.
	h.cache.Delete(k.Key)

	database.DB.First(&k, k.ID)
	c.JSON(http.StatusOK, gin.H{"key": k})
}

// GetKeyLogs returns all usage logs for a key.
func (h *AdminHandler) GetKeyLogs(c *gin.Context) {
	id := c.Param("id")
	var k models.Key
	if err := database.DB.First(&k, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var logs []models.KeyLog
	database.DB.Where("key_id = ?", k.ID).Order("id desc").Find(&logs)

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// --------------------------------------------------------------------------
// Alerts
// --------------------------------------------------------------------------

type alertKeyInfo struct {
	models.Key
	DistinctIPs int64 `json:"distinct_ips"`
}

// ListAlerts returns keys that had more than 2 distinct IPs in the last hour.
func (h *AdminHandler) ListAlerts(c *gin.Context) {
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	type alertRow struct {
		KeyID uint
		IPs   int64
	}
	var rows []alertRow
	database.DB.Model(&models.KeyLog{}).
		Select("key_id, count(distinct ip) as ips").
		Where("action = 'activate' AND created_at >= ?", oneHourAgo).
		Group("key_id").
		Having("count(distinct ip) > 2").
		Order("ips desc").
		Scan(&rows)

	alerts := make([]alertKeyInfo, 0, len(rows))
	for _, row := range rows {
		var k models.Key
		if err := database.DB.First(&k, row.KeyID).Error; err != nil {
			continue
		}
		alerts = append(alerts, alertKeyInfo{Key: k, DistinctIPs: row.IPs})
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
