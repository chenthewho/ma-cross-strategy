# MA Cross Strategy — Quant SaaS + Agent

## 全局规则（每次对话自动加载）

**提交规则：** 每完成一个功能或修复一个 bug，必须执行：
```bash
git add -A && git commit -m "描述" && git push origin main
```
不要等待、不要询问、不要批量提交。写完就 commit。

## 唯一功能真源

本项目的功能定义**仅**依据 `doc/` 下的三份文档：

1. **系统总体拓扑结构文档** — 定义 SaaS / Agent / Strategy 三层架构边界与通信协议
2. **策略数学引擎文档** — 定义标的信号公式、仓位三态、复利前置条件、无量纲价格计算
3. **进化计算引擎文档** (`进化文档.md`) — 定义 GA 寻优黑盒、八动词契约、坩埚评估

三份文档没有定义的功能**不进入实现**。任何新增能力必须先在三份文档中完成规格定义，再编码。

---

## 工作顺序

1. **涉及策略与回测**：先读取 `doc/` 下对应策略数学引擎文档，理解信号公式与 Step() 语义后再编码
2. **涉及 Go 后端 (SaaS)**：遵守 GORM Code-First 原则，只用 `AutoMigrate` 管理表结构，不手写 DDL
3. **涉及价格计算**：优先无量纲表达（对数收益率 / 比值），避免绝对价格硬编码
4. **涉及架构边界**：保持 SaaS-Strategy-Agent 三层分工，不做预防性解耦——只有明确的跨层需求出现时才抽象接口

---

## 核心约束（五条铁律）

1. **策略必须满足复利前置条件** —— 任何策略在进入回测或实盘前，必须通过复利适配校验（`CapitalPolicy` 与月度注资节奏一致性检查）
2. **回测与实盘调用同一 `Step()` 实现** —— 禁止回测和实盘各自维护一套信号逻辑。`internal/strategies/[name]/step.go` 的 `Step()` 是唯一信号入口
3. **`Step()` 只在 SaaS 侧执行** —— Agent 侧不运行 `Step()`，Agent 仅负责数据采集、预处理、上报。信号计算完全在 SaaS 的 `internal/saas/` 内完成
4. **策略包内部禁止网络/数据库/文件 I/O** —— `internal/strategies/` 下的代码必须是纯函数，只接收数据、产出信号。所有外部交互通过 `internal/saas/` 层的 Adapter 完成
5. **API Key 只能在 `config.agent.yaml`** —— 任何第三方 API 密钥不得硬编码、不得放在环境变量以外的文件中、不得进入 Git 仓库

---

## 代码目录

```
cmd/
├── saas/                  # SaaS HTTP 服务入口 (gin server)
└── agent/                 # Agent 采集进程入口

internal/
├── saas/                  # SaaS 业务逻辑
│   ├── handler/           # HTTP handler (gin)
│   ├── service/           # 业务服务层
│   ├── ga/                # 进化引擎 (EvolvableStrategy + engine)
│   └── middleware/        # 中间件 (auth/ratelimit/logging)
│
├── agent/                 # Agent 采集进程
│   ├── collector/         # 数据采集 (Binance/Alternative.me)
│   └── reporter/          # 向上游 SaaS 上报
│
├── strategy/              # 策略公共接口与抽象
│   └── interface.go       # Strategy 接口定义
│
├── strategies/            # 具体策略实现 (纯函数，无 I/O)
│   └── [策略名]/          # 如 golden_cross / grid / momentum
│       ├── params.go      # 染色体 Params 结构体
│       ├── step.go        # Step() 信号入口 (唯一)
│       ├── indicators.go  # 技术指标计算
│       ├── evolvable.go   # EvolvableStrategy 接口实现
│       └── step_test.go   # Step() 单元测试
│
├── quant/                 # 量化公共库
│   ├── types.go           # SpawnPoint / CapitalPolicy / RiskBounds
│   ├── dca.go             # Ghost DCA 基准计算
│   ├── nav.go             # NAV / Modified Dietz ROI
│   └── mc.go              # 蒙特卡洛模拟
│
└── adapters/
    └── backtest/          # 回测适配器 (Adapter)
        └── adapter.go     # 回测上下文、K 线注入、订单模拟
```

---

## 验证命令

```bash
# 编译检查所有包
go list ./...

# 运行所有测试
go test ./...

# 带竞态检测
go test -race ./...

# 运行特定策略测试
go test ./internal/strategies/...
```

---

## 技术栈

| 组件 | 选型 |
|------|------|
| HTTP 框架 | gin |
| ORM | gorm + postgres |
| 缓存 | go-redis |
| 认证 | golang-jwt |
| 定时任务 | robfig/cron |
| 日志 | zap |
| WebSocket | gorilla/websocket |
| 测试 | testify |
| CLI | cobra |
