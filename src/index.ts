/**
 * index.ts - 均线交叉策略回测主入口
 *
 * 用法：
 *   npm start                       # 默认 BTCUSDT 5 20 1h
 *   npm start ETHUSDT 10 30 4h      # 自定义参数
 *   npm start BTCUSDT 5 20 1h 1000  # 第5参数：拉取K线数量
 *
 * 执行流程（四阶段）：
 *   拉取数据 → 执行策略 → 回测模拟 → 输出报告
 */

import { fetchKlines } from './fetcher.js';
import { applyStrategy } from './strategy.js';
import { backtest } from './backtester.js';
import { generateReport } from './reporter.js';

async function main() {
  // ===== 解析命令行参数 =====
  const args = process.argv.slice(2);
  const symbol = args[0] || 'BTCUSDT';
  const shortPeriod = parseInt(args[1]) || 5;
  const longPeriod = parseInt(args[2]) || 20;
  const interval = args[3] || '1h';
  const limit = parseInt(args[4]) || 500;

  console.log(`
╔══════════════════════════════════════╗
║   MA 交叉策略回测器                  ║
║   交易对: ${symbol.padEnd(10)} 周期: ${interval.padEnd(6)}║
║   短均线: ${shortPeriod.toString().padEnd(10)} 长均线: ${longPeriod.toString().padEnd(6)}║
╚══════════════════════════════════════╝
`);

  // ===== 阶段 1：拉取数据 =====
  console.log('▶ 阶段 1/4: 拉取历史 K 线...');
  const candles = await fetchKlines(symbol, interval, limit);
  console.log(`  获取 ${candles.length} 根 K 线\n`);

  // ===== 阶段 2：执行策略 =====
  console.log('▶ 阶段 2/4: 执行均线交叉策略...');
  const strategyCandles = applyStrategy(candles, shortPeriod, longPeriod);

  const buySignals = strategyCandles.filter((c) => c.signal === 'buy').length;
  const sellSignals = strategyCandles.filter((c) => c.signal === 'sell').length;
  console.log(`  买入信号: ${buySignals}  卖出信号: ${sellSignals}\n`);

  // ===== 阶段 3：回测 =====
  console.log('▶ 阶段 3/4: 回测模拟交易...');
  const result = backtest(strategyCandles);
  console.log(`  完成 ${result.totalTrades} 笔交易\n`);

  // ===== 阶段 4：输出报告 =====
  console.log('▶ 阶段 4/4: 生成报告\n');
  const report = generateReport(
    strategyCandles,
    result,
    symbol,
    shortPeriod,
    longPeriod,
    interval
  );
  console.log(report);
}

main().catch((err) => {
  console.error('执行失败:', err.message);
  process.exit(1);
});
