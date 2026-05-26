# MA Cross Strategy Backtester v2.0

AI 增强量化回测器 — 多维度信号融合 + 市场情绪 + 费用模拟

## 快速开始

```bash
git clone git@github.com:chenthewho/ma-cross-strategy.git
cd ma-cross-strategy
python3 -m venv .venv && source .venv/bin/activate
pip install click rich pandas httpx
```

## 使用

```bash
# 默认回测（BTCUSDT 5/20 1h + 情绪）
python -m src.main run

# 自定义参数
python -m src.main run ETHUSDT 10 30 4h 500

# 关闭情绪模块，纯技术指标
python -m src.main run BTCUSDT --no-sentiment

# 调整信号阈值（越小越敏感）
python -m src.main run BTCUSDT --threshold 0.2

# 实时价格
python -m src.main price BTCUSDT

# 查看市场情绪
python -m src.main sentiment
```

## 信号引擎

不再只用均线交叉，改为 **四维加权融合**：

| 信号源 | 默认权重 | 说明 |
|--------|----------|------|
| 均线趋势 | 30% | 金叉/死叉 + 趋势方向 |
| RSI | 20% | 超卖(<30)偏多，超买(>70)偏空 |
| MACD | 25% | 柱状图方向变化 |
| 市场情绪 | 25% | 恐惧→逆势买入，贪婪→谨慎卖出 |

综合分数超过 ±0.3 触发买卖信号。

## 回测特性

- 初始资金 $10,000，满仓交易
- 手续费 0.1%，滑点 0.05%
- 夏普比率、最大回撤、买入持有对比
- 持仓期间情绪均值追踪

## 项目结构

```
src/
├── main.py           # CLI 入口 (click)
├── fetcher.py        # Binance 数据拉取 + 实时价格
├── sentiment.py      # 市场情绪 (Fear & Greed Index)
├── indicators.py     # SMA/RSI/MACD 技术指标
├── signals.py        # 多维信号融合引擎
├── backtester.py     # 回测引擎（含费用/滑点）
├── reporter.py       # 富文本报告 (rich)
└── ml/               # ML 预测模型（开发中）
data/                 # K 线缓存
```

## 技术栈

- Python 3.11+
- pandas, rich, click, httpx
- Binance Public API + Alternative.me Fear & Greed

## License

MIT
