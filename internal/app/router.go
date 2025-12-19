package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tuanhoang68/trustify-badge-backend/internal/shopify"
	"github.com/tuanhoang68/trustify-badge-backend/internal/storage"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(CORS())

	// Health
	r.GET("/health", Health())

	// Shopify OAuth start
	r.GET("/api/shopify/auth", func(c *gin.Context) {
		shop := c.Query("shop")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop"})
			return
		}
		state := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Redirect(http.StatusFound, shopify.AuthURL(shop, state))
	})

	// Shopify OAuth callback
	r.GET("/api/shopify/callback", func(c *gin.Context) {
		q := c.Request.URL.Query()
		if !shopify.VerifyHMAC(q) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid hmac"})
			return
		}
		shop := q.Get("shop")
		code := q.Get("code")

		// Exchange code for access token
		token, err := shopify.ExchangeToken(shop, code)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "token exchange failed"})
			return
		}

		// Save shop + token
		db.Where(storage.Shop{ShopDomain: shop}).Assign(storage.Shop{
			ShopDomain:  shop,
			AccessToken: token,
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}).FirstOrCreate(&storage.Shop{})

		// Redirect to app home (embedded)
		appURL := os.Getenv("SHOPIFY_APP_URL")
		c.Redirect(http.StatusFound, appURL+"?shop="+shop)
	})

	// Badge config APIs
	// GET: lấy config theo shop_domain (query param)
	r.GET("/api/badge-config", func(c *gin.Context) {
		shopDomain := c.Query("shop")
		if shopDomain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop"})
			return
		}
		var shop storage.Shop
		if err := db.Where("shop_domain = ?", shopDomain).First(&shop).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"data": nil})
			return
		}
		var cfg storage.BadgeConfig
		if err := db.Where("shop_id = ?", shop.ID).First(&cfg).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": cfg})
	})

	r.PUT("/api/badge-config", func(c *gin.Context) {
		var body struct {
			ShopDomain    string `json:"shop_domain"`
			Layout        string `json:"layout"`
			ReviewText    string `json:"reviewText"`
			PoweredByText string `json:"poweredByText"`
			HeaderText    string `json:"headerText"`
			CustomCode    string `json:"customCode"`
			CustomCSS     string `json:"customCSS"`
			Alignment     string `json:"alignment"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		fmt.Printf("FE gửi lên: %+v\n", body)
		if body.ShopDomain == "" || body.Layout == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop_domain or layout"})
			return
		}

		// Tìm hoặc tạo Shop
		var shop storage.Shop
		if err := db.Where("shop_domain = ?", body.ShopDomain).First(&shop).Error; err != nil {
			shop = storage.Shop{
				ID:          uuid.New().String(),
				ShopDomain:  body.ShopDomain,
				AccessToken: "", // sẽ cập nhật sau khi OAuth
				InstalledAt: time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := db.Create(&shop).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create shop failed"})
				return
			}
		}

		// Upsert BadgeConfig theo ShopID
		var existing storage.BadgeConfig
		err := db.Where("shop_id = ?", shop.ID).First(&existing).Error
		if err == nil {
			existing.Layout = body.Layout
			existing.ReviewText = body.ReviewText
			existing.PoweredByText = body.PoweredByText
			existing.HeaderText = body.HeaderText
			existing.CustomCode = body.CustomCode
			existing.CustomCSS = body.CustomCSS
			existing.Alignment = body.Alignment
			existing.UpdatedAt = time.Now()
			if err := db.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": existing})
			return
		}
		newCfg := storage.BadgeConfig{
			ID:            uuid.New().String(),
			ShopID:        shop.ID,
			Layout:        body.Layout,
			ReviewText:    body.ReviewText,
			PoweredByText: body.PoweredByText,
			HeaderText:    body.HeaderText,
			CustomCode:    body.CustomCode,
			CustomCSS:     body.CustomCSS,
			Alignment:     body.Alignment,
			UpdatedAt:     time.Now(),
		}
		if err := db.Create(&newCfg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": newCfg})
	})

	// App Proxy endpoint (optional)
	// In Shopify admin, set App Proxy (e.g., /apps/trustify-badge/proxy) to hit /api/proxy/badge-config
	r.GET("/api/proxy/badge-config", func(c *gin.Context) {
		shop := c.Query("shop")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop"})
			return
		}
		var cfg storage.BadgeConfig
		if err := db.Where("shop_domain = ?", shop).First(&cfg).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": cfg})
	})

	return r
}
