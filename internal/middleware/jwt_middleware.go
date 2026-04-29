package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"go-attendance-api/internal/config"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
)

//////////////////////////////////////////////////////////////
// CONFIG
//////////////////////////////////////////////////////////////

const MAX_REQUEST_AGE = 10000 // 10 detik (dalam milidetik)

var rdb = config.NewRedis()

//////////////////////////////////////////////////////////////
// MAIN MIDDLEWARE
//////////////////////////////////////////////////////////////

func SecureAuth(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		isDev := os.Getenv("APP_ENV") == "development"

		// Tarik header yang dibutuhkan
		tsStr := c.GetHeader("X-Timestamp")
		reqID := c.GetHeader("X-Request-ID")
		sig := c.GetHeader("X-Signature")

		////////////////////////////////////////////////////////
		// 1. TIMESTAMP CHECK (BYPASS IN DEV)
		////////////////////////////////////////////////////////
		if !isDev {
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				c.AbortWithStatusJSON(403, gin.H{"message": "Invalid timestamp format (Required in Production)"})
				return
			}

			nowMilli := time.Now().UnixMilli()
			deltaMilli := abs(nowMilli - ts)
			if deltaMilli > MAX_REQUEST_AGE {
				c.AbortWithStatusJSON(403, gin.H{"message": "Request expired"})
				return
			}
		}

		////////////////////////////////////////////////////////
		// 2. INTERNAL SECRET (BYPASS IN DEV)
		////////////////////////////////////////////////////////
		if !isDev && c.GetHeader("X-Internal-Secret") != os.Getenv("INTERNAL_SECRET") {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: Invalid Internal Secret"})
			return
		}

		////////////////////////////////////////////////////////
		// 3. COOKIE AUTH (ALWAYS REQUIRED)
		////////////////////////////////////////////////////////
		token, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized: Missing Session"})
			return
		}

		user, err := authService.GetMe(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"message": "Invalid session"})
			return
		}

		// Check if tenant is suspended
		if user.Tenant != nil && user.Tenant.IsSuspended {
			reason := "Tenant account is suspended"
			if user.Tenant.SuspendedReason != "" {
				reason = fmt.Sprintf("Tenant account is suspended: %s", user.Tenant.SuspendedReason)
			}
			c.AbortWithStatusJSON(403, gin.H{"message": reason})
			return
		}

		////////////////////////////////////////////////////////
		// 3.5 BLACKLIST CHECK (REDIS)
		////////////////////////////////////////////////////////
		blacklistKey := fmt.Sprintf("blacklist:%s", token)
		isBlacklisted, _ := rdb.Exists(config.Ctx, blacklistKey).Result()
		if isBlacklisted > 0 {
			c.AbortWithStatusJSON(401, gin.H{"message": "Session expired or logged out"})
			return
		}

		////////////////////////////////////////////////////////
		// 4. ANTI REPLAY (BYPASS IN DEV)
		////////////////////////////////////////////////////////
		if !isDev {
			if reqID == "" {
				c.AbortWithStatusJSON(403, gin.H{"message": "Missing request id (Required in Production)"})
				return
			}

			redisKey := fmt.Sprintf("nonce:%s", reqID)
			success, err := rdb.SetNX(config.Ctx, redisKey, "1", 15*time.Second).Result()
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "Security validation failed (Redis Error)"})
				return
			}
			if !success {
				c.AbortWithStatusJSON(403, gin.H{"message": "Replay detected: Request ID already used"})
				return
			}
		}

		////////////////////////////////////////////////////////
		// 5. SIGNATURE VERIFICATION (BYPASS IN DEV)
		////////////////////////////////////////////////////////
		if !isDev && sig != "" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "multipart") {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				if len(bodyBytes) == 0 {
					bodyBytes = []byte("{}")
				}

				expected := generateSignature(bodyBytes, tsStr, reqID)
				if !hmac.Equal([]byte(sig), []byte(expected)) {
					c.AbortWithStatusJSON(403, gin.H{"message": "Invalid signature"})
					return
				}
			}
		}

		////////////////////////////////////////////////////////
		// 6. SET CONTEXT
		////////////////////////////////////////////////////////

		// Inject into standard context for GORM plugin
		tenantID := user.TenantID
		if user.Role != nil && user.Role.BaseRole == model.BaseRoleSuperAdmin {
			tenantID = 0 // Bypass for Superadmin
		}

		ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("user_id", user.ID)
		c.Set("tenant_id", user.TenantID) // Keep original tenant_id in Gin context if needed
		c.Set("permissions", user.Permissions)
		c.Set("plan_features", user.PlanFeatures)

		// Set role name string for RequireRole middleware
		if user.Role != nil {
			c.Set("role", user.Role.Name)
			c.Set("base_role", user.Role.BaseRole)
		} else {
			c.Set("role", "")
			c.Set("base_role", "")
		}

		c.Next()
	}
}

func HasPermission(permissionID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dual-nature Superadmin logic: If it's ADMIN base role (Tenant 1), bypass permission check
		baseRole, _ := c.Get("base_role")
		tenantID, _ := c.Get("tenant_id")

		if baseRole == string(model.BaseRoleSuperAdmin) || (tenantID == uint(1) && baseRole == string(model.BaseRoleAdmin)) {
			c.Next()
			return
		}

		// 🆕 PLAN ENFORCEMENT: Check if module is allowed in Plan
		planFeaturesVal, planExists := c.Get("plan_features")
		if planExists {
			planFeatures := planFeaturesVal.([]string)
			isModuleAllowed := false

			// Check for wildcard
			for _, f := range planFeatures {
				if f == "*" {
					isModuleAllowed = true
					break
				}
			}

			if !isModuleAllowed {
				// Get module from permissionID (e.g., "attendance.view" -> "attendance")
				parts := strings.Split(permissionID, ".")
				if len(parts) > 0 {
					module := parts[0]
					for _, f := range planFeatures {
						if f == module {
							isModuleAllowed = true
							break
						}
					}
				}
			}

			if !isModuleAllowed {
				c.AbortWithStatusJSON(403, gin.H{"message": "Feature not available in your current plan. Please upgrade."})
				return
			}
		}

		permissionsVal, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: No permissions assigned"})
			return
		}

		permissions := permissionsVal.([]string)
		hasPermission := false
		for _, p := range permissions {
			if p == permissionID {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(403, gin.H{"message": fmt.Sprintf("Forbidden: Missing permission %s", permissionID)})
			return
		}

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		userRole := fmt.Sprintf("%v", userRoleVal)
		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: Insufficient permissions"})
			return
		}

		c.Next()
	}
}

func RequireBaseRole(baseRoles ...model.BaseRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseRoleVal, exists := c.Get("base_role")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		userBaseRole := model.BaseRole(fmt.Sprintf("%v", baseRoleVal))
		allowed := false
		for _, role := range baseRoles {
			if userBaseRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: Insufficient base role permissions"})
			return
		}

		c.Next()
	}
}

func RequireTenant(tenantID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		tIDVal, exists := c.Get("tenant_id")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"message": "Unauthorized"})
			return
		}

		tID := tIDVal.(uint)
		if tID != tenantID {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: Access restricted to specific tenant"})
			return
		}

		c.Next()
	}
}

//////////////////////////////////////////////////////////////
// HELPERS
//////////////////////////////////////////////////////////////

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func generateSignature(body []byte, timestamp string, requestID string) string {
	secret := os.Getenv("SIGN_SECRET")

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	h.Write([]byte(timestamp))
	h.Write([]byte(requestID))

	return hex.EncodeToString(h.Sum(nil))
}
