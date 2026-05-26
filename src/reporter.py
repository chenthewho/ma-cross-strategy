"""reporter.py — 富文本报告生成器

使用 Rich 输出终端友好的彩色报告。
"""
from datetime import datetime, timezone
from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.text import Text

from src.backtester import BacktestResult
from src.sentiment import get_sentiment_score

console = Console()


def _pct(v: float) -> str:
    color = "green" if v >= 0 else "red"
    return f"[{color}]{v*100:+.2f}%[/{color}]"


def _usd(v: float) -> str:
    return f"${v:,.2f}"


def print_price(symbol: str, price: float):
    """实时价格输出"""
    sent = get_sentiment_score()
    text = Text()
    text.append(f"\n  {symbol}  ", style="bold cyan")
    text.append(f"{_usd(price)}", style="bold yellow")
    text.append(f"\n  市场情绪: {sent['classification']} ({sent['raw_value']})", style="dim")
    console.print(text)


def print_report(
    df,
    result: BacktestResult,
    symbol: str,
    short: int,
    long: int,
    interval: str,
):
    """主报告"""
    console.print()

    # ── 概览面板 ──
    title = Text("MA 交叉策略回测报告", style="bold white on blue")
    overview = Table.grid(padding=(0, 2))
    overview.add_column(style="dim")
    overview.add_column()
    overview.add_row("交易对", f"[cyan]{symbol}[/cyan]")
    overview.add_row("周期", interval)
    overview.add_row("短均线", str(short))
    overview.add_row("长均线", str(long))
    overview.add_row("K 线数", str(len(df)))
    overview.add_row("数据范围", f"{_ts2str(df['time'].iloc[0])} ~ {_ts2str(df['time'].iloc[-1])}")

    # 当前情绪
    sent = get_sentiment_score()
    sent_color = "red" if sent["score"] < -0.3 else "green" if sent["score"] > 0.3 else "yellow"
    overview.add_row("当前情绪", f"[{sent_color}]{sent['classification']} ({sent['raw_value']})[/{sent_color}]")

    console.print(Panel(overview, title=title, border_style="blue"))
    console.print()

    # ── 收益统计 ──
    perf = Table(title="资金统计", border_style="green")
    perf.add_column("指标", style="dim")
    perf.add_column("数值", justify="right")
    perf.add_row("初始资金", _usd(result.initial_capital))
    perf.add_row("最终资金", _usd(result.final_capital))
    perf.add_row("总收益率", _pct(result.total_return))
    perf.add_row("最大回撤", _pct(-result.max_drawdown))
    perf.add_row("夏普比率", f"{result.sharpe_ratio:.2f}")
    perf.add_row("买入持有", _pct(result.buy_hold_return))
    excess = result.total_return - result.buy_hold_return
    perf.add_row("策略超额", _pct(excess))
    if result.sentiment_avg != 0:
        mood = "恐惧" if result.sentiment_avg < -0.2 else "贪婪" if result.sentiment_avg > 0.2 else "中性"
        perf.add_row("持仓情绪均值", f"{result.sentiment_avg:+.2f} ({mood})")
    console.print(perf)
    console.print()

    # ── 交易统计 ──
    buys = [t for t in result.trades if t.type == "buy"]
    sells = [t for t in result.trades if t.type == "sell"]
    pairs = min(len(buys), len(sells))
    wins = sum(
        1 for i in range(pairs)
        if sells[i].value > buys[i].value
    )
    win_rate = (wins / pairs * 100) if pairs > 0 else 0

    trade_table = Table(title="交易统计", border_style="yellow")
    trade_table.add_column("指标", style="dim")
    trade_table.add_column("数值", justify="right")
    trade_table.add_row("总交易笔数", str(len(result.trades)))
    trade_table.add_row("完整买卖次数", str(pairs))
    trade_table.add_row("胜率", f"[{'green' if win_rate > 50 else 'red'}]{win_rate:.1f}%[/]")
    avg_profit = sum(sells[i].value - buys[i].value for i in range(pairs)) / pairs if pairs else 0
    color = "green" if avg_profit > 0 else "red"
    trade_table.add_row("平均每笔盈亏", f"[{color}]{_usd(avg_profit)}[/{color}]")
    console.print(trade_table)
    console.print()

    # ── 交易明细（最近 10 笔） ──
    detail = Table(title="最近交易明细", border_style="dim")
    detail.add_column("时间", style="dim")
    detail.add_column("方向")
    detail.add_column("价格", justify="right")
    detail.add_column("金额", justify="right")
    for t in result.trades[-10:]:
        arrow = "[green]买 ↑[/green]" if t.type == "buy" else "[red]卖 ↓[/red]"
        detail.add_row(
            _ts2str(t.time),
            arrow,
            f"{t.price:,.2f}",
            _usd(t.value),
        )
    console.print(detail)
    console.print()


def print_weights(weights: dict):
    """输出信号权重表"""
    t = Table(title="信号权重配置", border_style="magenta")
    t.add_column("信号源", style="dim")
    t.add_column("权重", justify="right")
    for k, v in weights.items():
        t.add_row(k, f"{v:.0%}")
    console.print(t)
    console.print()


def _ts2str(ts: int) -> str:
    return datetime.fromtimestamp(ts / 1000, tz=timezone.utc).strftime("%Y-%m-%d %H:%M")
