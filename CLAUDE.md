# MA Cross Strategy — Go Edition

AI 增强加密货币量化回测系统，从 Python 迁移到 Go。

## 项目定位

四维信号融合的均线交叉策略回测器，从 Binance 拉取 K 线，结合技术指标
和恐惧贪婪指数，输出详细回测报告。

## 技术栈

- **语言**: Go 1.26+
- **CLI**: cobra + pflag
- **终端美化**: lipgloss / bubbletea
- **HTTP**: net/http (stdlib)
- **数据处理**: 自实现或 gonum
- **外部 API**: Binance Public REST + Alternative.me Fear & Greed

## 项目结构 (Go)

```
cmd/ma-cross/           # CLI 入口 (cobra)
internal/
  ├── fetcher/          # Binance K 线拉取 + 缓存
  ├── sentiment/        # 恐惧贪婪指数
  ├── indicators/       # SMA / RSI / MACD
  ├── signals/          # 多维信号融合引擎
  ├── backtester/       # 回测引擎（费用/滑点/夏普）
  └── reporter/         # 报告输出
data/                   # K 线缓存 JSON
```

## 代码规范

- 所有包必须有文档注释（`// Package xxx ...`）
- 导出函数必须有 godoc 风格注释
- 使用 `float64` 处理金融数据，不用 `float32`
- 错误不吞没：函数返回 `(T, error)`，调用方处理
- 外部 API 调用统一带 `context.Context` 支持超时
- 单元测试文件 `*_test.go` 与源文件同目录
- 使用 `testing` 标准库 + 表驱动测试
- 常量集中定义，魔法数字必须有注释

## 当前状态

Python 原版在 `src/` 目录，Go 版在 `internal/` + `cmd/` 逐步重建。
迁移原则：保持原 Python 版的业务逻辑不变，用 Go 惯用方式重写。
