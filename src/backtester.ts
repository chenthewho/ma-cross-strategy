/**
 * backtester.ts - 回测引擎
 *
 * 职责：用历史 K 线和策略信号模拟真实交易，计算收益和风险指标。
 *
 * 交易规则：
 *   - 初始资金 $10,000，持仓 0
 *   - 买入信号 → 全部资金买入（按当前收盘价）
 *   - 卖出信号 → 全部持仓卖出
 *   - 最后如果还持有，按最后一根收盘价强制平仓
 */

import { StrategyCandle } from './strategy.js';

/** 单笔交易记录 */
export interface Trade {
  time: number;
  type: 'buy' | 'sell';
  price: number;
  amount: number;       // 币数量
  value: number;        // 交易金额
  balanceAfter: number; // 交易后现金余额
}

/** 回测结果 */
export interface BacktestResult {
  initialCapital: number;
  finalCapital: number;
  totalReturn: number;
  maxDrawdown: number;
  trades: Trade[];
  totalTrades: number;
  winningTrades: number;
  buyAndHoldReturn: number;
}

/**
 * 执行回测
 */
export function backtest(
  candles: StrategyCandle[],
  initialCapital: number = 10000
): BacktestResult {
  let cash = initialCapital;
  let coins = 0;
  const trades: Trade[] = [];

  // 记录每一步的资产价值，用于计算回撤
  const equityHistory: number[] = [];

  for (let i = 0; i < candles.length; i++) {
    const c = candles[i];
    const price = c.close;

    // === 处理买入信号 ===
    if (c.signal === 'buy' && coins === 0 && cash > 0) {
      coins = cash / price;
      cash = 0;

      trades.push({
        time: c.time,
        type: 'buy',
        price,
        amount: coins,
        value: coins * price,
        balanceAfter: 0,
      });
    }

    // === 处理卖出信号 ===
    if (c.signal === 'sell' && coins > 0) {
      const soldAmount = coins;   // 先保存数量
      cash = coins * price;
      coins = 0;

      trades.push({
        time: c.time,
        type: 'sell',
        price,
        amount: soldAmount,        // 使用保存的数量
        value: cash,
        balanceAfter: cash,
      });
    }

    // === 记录当前资产（现金 + 持仓市值） ===
    equityHistory.push(cash + coins * price);
  }

  // === 最后如果还持仓，强制平仓 ===
  if (coins > 0) {
    const lastPrice = candles[candles.length - 1].close;
    const finalAmount = coins;
    cash = coins * lastPrice;
    coins = 0;

    trades.push({
      time: candles[candles.length - 1].time,
      type: 'sell',
      price: lastPrice,
      amount: finalAmount,
      value: cash,
      balanceAfter: cash,
    });
  }

  // === 计算总收益 ===
  const finalCapital = cash;
  const totalReturn = (finalCapital - initialCapital) / initialCapital;

  // === 计算最大回撤 ===
  let maxDrawdown = 0;
  let peak = initialCapital;
  for (const equity of equityHistory) {
    if (equity > peak) peak = equity;
    const drawdown = (peak - equity) / peak;
    if (drawdown > maxDrawdown) maxDrawdown = drawdown;
  }

  // === 买入持有收益（对照组） ===
  const firstPrice = candles[0].close;
  const lastPrice = candles[candles.length - 1].close;
  const bhCoins = initialCapital / firstPrice;
  const bhFinal = bhCoins * lastPrice;
  const buyAndHoldReturn = (bhFinal - initialCapital) / initialCapital;

  // === 胜率 ===
  let winningTrades = 0;
  // 每对买卖（i 是 buy, i+1 是 sell）比较盈亏
  for (let i = 0; i < trades.length - 1; i += 2) {
    if (trades[i].type === 'buy' && trades[i + 1].type === 'sell') {
      if (trades[i + 1].value > trades[i].value) {
        winningTrades++;
      }
    }
  }

  return {
    initialCapital,
    finalCapital,
    totalReturn,
    maxDrawdown,
    trades,
    totalTrades: trades.length,
    winningTrades,
    buyAndHoldReturn,
  };
}
