"""fetcher.py — Binance 数据拉取模块

职责：从 Binance 公开 API 获取历史 K 线和实时价格，本地 JSON 缓存。
"""
import json
import time
from pathlib import Path
import httpx

API_BASE = "https://api.binance.com/api/v3"
DATA_DIR = Path("data")


def _cache_path(symbol: str, interval: str, limit: int) -> Path:
    DATA_DIR.mkdir(exist_ok=True)
    return DATA_DIR / f"{symbol}_{interval}_{limit}.json"


def fetch_klines(
    symbol: str = "BTCUSDT",
    interval: str = "1h",
    limit: int = 500,
) -> list[dict]:
    """拉取历史 K 线，优先读缓存。

    返回: [{"time": timestamp_ms, "open":, "high":, "low":, "close":, "volume":}, ...]
    """
    cache = _cache_path(symbol, interval, limit)
    if cache.exists():
        return json.loads(cache.read_text())

    url = f"{API_BASE}/klines"
    params = {"symbol": symbol, "interval": interval, "limit": limit}
    resp = httpx.get(url, params=params, timeout=30)
    resp.raise_for_status()

    candles = []
    for k in resp.json():
        candles.append({
            "time": k[0],
            "open": float(k[1]),
            "high": float(k[2]),
            "low": float(k[3]),
            "close": float(k[4]),
            "volume": float(k[5]),
        })

    cache.write_text(json.dumps(candles, indent=2))
    return candles


def fetch_current_price(symbol: str = "BTCUSDT") -> dict:
    """查询实时价格。返回 {"symbol": "BTCUSDT", "price": 76460.45}"""
    resp = httpx.get(
        f"{API_BASE}/ticker/price",
        params={"symbol": symbol},
        timeout=10,
    )
    resp.raise_for_status()
    data = resp.json()
    return {"symbol": data["symbol"], "price": float(data["price"])}
