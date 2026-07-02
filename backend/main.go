package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"hot-contracts-backend/model"
	"hot-contracts-backend/service"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var dataCache = cache.New(60*time.Second, 2*time.Minute)
var binanceSvc = service.NewBinanceService()

func main() {
	// 设置 Gin 为 Release 模式，生产环境推荐
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 简单的 CORS 中间件，允许前端跨域访问
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

	// API 路由组
	api := r.Group("/api/v1")
	{
		api.GET("/contracts/hot", getHotContractsHandler)
	}

	log.Println("Server starting on :8082...")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Server run failed: %v", err)
	}
}

func getHotContractsHandler(c *gin.Context) {
	// 获取排序参数
	sortBy := c.DefaultQuery("sort_by", "volume24hUsd")
	order := c.DefaultQuery("order", "descending")

	// 获取分页参数
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

	// 浅拷贝一份数据用于排序，避免污染缓存中的原始顺序
	sortedData := make([]model.HotContract, len(data))
	copy(sortedData, data)

	// 在后端执行排序
	service.SortContracts(sortedData, sortBy, order)

	// 分页
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
