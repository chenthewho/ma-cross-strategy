"""indicators.py — 技术指标计算

基于 pandas，快速计算 SMA / EMA / RSI / MACD。
"""
from typing import Optional
import pandas as pd


def sma(series: pd.Series, period: int) -> pd.Series:
    """简单移动平均"""
    return series.rolling(window=period).mean()


def ema(series: pd.Series, period: int) -> pd.Series:
    """指数移动平均"""
    return series.ewm(span=period, adjust=False).mean()


def rsi(series: pd.Series, period: int = 14) -> pd.Series:
    """相对强弱指标 RSI"""
    delta = series.diff()
    gain = delta.clip(lower=0)
    loss = (-delta).clip(lower=0)
    avg_gain = gain.ewm(alpha=1 / period, adjust=False).mean()
    avg_loss = loss.ewm(alpha=1 / period, adjust=False).mean()
    rs = avg_gain / avg_loss
    return 100 - (100 / (1 + rs))


def macd(
    series: pd.Series,
    fast: int = 12,
    slow: int = 26,
    signal: int = 9,
) -> pd.DataFrame:
    """MACD 指标

    返回 DataFrame 含 macd, signal, histogram 三列。
    """
    ema_fast = ema(series, fast)
    ema_slow = ema(series, slow)
    macd_line = ema_fast - ema_slow
    signal_line = ema(macd_line, signal)
    histogram = macd_line - signal_line
    return pd.DataFrame({
        "macd": macd_line,
        "signal": signal_line,
        "histogram": histogram,
    })


def compute_all(df: pd.DataFrame, short: int = 5, long: int = 20) -> pd.DataFrame:
    """一次性计算所有指标，附加到 DataFrame。

    输入: df 需包含 close 列
    输出: df + sma_short, sma_long, rsi_14, macd, macd_signal, macd_hist
    """
    close = df["close"]
    df = df.copy()
    df["sma_short"] = sma(close, short)
    df["sma_long"] = sma(close, long)
    df["rsi_14"] = rsi(close, 14)
    macd_df = macd(close)
    df["macd"] = macd_df["macd"]
    df["macd_signal"] = macd_df["signal"]
    df["macd_hist"] = macd_df["histogram"]
    return df
