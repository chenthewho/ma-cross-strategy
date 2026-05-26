"""backtester.py — 回测引擎

模拟真实交易：资金管理、滑点、手续费、最大回撤计算。
"""
from dataclasses import dataclass, field
import pandas as pd


@dataclass
class Trade:
    time: int          # 时间戳 ms
    type: str          # "buy" | "sell"
    price: float
    amount: float      # 币数量
    value: float       # 交易金额
    balance_after: float


@dataclass
class BacktestResult:
    initial_capital: float
    final_capital: float
    total_return: float
    max_drawdown: float
    sharpe_ratio: float
    trades: list[Trade] = field(default_factory=list)
    buy_hold_return: float = 0.0
    sentiment_avg: float = 0.0  # 持仓期间平均情绪


def backtest(
    df: pd.DataFrame,
    initial_capital: float = 10000,
    fee_rate: float = 0.001,  # 0.1% 手续费
    slippage: float = 0.0005,  # 0.05% 滑点
) -> BacktestResult:
    """执行回测。

    df 需包含: close, signal 列
    """
    cash = initial_capital
    coins = 0.0
    trades: list[Trade] = []
    equity_curve: list[float] = []
    sentiment_held: list[float] = []

    for i in range(len(df)):
        row = df.iloc[i]
        price = row["close"]
        signal = row["signal"]
        sentiment = row.get("sentiment", 0)

        # 执行价格含滑点
        exec_price = price * (1 + slippage) if signal == "buy" else price * (1 - slippage)

        # —— 买入 ——
        if signal == "buy" and coins == 0 and cash > 0:
            gross = cash * (1 - fee_rate)
            coins = gross / exec_price
            cash = 0
            trades.append(Trade(
                time=row["time"],
                type="buy",
                price=exec_price,
                amount=coins,
                value=coins * exec_price,
                balance_after=0,
            ))

        # —— 卖出 ——
        if signal == "sell" and coins > 0:
            gross = coins * exec_price
            cash = gross * (1 - fee_rate)
            coins = 0
            trades.append(Trade(
                time=row["time"],
                type="sell",
                price=exec_price,
                amount=trades[-1].amount if trades else 0,
                value=gross,
                balance_after=cash,
            ))

        # 记录持仓时的情绪
        if coins > 0 and not pd.isna(sentiment):
            sentiment_held.append(sentiment)

        equity_curve.append(cash + coins * price)

    # —— 强制平仓 ——
    if coins > 0:
        last_price = df["close"].iloc[-1] * (1 - slippage)
        cash = coins * last_price * (1 - fee_rate)
        trades.append(Trade(
            time=df["time"].iloc[-1],
            type="sell",
            price=last_price,
            amount=coins,
            value=coins * last_price,
            balance_after=cash,
        ))
        coins = 0

    final_capital = cash
    total_return = (final_capital - initial_capital) / initial_capital

    # 最大回撤
    peak = initial_capital
    max_dd = 0.0
    for eq in equity_curve:
        if eq > peak:
            peak = eq
        dd = (peak - eq) / peak
        if dd > max_dd:
            max_dd = dd

    # 夏普比率（简化，假设无风险利率=0）
    if len(equity_curve) > 1:
        returns = pd.Series(equity_curve).pct_change().dropna()
        sharpe = (returns.mean() / returns.std()) * (252 ** 0.5) if returns.std() > 0 else 0
    else:
        sharpe = 0

    # 买入持有
    first = df["close"].iloc[0]
    last = df["close"].iloc[-1]
    bh = (initial_capital / first) * last
    bh_return = (bh - initial_capital) / initial_capital

    return BacktestResult(
        initial_capital=initial_capital,
        final_capital=final_capital,
        total_return=total_return,
        max_drawdown=max_dd,
        sharpe_ratio=round(sharpe, 4),
        trades=trades,
        buy_hold_return=bh_return,
        sentiment_avg=round(sum(sentiment_held) / len(sentiment_held), 4) if sentiment_held else 0,
    )
