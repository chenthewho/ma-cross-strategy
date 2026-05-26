/**
 * strategy.ts - 均线交叉策略模块
 *
 * 策略逻辑（最经典的均线交叉）：
 *   金叉（Golden Cross）：短均线上穿长均线 → 买入信号
 *   死叉（Death Cross）：短均线下穿长均线 → 卖出信号
 *
 * 实现方式：
 *   逐根 K 线遍历，计算 SMA（简单移动平均），比较当前与前一根的位置关系。
 *
 * SMA 计算：
 *   SMA(N) = 最近 N 根收盘价的算术平均
 *   前 N-1 根 K 线 SMA 为 null（数据不够）
 */

import { Candle } from './fetcher.js';

/** 策略给出的操作信号 */
export type Signal = 'buy' | 'sell' | 'hold';

/** 带策略标记的 K 线 */
export interface StrategyCandle extends Candle {
  shortMA: number | null;   // 短均线值
  longMA: number | null;    // 长均线值
  signal: Signal;           // 当前信号
}

/**
 * 计算简单移动平均 (SMA)
 *
 * @param prices 收盘价数组
 * @param period 周期（如 5 表示 5 日均线）
 * @returns 每根 K 线对应的 SMA 值（前 period-1 根为 null）
 */
function calcSMA(prices: number[], period: number): (number | null)[] {
  const result: (number | null)[] = [];
  for (let i = 0; i < prices.length; i++) {
    if (i < period - 1) {
      result.push(null); // 数据不够，不计算
    } else {
      let sum = 0;
      for (let j = i - period + 1; j <= i; j++) {
        sum += prices[j];
      }
      result.push(sum / period);
    }
  }
  return result;
}

/**
 * 执行均线交叉策略
 *
 * @param candles 原始 K 线数据
 * @param shortPeriod 短均线周期（如 5）
 * @param longPeriod 长均线周期（如 20）
 * @returns 每根 K 线附加上信号和均线值
 */
export function applyStrategy(
  candles: Candle[],
  shortPeriod: number = 5,
  longPeriod: number = 20
): StrategyCandle[] {
  const closes = candles.map((c) => c.close);
  const shortMA = calcSMA(closes, shortPeriod);
  const longMA = calcSMA(closes, longPeriod);

  const result: StrategyCandle[] = [];

  for (let i = 0; i < candles.length; i++) {
    const s = shortMA[i];
    const l = longMA[i];

    let signal: Signal = 'hold';

    // 需要两根 K 线的均线都有值才能判断交叉
    if (i > 0 && s !== null && l !== null) {
      const prevShort = shortMA[i - 1];
      const prevLong = longMA[i - 1];

      if (prevShort !== null && prevLong !== null) {
        // 金叉：之前短 ≤ 长，现在短 > 长
        if (prevShort <= prevLong && s > l) {
          signal = 'buy';
        }
        // 死叉：之前短 ≥ 长，现在短 < 长
        else if (prevShort >= prevLong && s < l) {
          signal = 'sell';
        }
      }
    }

    result.push({
      ...candles[i],
      shortMA: s,
      longMA: l,
      signal,
    });
  }

  return result;
}
