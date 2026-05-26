/**
 * reporter.ts - ASCII 报告生成器
 *
 * 职责：将回测结果格式化为终端友好的 ASCII 文本报告。
 *
 * 输出内容：
 *   1. 概览表头
 *   2. 资金统计（含收益率、最大回撤）
 *   3. 交易统计（次数、胜率）
 *   4. 与买入持有对比
 *   5. 价格走势 + 买卖信号 ASCII 图
 */

import { BacktestResult } from './backtester.js';
import { StrategyCandle } from './strategy.js';

/** 百分比格式化 */
function pct(n: number): string {
  const sign = n >= 0 ? '+' : '';
  return `${sign}${(n * 100).toFixed(2)}%`;
}

/** 金额格式化 */
function usd(n: number): string {
  return '$' + n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

/** 画分隔线 */
const SEP = '═'.repeat(50);

/**
 * 生成完整回测报告
 */
export function generateReport(
  candles: StrategyCandle[],
  result: BacktestResult,
  symbol: string,
  shortPeriod: number,
  longPeriod: number,
  interval: string
): string {
  const lines: string[] = [];

  // ==================== 表头 ====================
  lines.push(SEP);
  lines.push('  MA 交叉策略回测报告');
  lines.push(`  交易对: ${symbol.padEnd(14)} 周期: ${interval}`);
  lines.push(`  短均线: ${shortPeriod.toString().padEnd(14)} 长均线: ${longPeriod}`);
  lines.push(SEP);
  lines.push('');

  // ==================== 资金统计 ====================
  lines.push('  ── 资金统计 ──');
  lines.push(`  初始资金: ${usd(result.initialCapital)}`);
  lines.push(`  最终资金: ${usd(result.finalCapital)}`);
  lines.push(`  总收益率: ${pct(result.totalReturn)}`);
  lines.push(`  最大回撤: ${pct(-result.maxDrawdown)}`);
  lines.push('');

  // ==================== 交易统计 ====================
  const pairs = Math.floor(result.trades.length / 2);
  const winRate = pairs > 0 ? (result.winningTrades / pairs * 100) : 0;
  lines.push('  ── 交易统计 ──');
  lines.push(`  交易次数: ${result.totalTrades} (${pairs} 次完整买卖)`);
  lines.push(`  胜率:     ${winRate.toFixed(1)}%`);
  lines.push('');

  // ==================== 对比基准 ====================
  lines.push('  ── 基准对比 ──');
  lines.push(`  买入持有: ${pct(result.buyAndHoldReturn)}`);
  const excess = result.totalReturn - result.buyAndHoldReturn;
  lines.push(`  策略超额: ${pct(excess)}`);
  lines.push('');

  // ==================== 交易明细 ====================
  lines.push('  ── 交易明细 ──');
  if (result.trades.length === 0) {
    lines.push('  (无交易)');
  } else {
    for (let i = 0; i < result.trades.length; i++) {
      const t = result.trades[i];
      const date = new Date(t.time).toISOString().slice(0, 16).replace('T', ' ');
      const arrow = t.type === 'buy' ? '买 ↑' : '卖 ↓';
      lines.push(
        `  ${date}  ${arrow}  ${t.price.toFixed(2).padStart(10)}  × ${t.amount.toFixed(4).padStart(8)} = ${usd(t.value)}`
      );
    }
  }
  lines.push('');

  // ==================== ASCII 走势图 ====================
  lines.push('  ── 价格走势 ──');
  lines.push(drawChart(candles));
  lines.push('');

  lines.push(SEP);

  return lines.join('\n');
}

/**
 * 绘制简易 ASCII 价格走势图 + 买卖信号标记
 *
 * 原理：归一化价格到 [0, chartHeight]，用字符画折线。
 */
function drawChart(candles: StrategyCandle[]): string {
  // 采样：如果数据太多，降采样到 80 个点
  const MAX_POINTS = 80;
  const step = Math.max(1, Math.floor(candles.length / MAX_POINTS));
  const sampled: StrategyCandle[] = [];
  for (let i = 0; i < candles.length; i += step) {
    sampled.push(candles[i]);
  }

  const chartHeight = 8;
  const closes = sampled.map((c) => c.close);
  const min = Math.min(...closes);
  const max = Math.max(...closes);
  const range = max - min || 1;

  // 归一化价格到 0..chartHeight
  const norm = (price: number) =>
    Math.round(((price - min) / range) * chartHeight);

  const lines: string[] = [];

  // 从上到下画
  for (let row = chartHeight; row >= 0; row--) {
    let line = '  ';
    if (row === chartHeight) line += max.toFixed(0).padStart(6) + '┤';
    else if (row === Math.floor(chartHeight / 2)) line += ' '.repeat(6) + '┤';
    else if (row === 0) line += min.toFixed(0).padStart(6) + '┤';
    else line += ' '.repeat(7);

    for (let i = 0; i < sampled.length; i++) {
      const y = norm(sampled[i].close);
      const prevY = i > 0 ? norm(sampled[i - 1].close) : y;

      if (y === row) {
        // 信号标记优先于折线
        if (sampled[i].signal === 'buy') {
          line += 'B'; // Buy
        } else if (sampled[i].signal === 'sell') {
          line += 'S'; // Sell
        } else {
          line += '─';
        }
      } else if (prevY < row && y > row || prevY > row && y < row) {
        line += '│'; // 竖线连接
      } else {
        line += ' ';
      }
    }
    lines.push(line);
  }

  // X 轴
  lines.push('  ' + ' '.repeat(6) + '└' + '─'.repeat(sampled.length) + '→ 时间');
  lines.push('  ' + ' '.repeat(6) + ' B=买入  S=卖出');

  return lines.join('\n');
}
