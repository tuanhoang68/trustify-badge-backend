package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	r.GET("/api/badge-config", func(c *gin.Context) {
		shop := c.Query("shop")
		if shop == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop"})
			return
		}

		var cfg storage.BadgeConfig
		if err := db.Where("shop_domain = ?", shop).First(&cfg).Error; err != nil {
			if err.Error() == "record not found" {
				c.JSON(http.StatusOK, gin.H{"data": nil})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": cfg})
	})

	r.PUT("/api/badge-config", func(c *gin.Context) {
		var body storage.BadgeConfig
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		if body.ShopDomain == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop_domain"})
			return
		}
		// Upsert by shop_domain
		var existing storage.BadgeConfig
		err := db.Where("shop_domain = ?", body.ShopDomain).First(&existing).Error
		if err == nil {
			body.ID = existing.ID
		}
		body.UpdatedAt = time.Now()
		if err := db.Save(&body).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": body})
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
