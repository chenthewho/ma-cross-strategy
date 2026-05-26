"""main.py — CLI 入口

用法:
  ma-cross run BTCUSDT 5 20 1h 500         # 回测
  ma-cross price BTCUSDT                     # 实时价格
  ma-cross run ETHUSDT 10 30 4h --no-sentiment  # 关闭情绪
"""
import sys
import time
import click
import pandas as pd
from rich.console import Console
from rich.table import Table

from src.fetcher import fetch_klines, fetch_current_price
from src.sentiment import attach_sentiment, get_sentiment_score
from src.indicators import compute_all
from src.signals import compute_signals
from src.backtester import backtest
from src.reporter import print_price, print_report, print_weights

console = Console()


@click.group()
def cli():
    """MA Cross Strategy — AI 增强量化回测器"""


@cli.command()
@click.argument("symbol", default="BTCUSDT")
@click.argument("short_period", default=5, type=int)
@click.argument("long_period", default=20, type=int)
@click.argument("interval", default="1h")
@click.argument("limit", default=500, type=int)
@click.option("--no-sentiment", is_flag=True, help="关闭情绪模块")
@click.option("--threshold", default=0.3, type=float, help="信号触发阈值")
def run(
    symbol, short_period, long_period, interval, limit,
    no_sentiment, threshold,
):
    """运行回测"""
    symbol = symbol.upper()

    console.print()
    console.print(f"[bold blue]{'═' * 55}[/]")
    console.print(f"[bold cyan]  MA 交叉策略回测器 v2.0  [AI Enhanced][/]")
    console.print(f"  交易对: {symbol}  周期: {interval}  短/长: {short_period}/{long_period}")
    console.print(f"[bold blue]{'═' * 55}[/]\n")

    # ── 阶段 1: 拉数据 ──
    with console.status("[bold green]拉取 K 线数据...[/]"):
        candles = fetch_klines(symbol, interval, limit)
    console.print(f"  ✓ K 线数据: {len(candles)} 根")

    # ── 阶段 2: 情绪 ──
    if not no_sentiment:
        with console.status("[bold green]拉取市场情绪...[/]"):
            candles = attach_sentiment(candles, interval)
        sent = get_sentiment_score()
        console.print(f"  ✓ 当前情绪: [bold]{sent['classification']}[/] ({sent['raw_value']})")

    # ── 阶段 3: 技术指标 ──
    df = pd.DataFrame(candles)
    with console.status("[bold green]计算技术指标...[/]"):
        df = compute_all(df, short_period, long_period)

    # ── 阶段 4: 信号 ──
    weights = {"ma": 0.3, "rsi": 0.2, "macd": 0.25, "sentiment": 0.25 if not no_sentiment else 0}
    if no_sentiment:
        weights = {"ma": 0.4, "rsi": 0.25, "macd": 0.35, "sentiment": 0}
    else:
        print_weights(weights)

    with console.status("[bold green]融合信号...[/]"):
        df = compute_signals(df, weights=weights, threshold=threshold)

    buys = (df["signal"] == "buy").sum()
    sells = (df["signal"] == "sell").sum()
    console.print(f"  ✓ 信号: 买入 {buys}  卖出 {sells}")

    # ── 阶段 5: 回测 ──
    with console.status("[bold green]回测模拟...[/]"):
        result = backtest(df)
    console.print(f"  ✓ 交易: {len(result.trades)} 笔\n")

    # ── 阶段 6: 报告 ──
    print_report(df, result, symbol, short_period, long_period, interval)


@cli.command()
@click.argument("symbol", default="BTCUSDT")
def price(symbol):
    """查询实时价格"""
    symbol = symbol.upper()
    data = fetch_current_price(symbol)
    print_price(symbol, data["price"])


@cli.command()
def sentiment():
    """查看当前市场情绪"""
    s = get_sentiment_score()
    console.print()
    t = Table(title="市场情绪")
    t.add_column("指标", style="dim")
    t.add_column("数值")
    t.add_row("恐惧贪婪指数", str(s["raw_value"]))
    t.add_row("分类", s["classification"])
    t.add_row("归一化分数", f"{s['score']:+.2f}")
    console.print(t)
    console.print()


if __name__ == "__main__":
    cli()
