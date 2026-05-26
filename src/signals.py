"""signals.py — 多维信号融合引擎

不再只用均线交叉，而是多维度打分：
  1. 均线趋势 (0~1)
  2. RSI 超买超卖 (-1~+1)
  3. MACD 方向 (0~1)
  4. 市场情绪 (-1~+1，恐惧=反向买入，贪婪=谨慎)

加权融合后，超过阈值触发交易。
"""
import pandas as pd
import numpy as np


def ma_cross_signal(df: pd.DataFrame, i: int) -> float:
    """均线交叉信号。金叉 → +1，死叉 → -1，无 → 0"""
    if i < 1:
        return 0.0
    s, l = df["sma_short"].iloc[i], df["sma_long"].iloc[i]
    ps, pl = df["sma_short"].iloc[i - 1], df["sma_long"].iloc[i - 1]
    if pd.isna(s) or pd.isna(l) or pd.isna(ps) or pd.isna(pl):
        return 0.0
    if ps <= pl and s > l:
        return 1.0   # 金叉
    if ps >= pl and s < l:
        return -1.0  # 死叉
    # 持有状态：短 > 长 = 偏多
    return 0.3 if s > l else -0.3


def rsi_signal(df: pd.DataFrame, i: int) -> float:
    """RSI 信号。超卖 → 买入倾向，超买 → 卖出倾向"""
    r = df["rsi_14"].iloc[i]
    if pd.isna(r):
        return 0.0
    if r < 30:
        return 1.0 - (r / 30)   # 越超卖越看多
    if r > 70:
        return -((r - 70) / 30)  # 越超买越看空
    return 0.0


def macd_signal(df: pd.DataFrame, i: int) -> float:
    """MACD 信号。柱状图翻正 → 看多，翻负 → 看空"""
    if i < 1:
        return 0.0
    h = df["macd_hist"].iloc[i]
    ph = df["macd_hist"].iloc[i - 1]
    if pd.isna(h) or pd.isna(ph):
        return 0.0
    if ph <= 0 and h > 0:
        return 1.0
    if ph >= 0 and h < 0:
        return -1.0
    return 0.3 if h > 0 else -0.3


def sentiment_signal(df: pd.DataFrame, i: int) -> float:
    """情绪信号。极度恐惧 → 逆势买入，极度贪婪 → 谨慎卖出"""
    s = df["sentiment"].iloc[i] if "sentiment" in df.columns else 0
    if pd.isna(s):
        return 0.0
    # 恐惧时反向做多，贪婪时反向做空
    return -s  # sentiment ∈ [-1,1]，取反


def compute_signals(
    df: pd.DataFrame,
    weights: dict | None = None,
    threshold: float = 0.3,
) -> pd.DataFrame:
    """计算每根 K 线的融合信号。

    weights: {"ma": 0.3, "rsi": 0.2, "macd": 0.2, "sentiment": 0.3}
    threshold: 综合信号超过此值触发买卖

    返回 DataFrame 附加字段:
      signal_raw: 原始融合分数
      signal: "buy" / "sell" / "hold"
    """
    if weights is None:
        weights = {"ma": 0.3, "rsi": 0.2, "macd": 0.25, "sentiment": 0.25}

    df = df.copy()
    raw_signals = []
    actions = []

    for i in range(len(df)):
        s_ma = ma_cross_signal(df, i) * weights.get("ma", 0.3)
        s_rsi = rsi_signal(df, i) * weights.get("rsi", 0.2)
        s_macd = macd_signal(df, i) * weights.get("macd", 0.25)
        s_sent = sentiment_signal(df, i) * weights.get("sentiment", 0.25)

        total = s_ma + s_rsi + s_macd + s_sent
        raw_signals.append(round(total, 4))

        if total > threshold:
            actions.append("buy")
        elif total < -threshold:
            actions.append("sell")
        else:
            actions.append("hold")

    df["signal_raw"] = raw_signals
    df["signal"] = actions
    return df
