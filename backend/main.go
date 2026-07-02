package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"hot-contracts-backend/model"
	"hot-contracts-backend/service"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

//go:embed all:dist
var frontendFS embed.FS

var dataCache = cache.New(60*time.Second, 2*time.Minute)
var binanceSvc = service.NewBinanceService()

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API
	r.GET("/api/v1/contracts/hot", getHotContractsHandler)

	// 嵌入的前端静态文件
	distFS, _ := fs.Sub(frontendFS, "dist")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path[1:] // 去掉前导 /
		if p == "" {
			p = "index.html"
		}
		// 尝试读取静态文件
		f, err := distFS.(fs.ReadFileFS).ReadFile(p)
		if err != nil {
			// SPA fallback
			f, _ = distFS.(fs.ReadFileFS).ReadFile("index.html")
		}
		ct := mimeByExt(path.Ext(p))
		c.Data(http.StatusOK, ct, f)
	})

	// 端口：优先使用环境变量 PORT（Render 等平台使用），否则默认 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Printf("Open http://localhost:%s in browser", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server run failed: %v", err)
	}
}

func mimeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".ico":
		return "image/x-icon"
	case ".woff", ".woff2":
		return "font/woff"
	default:
		return "application/octet-stream"
	}
}

func getHotContractsHandler(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "volume24hUsd")
	order := c.DefaultQuery("order", "descending")

	page := 1
	pageSize := 25
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	cacheKey := "hot_contracts_data"
	var data []model.HotContract

	if cachedData, found := dataCache.Get(cacheKey); found {
		data = cachedData.([]model.HotContract)
	} else {
		var err error
		data, err = binanceSvc.BuildHotContracts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dataCache.Set(cacheKey, data, cache.DefaultExpiration)
	}

	sortedData := make([]model.HotContract, len(data))
	copy(sortedData, data)
	service.SortContracts(sortedData, sortBy, order)

	total := len(sortedData)
	start := (page - 1) * pageSize
	if start >= total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      sortedData[start:end],
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
