package middleware

import (
	"bytes"
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
		// 1. TIMESTAMP CHECK & DEBUG LOGGING
		////////////////////////////////////////////////////////
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"message": "Invalid timestamp format"})
			return
		}

		nowMilli := time.Now().UnixMilli()
		deltaMilli := abs(nowMilli - ts)
		deltaSec := float64(deltaMilli) / 1000.0

		// LOG DEBUGGING YANG LEBIH READABLE
		fmt.Println("\n================= 🛡️ SECURITY DEBUG 🛡️ =================")
		fmt.Printf(" 📍 Path      : [%s] %s\n", c.Request.Method, c.Request.URL.Path)
		fmt.Printf(" 🆔 Request ID: %s\n", reqID)
		fmt.Printf(" 🕒 Header TS : %d\n", ts)
		fmt.Printf(" 🕒 Server TS : %d\n", nowMilli)
		fmt.Printf(" ⏱️ Selisih   : %.2f detik (%d ms)\n", deltaSec, deltaMilli)
		fmt.Printf(" 🔐 Signature : %s\n", sig)
		fmt.Println("=========================================================")

		// Validasi Expired
		if deltaMilli > MAX_REQUEST_AGE {
			fmt.Printf(" ❌ BLOCKED: Request terlalu tua (%.2f detik > 10 detik)\n", deltaSec)
			c.AbortWithStatusJSON(403, gin.H{"message": "Request expired"})
			return
		}

		////////////////////////////////////////////////////////
		// 2. INTERNAL SECRET
		////////////////////////////////////////////////////////
		if !isDev && c.GetHeader("X-Internal-Secret") != os.Getenv("INTERNAL_SECRET") {
			c.AbortWithStatusJSON(403, gin.H{"message": "Forbidden: Invalid Internal Secret"})
			return
		}

		////////////////////////////////////////////////////////
		// 3. COOKIE AUTH
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

		////////////////////////////////////////////////////////
		// 4. ANTI REPLAY (REDIS SETNX)
		////////////////////////////////////////////////////////
		if reqID == "" {
			c.AbortWithStatusJSON(403, gin.H{"message": "Missing request id"})
			return
		}

		redisKey := fmt.Sprintf("nonce:%s", reqID)
		success, err := rdb.SetNX(config.Ctx, redisKey, "1", 15*time.Second).Result()

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"message": "Security validation failed (Redis Error)"})
			return
		}

		if !success {
			fmt.Println(" ❌ BLOCKED: Replay Attack Detected (Request ID sudah dipakai)")
			c.AbortWithStatusJSON(403, gin.H{"message": "Replay detected: Request ID already used"})
			return
		}

		////////////////////////////////////////////////////////
		// 5. SIGNATURE VERIFICATION
		////////////////////////////////////////////////////////
		contentType := c.GetHeader("Content-Type")

		if sig != "" && !strings.Contains(contentType, "multipart") {

			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) == 0 {
				bodyBytes = []byte("{}")
			}

			expected := generateSignature(bodyBytes, tsStr, reqID)

			if !hmac.Equal([]byte(sig), []byte(expected)) {
				fmt.Printf(" ❌ BLOCKED: Invalid Signature\n    Expected: %s\n    Received: %s\n", expected, sig)
				if !isDev {
					c.AbortWithStatusJSON(403, gin.H{"message": "Invalid signature"})
					return
				} else {
					// Di dev mode kita biarkan lewat, tapi log errornya
					fmt.Println(" ⚠️ WARNING: Invalid signature diabaikan karena mode Development")
				}
			} else {
				fmt.Println(" ✅ SUCCESS: Signature Valid!")
			}
		}

		////////////////////////////////////////////////////////
		// 6. SET CONTEXT
		////////////////////////////////////////////////////////
		c.Set("user_id", user.ID)
		c.Set("tenant_id", user.TenantID)
		c.Set("role", user.Role)

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
