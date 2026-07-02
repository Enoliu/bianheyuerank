<template>
  <div class="p-4 h-screen flex flex-col bg-gray-900 text-gray-200 font-sans">
    <div class="flex justify-between items-center mb-3">
      <h1 class="text-xl font-bold tracking-tight text-white">🔥 热门合约看板</h1>
      <div class="flex items-center space-x-3">
        <div class="text-xs text-gray-400">
          更新: {{ lastUpdated }}
        </div>
        <el-tag :type="isPolling ? 'success' : 'danger'" effect="dark" size="small">
          {{ isPolling ? '实时' : '暂停' }}
        </el-tag>
      </div>
    </div>

    <div class="flex-grow overflow-hidden border border-gray-700/50 rounded-lg">
      <el-table
        :data="rawData"
        style="width: 100%; height: 100%"
        size="small"
        :row-class-name="tableRowClassName"
        @sort-change="handleSortChange"
        :default-sort="{ prop: 'volume24hUsd', order: 'descending' }"
        :header-cell-style="{ background: '#1a1f2e', color: '#9ca3af', borderBottom: '1px solid #374151', padding: '6px 0', fontSize: '12px' }"
        :cell-style="{ padding: '5px 0', borderBottom: '1px solid #1f2937' }"
      >
        <!-- 交易对 -->
        <el-table-column prop="symbol" label="交易对" width="110" fixed sortable="custom">
          <template #default="{ row }">
            <a :href="`https://www.binance.com/zh-CN/futures/${row.symbol}`" target="_blank" class="font-semibold text-blue-400 hover:text-blue-300 text-xs">
              {{ row.symbol }}
            </a>
          </template>
        </el-table-column>

        <!-- 最新价 -->
        <el-table-column prop="price" min-width="100" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              最新价
              <el-tooltip content="合约当前最新成交价格(USDT)" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs">{{ formatPrice(row.price) }}</span>
          </template>
        </el-table-column>

        <!-- 24h涨跌 -->
        <el-table-column prop="priceChangePercent" min-width="90" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              涨跌
              <el-tooltip content="过去24小时价格涨跌幅" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span :class="getColorClass(row.priceChangePercent)" class="font-mono text-xs font-semibold">
              {{ row.priceChangePercent > 0 ? '+' : '' }}{{ row.priceChangePercent.toFixed(2) }}%
            </span>
          </template>
        </el-table-column>

        <!-- 相对BTC强弱 -->
        <el-table-column prop="relativeStrength" min-width="90" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              强弱
              <el-tooltip content="该币种涨跌幅 - BTC涨跌幅" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span :class="getColorClass(row.relativeStrength)" class="font-mono text-xs">
              {{ row.relativeStrength > 0 ? '+' : '' }}{{ row.relativeStrength.toFixed(2) }}%
            </span>
          </template>
        </el-table-column>

        <!-- 24h振幅 -->
        <el-table-column prop="amplitude24h" min-width="80" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              振幅
              <el-tooltip content="(最高价-最低价)/最低价" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs text-gray-400">{{ (row.amplitude24h * 100).toFixed(1) }}%</span>
          </template>
        </el-table-column>

        <!-- 期现溢价率 -->
        <el-table-column prop="premiumRate" min-width="85" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              溢价率
              <el-tooltip content="(合约-现货)/现货" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span :class="getColorClass(row.premiumRate)" class="font-mono text-xs">
              {{ row.premiumRate !== 0 ? (row.premiumRate * 100).toFixed(2) + '%' : '-' }}
            </span>
          </template>
        </el-table-column>

        <!-- 主动买卖比 -->
        <el-table-column prop="takerBuyRatio" min-width="110" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              买卖比
              <el-tooltip content="Taker买入额占比，>50%买入强" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <div class="relative w-full h-4 bg-gray-800 rounded overflow-hidden flex items-center justify-center text-[10px] font-mono border border-gray-700/50">
              <div class="absolute left-0 top-0 bottom-0 bg-emerald-500/25" :style="{ width: `${row.takerBuyRatio * 100}%` }"></div>
              <div class="absolute right-0 top-0 bottom-0 bg-rose-500/25" :style="{ width: `${(1 - row.takerBuyRatio) * 100}%` }"></div>
              <span class="relative z-10 font-bold" :class="row.takerBuyRatio > 0.5 ? 'text-emerald-400' : 'text-rose-400'">
                {{ (row.takerBuyRatio * 100).toFixed(1) }}%
              </span>
            </div>
          </template>
        </el-table-column>

        <!-- 多空比 -->
        <el-table-column prop="longShortRatio" min-width="80" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              多空比
              <el-tooltip content="多头/空头账户数" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs" :class="row.longShortRatio > 1 ? 'text-emerald-400' : row.longShortRatio < 1 ? 'text-rose-400' : 'text-gray-400'">
              {{ row.longShortRatio ? row.longShortRatio.toFixed(2) : '-' }}
            </span>
          </template>
        </el-table-column>

        <!-- 净流入额 -->
        <el-table-column prop="netTakerVolumeUsd" min-width="100" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              净流入
              <el-tooltip content="Taker买入额-卖出额" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span :class="getColorClass(row.netTakerVolumeUsd)" class="font-mono text-xs">
              {{ row.netTakerVolumeUsd > 0 ? '+' : '' }}{{ formatBigNumber(row.netTakerVolumeUsd) }}
            </span>
          </template>
        </el-table-column>

        <!-- 24h成交额 -->
        <el-table-column prop="volume24hUsd" min-width="90" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              成交额
              <el-tooltip content="24h合约成交额(USDT)" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs text-gray-400">{{ formatBigNumber(row.volume24hUsd) }}</span>
          </template>
        </el-table-column>

        <!-- 资金费率 -->
        <el-table-column prop="fundingRate" min-width="120" align="center" sortable="custom">
          <template #header>
            <div class="inline-flex items-center justify-center">
              资金费率
              <el-tooltip content="正=多头付空头，负=空头付多头" placement="top" effect="dark">
                <el-icon class="ml-0.5 cursor-help text-gray-500 hover:text-gray-300"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs" :class="getColorClass(row.fundingRate)">
              {{ row.fundingRate !== 0 ? (row.fundingRate * 100).toFixed(4) + '%' : '-' }}
            </span>
            <span class="text-[10px] text-gray-500 ml-0.5">/ {{ row.fundingIntervalHours || 8 }}h</span>
          </template>
        </el-table-column>

        <!-- 结算倒计时 -->
        <el-table-column prop="nextFundingTime" min-width="90" align="center">
          <template #header>
            <span>倒计时</span>
          </template>
          <template #default="{ row }">
            <span class="font-mono text-xs" :class="{ 'animate-pulse text-amber-400': isFundingSoon(row.nextFundingTime) }">
              {{ formatCountdown(row.nextFundingTime) }}
            </span>
          </template>
        </el-table-column>

        <template #append>
          <div class="scroll-sentinel py-3 text-center text-xs text-gray-500">
            <template v-if="isLoading">
              <span class="inline-flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-3 w-3 text-blue-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                加载中...
              </span>
            </template>
            <template v-else-if="!hasMore && rawData.length > 0">
              <span class="text-gray-600">已加载全部 {{ total }} 条数据</span>
            </template>
          </div>
        </template>
      </el-table>
    </div>

    <div class="mt-2 text-xs text-gray-500 text-right">
      已加载 {{ rawData.length }} / {{ total }} 条
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useIntervalFn, useNow } from '@vueuse/core'
import { InfoFilled } from '@element-plus/icons-vue'

interface HotContract {
  symbol: string
  price: number
  spotPrice: number
  markPriceDeviation: number
  premiumRate: number
  priceChangePercent: number
  relativeStrength: number
  amplitude24h: number
  volume24hUsd: number
  takerBuyRatio: number
  netTakerVolumeUsd: number
  openInterestUsd: number
  oiChangePercent24h: number
  volToOiRatio: number
  fundingRate: number
  fundingIntervalHours: number
  annualizedFundingRate: number
  nextFundingTime: number
  longShortRatio: number
}

interface ApiResponse {
  data: HotContract[]
  total: number
  page: number
  page_size: number
}

const API_URL = '/api/v1/contracts/hot'

const rawData = ref<HotContract[]>([])
const lastUpdated = ref('')
const sortState = ref({ prop: 'volume24hUsd', order: 'descending' })
const now = useNow({ interval: 5000 })

// 分页状态
const currentPage = ref(1)
const pageSize = 25
const total = ref(0)
const isLoading = ref(false)
const hasMore = ref(true)

// 无限滚动
let observer: IntersectionObserver | null = null

// Fetch Logic
const fetchData = async (page: number = 1, append: boolean = false) => {
  if (isLoading.value) return
  isLoading.value = true

  try {
    const params = new URLSearchParams({
      sort_by: sortState.value.prop || 'volume24hUsd',
      order: sortState.value.order || 'descending',
      page: page.toString(),
      page_size: pageSize.toString(),
    })

    const response = await fetch(`${API_URL}?${params}`)
    if (!response.ok) {
      console.error("Fetch error:", response.statusText)
      return
    }

    const result: ApiResponse = await response.json()
    if (result.data && Array.isArray(result.data)) {
      if (append) {
        rawData.value = [...rawData.value, ...result.data]
      } else {
        rawData.value = result.data
      }
      total.value = result.total
      currentPage.value = result.page
      hasMore.value = rawData.value.length < total.value
      lastUpdated.value = new Date().toLocaleTimeString()
    }
  } catch (e) {
    console.error("Fetch error:", e)
  } finally {
    isLoading.value = false
  }
}

// 加载更多
const loadMore = async () => {
  if (isLoading.value || !hasMore.value) return
  await fetchData(currentPage.value + 1, true)
}

// 初始化无限滚动
const initInfiniteScroll = () => {
  // 获取表格滚动容器
  const tableEl = document.querySelector('.el-table__body-wrapper')
  if (!tableEl) return

  observer = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting && hasMore.value && !isLoading.value) {
        loadMore()
      }
    },
    { root: tableEl, threshold: 0.1 }
  )

  // 观察哨兵元素
  nextTick(() => {
    const sentinelEl = document.querySelector('.scroll-sentinel')
    if (sentinelEl) {
      observer?.observe(sentinelEl)
    }
  })
}

// 排序变化时重置分页
const handleSortChange = ({ prop, order }: { prop: string, order: string }) => {
  sortState.value = { prop, order }
  currentPage.value = 1
  hasMore.value = true
  fetchData(1, false)
}

// Polling - 只刷新当前页数据
const { isActive: isPolling } = useIntervalFn(() => {
  fetchData(1, false)
}, 30000) // 30秒刷新一次

onMounted(() => {
  fetchData(1, false)
  setTimeout(() => {
    initInfiniteScroll()
  }, 500)
})

onUnmounted(() => {
  observer?.disconnect()
})

// Formatters
const getColorClass = (val: number) => {
  if (val > 0) return 'text-green-400'
  if (val < 0) return 'text-red-400'
  return 'text-gray-400'
}

const formatPrice = (val: number) => {
  if (!val) return '0.00'
  if (val >= 1000) return val.toFixed(2)
  if (val >= 10) return val.toFixed(3)
  if (val >= 0.1) return val.toFixed(4)
  return val.toFixed(6)
}

const formatBigNumber = (val: number) => {
  if (!val) return '0'
  const absVal = Math.abs(val)
  if (absVal >= 1.0e9) return (val / 1.0e9).toFixed(2) + 'B'
  if (absVal >= 1.0e6) return (val / 1.0e6).toFixed(2) + 'M'
  if (absVal >= 1.0e3) return (val / 1.0e3).toFixed(2) + 'K'
  return val.toFixed(2)
}

const formatCountdown = (timestamp: number) => {
  if (!timestamp) return '-'
  const diff = timestamp - now.value.getTime()
  if (diff <= 0) return '00:00:00'
  
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)
  
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

const isFundingSoon = (timestamp: number) => {
  if (!timestamp) return false
  const diff = timestamp - now.value.getTime()
  return diff > 0 && diff < 15 * 60 * 1000
}

// Element Plus specific styling
const tableRowClassName = () => {
  return 'hover:bg-gray-800/50'
}
</script>

<style>
.el-table {
  --el-table-bg-color: #0f1219 !important;
  --el-table-tr-bg-color: #0f1219 !important;
  --el-table-header-bg-color: #141a25 !important;
  --el-table-row-hover-bg-color: #1a2030 !important;
  --el-table-border-color: #1e2736 !important;
  --el-table-text-color: #c9d1d9 !important;
  --el-table-header-text-color: #8b949e !important;
}

.el-table .cell {
  padding: 0 8px !important;
}

.el-table th.el-table__cell {
  border-bottom: 1px solid #21262d !important;
}

.el-table .ascending .caret-wrapper .sort-caret.ascending {
  border-bottom-color: #58a6ff !important;
}

.el-table .descending .caret-wrapper .sort-caret.descending {
  border-top-color: #58a6ff !important;
}
</style>