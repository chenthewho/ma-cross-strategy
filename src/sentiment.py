"""sentiment.py — 市场情绪数据模块

数据源：
- Alternative.me Fear & Greed Index（免费，无需 API Key）
- 未来可扩展：CryptoPanic 新闻情绪、Twitter/X 情绪
"""
import httpx

FNG_URL = "https://api.alternative.me/fng/"


def fetch_fear_greed(limit: int = 1) -> list[dict]:
    """获取恐惧贪婪指数。

    返回: [{"timestamp":, "value": 0-100, "classification": "Fear"/"Greed"/...}, ...]
    """
    resp = httpx.get(FNG_URL, params={"limit": limit}, timeout=10)
    resp.raise_for_status()
    data = resp.json()

    results = []
    for d in data.get("data", []):
        results.append({
            "timestamp": int(d["timestamp"]),
            "value": int(d["value"]),
            "classification": d["value_classification"],
        })
    return results


def get_sentiment_score() -> dict:
    """获取当前情绪评分，归一化到 -1 ~ +1。

    -1 = 极度恐惧（市场恐慌，可能底部）
    +1 = 极度贪婪（市场狂热，可能顶部）
    0  = 中性

    返回 {"score": -0.32, "classification": "Fear", "raw_value": 34}
    """
    data = fetch_fear_greed(limit=1)
    if not data:
        return {"score": 0, "classification": "Unknown", "raw_value": 50}

    raw = data[0]["value"]
    score = (raw - 50) / 50  # 0-100 → -1 ~ +1
    return {
        "score": round(score, 4),
        "classification": data[0]["classification"],
        "raw_value": raw,
    }


def attach_sentiment(
    candles: list[dict], interval: str = "1h"
) -> list[dict]:
    """将情绪数据附加到 K 线上。

    根据 interval 决定粒度：
    - 1h/4h: 取当日情绪值
    - 1d: 精确匹配日期

    返回带 "sentiment" 和 "sentiment_class" 字段的 K 线列表。
    """
    from datetime import datetime, timezone

    # 拉取足够天数（最多 365 天）
    days_needed = max(1, len(candles) // 24 + 2) if interval in ("1h", "4h") else len(candles) + 2
    fng = fetch_fear_greed(limit=min(days_needed, 365))

    # 按日期建索引
    fng_by_date = {}
    for f in fng:
        dt = datetime.fromtimestamp(f["timestamp"], tz=timezone.utc)
        fng_by_date[dt.strftime("%Y-%m-%d")] = f

    for c in candles:
        dt = datetime.fromtimestamp(c["time"] / 1000, tz=timezone.utc)
        date_key = dt.strftime("%Y-%m-%d")
        f = fng_by_date.get(date_key)
        if f:
            c["sentiment"] = (f["value"] - 50) / 50
            c["sentiment_class"] = f["classification"]
        else:
            c["sentiment"] = 0
            c["sentiment_class"] = "unknown"

    return candles
