export interface Env {
  CACHE_KV: KVNamespace;
}

interface HotContract {
  symbol: string;
  price: number;
  spotPrice: number;
  markPriceDeviation: number;
  premiumRate: number;
  priceChangePercent: number;
  relativeStrength: number;
  amplitude24h: number;
  volume24hUsd: number;
  takerBuyRatio: number;
  netTakerVolumeUsd: number;
  fundingRate: number;
  fundingIntervalHours: number;
  annualizedFundingRate: number;
  nextFundingTime: number;
  longShortRatio: number;
}

interface RawTicker {
  symbol: string;
  lastPrice: string;
  priceChangePercent: string;
  highPrice: string;
  lowPrice: string;
  quoteVolume: string;
}

interface RawSpotTicker {
  symbol: string;
  price: string;
}

interface RawPremiumIndex {
  symbol: string;
  markPrice: string;
  lastFundingRate: string;
  nextFundingTime: number;
  time: number;
}

interface RawFundingInfo {
  symbol: string;
  fundingIntervalHours: number;
}

interface RawTakerRatio {
  symbol: string;
  buySellRatio: string;
}

interface RawLongShortRatio {
  symbol: string;
  longShortRatio: string;
}

const API_BASE = "https://api.binance.com";
const DATA_BASE = "https://www.binance.com";
const PROXY_BASE = "https://api.allorigins.win/raw?url="; // CORS代理

function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchJson<T>(url: string, useProxy = true): Promise<T> {
  const fetchUrl = useProxy ? `${PROXY_BASE}${encodeURIComponent(url)}` : url;
  const resp = await fetch(fetchUrl, {
    headers: {
      "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
      "Accept": "application/json",
    },
  });
  if (!resp.ok) throw new Error(`Fetch failed: ${resp.status} ${url}`);
  return resp.json() as Promise<T>;
}

function parseFloatSafe(s: string): number {
  const f = parseFloat(s);
  return isNaN(f) ? 0 : f;
}

async function fetchAllData() {
  const tickerPromise = fetchJson<any[]>(`${API_BASE}/api/v3/ticker/24hr`).catch(e => {
    console.log("Ticker error:", e.message);
    return [] as any[];
  });
  const spotPromise = fetchJson<any[]>(`${API_BASE}/api/v3/ticker/price`).catch(e => {
    console.log("Spot error:", e.message);
    return [] as any[];
  });
  const premiumPromise = fetchJson<any[]>(`${DATA_BASE}/fapi/v1/premiumIndex`).catch(e => {
    console.log("Premium error:", e.message);
    return [] as any[];
  });
  const fundingInfoPromise = fetchJson<any[]>(`${DATA_BASE}/fapi/v1/fundingInfo`).catch(e => {
    console.log("FundingInfo error:", e.message);
    return [] as any[];
  });

  const [tickers, spotTickers, premiumIndices, fundingInfo] = await Promise.all([
    tickerPromise, spotPromise, premiumPromise, fundingInfoPromise
  ]);

  console.log(`Fetched: tickers=${tickers.length}, spots=${spotTickers.length}, premium=${premiumIndices.length}, funding=${fundingInfo.length}`);

  return { tickers, spotTickers, premiumIndices, fundingInfo };
}

async function fetchTakerRatios(symbols: string[]): Promise<Map<string, number>> {
  const result = new Map<string, number>();
  const batchSize = 5; // 减小批次大小
  for (let i = 0; i < symbols.length; i += batchSize) {
    const batch = symbols.slice(i, i + batchSize);
    await Promise.all(batch.map(async (symbol) => {
      try {
        const data = await fetchJson<RawTakerRatio[]>(
          `${DATA_BASE}/futures/data/takerlongshortRatio?symbol=${symbol}&period=1h&limit=1`
        );
        if (data.length > 0) {
          const ratio = parseFloatSafe(data[0].buySellRatio);
          if (ratio > 0) result.set(symbol, ratio);
        }
      } catch {}
    }));
    if (i + batchSize < symbols.length) await sleep(100); // 批次间延迟100ms
  }
  return result;
}

async function fetchLongShortRatios(symbols: string[]): Promise<Map<string, number>> {
  const result = new Map<string, number>();
  const batchSize = 5; // 减小批次大小
  for (let i = 0; i < symbols.length; i += batchSize) {
    const batch = symbols.slice(i, i + batchSize);
    await Promise.all(batch.map(async (symbol) => {
      try {
        const data = await fetchJson<RawLongShortRatio[]>(
          `${DATA_BASE}/futures/data/globalLongShortAccountRatio?symbol=${symbol}&period=1h&limit=1`
        );
        if (data.length > 0) {
          const ratio = parseFloatSafe(data[0].longShortRatio);
          if (ratio > 0) result.set(symbol, ratio);
        }
      } catch {}
    }));
    if (i + batchSize < symbols.length) await sleep(100); // 批次间延迟100ms
  }
  return result;
}

async function buildHotContracts(): Promise<HotContract[]> {
  const { tickers, spotTickers, premiumIndices, fundingInfo } = await fetchAllData();

  if (tickers.length === 0) {
    console.log("No tickers fetched");
    return [];
  }

  const spotMap = new Map<string, number>();
  for (const st of spotTickers) spotMap.set(st.symbol, parseFloatSafe(st.price));

  const premiumMap = new Map<string, RawPremiumIndex>();
  for (const pi of premiumIndices) premiumMap.set(pi.symbol, pi);

  const fundingInfoMap = new Map<string, number>();
  for (const fi of fundingInfo) fundingInfoMap.set(fi.symbol, fi.fundingIntervalHours);

  let btcChangePercent = 0;
  for (const t of tickers) {
    if (t.symbol === "BTCUSDT") {
      btcChangePercent = parseFloatSafe(t.priceChangePercent);
      break;
    }
  }

  const activeSymbols = tickers
    .filter(t => t.symbol.endsWith("USDT") && parseFloatSafe(t.quoteVolume) >= 1000000)
    .map(t => t.symbol);

  const [takerRatioMap, longShortMap] = await Promise.all([
    fetchTakerRatios(activeSymbols),
    fetchLongShortRatios(activeSymbols),
  ]);

  const contracts: HotContract[] = [];
  for (const t of tickers) {
    if (!t.symbol.endsWith("USDT")) continue;
    const price = parseFloatSafe(t.lastPrice);
    const quoteVol = parseFloatSafe(t.quoteVolume);
    if (quoteVol < 1000000) continue;

    const highPrice = parseFloatSafe(t.highPrice);
    const lowPrice = parseFloatSafe(t.lowPrice);
    const priceChangePct = parseFloatSafe(t.priceChangePercent);
    const amplitude = lowPrice > 0 ? (highPrice - lowPrice) / lowPrice : 0;

    const buySellRatio = takerRatioMap.get(t.symbol) || 0;
    const takerBuyRatio = buySellRatio > 0 ? buySellRatio / (1 + buySellRatio) : 0;
    const netTakerVolume = (takerBuyRatio - 0.5) * 2 * quoteVol;

    const spotPrice = spotMap.get(t.symbol) || 0;
    const premiumRate = spotPrice > 0 ? (price - spotPrice) / spotPrice : 0;

    let fundingIntervalHours = 8;
    const fh = fundingInfoMap.get(t.symbol);
    if (fh && fh > 0) fundingIntervalHours = fh;

    let markPriceDeviation = 0, fundingRate = 0, annualizedFundingRate = 0, nextFundingTime = 0;
    const pi = premiumMap.get(t.symbol);
    if (pi) {
      const markPrice = parseFloatSafe(pi.markPrice);
      if (markPrice > 0) markPriceDeviation = (price - markPrice) / markPrice;
      fundingRate = parseFloatSafe(pi.lastFundingRate);
      annualizedFundingRate = fundingRate * (24 / fundingIntervalHours) * 365;
      nextFundingTime = pi.nextFundingTime;
    }

    contracts.push({
      symbol: t.symbol, price, spotPrice, markPriceDeviation, premiumRate,
      priceChangePercent: priceChangePct, relativeStrength: priceChangePct - btcChangePercent,
      amplitude24h: amplitude, volume24hUsd: quoteVol, takerBuyRatio,
      netTakerVolumeUsd: netTakerVolume, fundingRate, fundingIntervalHours,
      annualizedFundingRate, nextFundingTime, longShortRatio: longShortMap.get(t.symbol) || 0,
    });
  }
  return contracts;
}

function sortContracts(data: HotContract[], sortBy: string, order: string): HotContract[] {
  return [...data].sort((a, b) => {
    let va: number | string, vb: number | string;
    switch (sortBy) {
      case "symbol": va = a.symbol; vb = b.symbol; return order === "ascending" ? va.localeCompare(vb as string) : (vb as string).localeCompare(va as string);
      case "price": va = a.price; vb = b.price; break;
      case "priceChangePercent": va = a.priceChangePercent; vb = b.priceChangePercent; break;
      case "relativeStrength": va = a.relativeStrength; vb = b.relativeStrength; break;
      case "amplitude24h": va = a.amplitude24h; vb = b.amplitude24h; break;
      case "premiumRate": va = a.premiumRate; vb = b.premiumRate; break;
      case "takerBuyRatio": va = a.takerBuyRatio; vb = b.takerBuyRatio; break;
      case "netTakerVolumeUsd": va = a.netTakerVolumeUsd; vb = b.netTakerVolumeUsd; break;
      case "volume24hUsd": va = a.volume24hUsd; vb = b.volume24hUsd; break;
      case "fundingRate": va = a.annualizedFundingRate; vb = b.annualizedFundingRate; break;
      case "longShortRatio": va = a.longShortRatio; vb = b.longShortRatio; break;
      default: va = a.volume24hUsd; vb = b.volume24hUsd;
    }
    return order === "ascending" ? (va as number) - (vb as number) : (vb as number) - (va as number);
  });
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const corsHeaders: Record<string, string> = {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    };

    if (request.method === "OPTIONS") {
      return new Response(null, { status: 204, headers: corsHeaders });
    }

    // Debug endpoint
    if (url.pathname === "/debug") {
      const results: Record<string, any> = {};
      try {
        const resp = await fetch(`${API_BASE}/api/v3/ticker/24hr?symbol=BTCUSDT`, {
          headers: { "User-Agent": "Mozilla/5.0" }
        });
        results.ticker = { status: resp.status };
        if (resp.ok) {
          const data = await resp.json() as any;
          results.ticker.symbol = data.symbol;
          results.ticker.price = data.lastPrice;
        }
      } catch (e) {
        results.ticker = { error: (e as Error).message };
      }
      return Response.json(results, { headers: corsHeaders });
    }

    if (url.pathname !== "/api/v1/contracts/hot") {
      return Response.json({ error: "Not found" }, { status: 404, headers: corsHeaders });
    }

    const sortBy = url.searchParams.get("sort_by") || "volume24hUsd";
    const order = url.searchParams.get("order") || "descending";
    const page = Math.max(1, parseInt(url.searchParams.get("page") || "1"));
    const pageSize = Math.min(100, Math.max(1, parseInt(url.searchParams.get("page_size") || "25")));

    const cache = caches.default;
    const cacheKey = new Request("https://api.internal/cache/hot_contracts", request);
    let data: HotContract[];

    const cachedResponse = await cache.match(cacheKey);
    if (cachedResponse) {
      data = await cachedResponse.json();
    } else {
      data = await buildHotContracts();
      const cacheResponse = new Response(JSON.stringify(data), {
        headers: { "Content-Type": "application/json" },
      });
      await cache.put(cacheKey, cacheResponse.clone());
    }

    const sorted = sortContracts(data, sortBy, order);
    const total = sorted.length;
    const start = (page - 1) * pageSize;
    const end = Math.min(start + pageSize, total);

    return Response.json(
      { data: sorted.slice(start, end), total, page, page_size: pageSize },
      { headers: corsHeaders }
    );
  },
};
