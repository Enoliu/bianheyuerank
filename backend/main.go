package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
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
	port := flag.String("port", "8082", "server port")
	flag.Parse()

	// 环境变量优先于命令行参数
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	r.GET("/api/v1/contracts/hot", getHotContractsHandler)

	distFS, _ := fs.Sub(frontendFS, "dist")
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path[1:]
		if p == "" {
			p = "index.html"
		}
		f, err := distFS.(fs.ReadFileFS).ReadFile(p)
		if err != nil {
			f, _ = distFS.(fs.ReadFileFS).ReadFile("index.html")
		}
		ct := mimeByExt(p)
		c.Data(http.StatusOK, ct, f)
	})

	log.Printf("Server starting on :%s", *port)
	log.Printf("Open http://localhost:%s in browser", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("Server run failed: %v", err)
	}
}

func mimeByExt(p string) string {
	ext := ""
	if i := len(p) - 1; i >= 0 {
		for j := i; j >= 0; j-- {
			if p[j] == '.' {
				ext = p[j:]
				break
			}
			if p[j] == '/' {
				break
			}
		}
	}
	switch ext {
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
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
		return "text/html; charset=utf-8"
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
