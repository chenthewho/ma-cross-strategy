/**
 * fetcher.ts - 数据拉取模块
 *
 * 职责：从 Binance 公开 API 获取历史 K 线数据，缓存到本地 JSON 文件。
 *
 * Binance API（无需认证）：
 *   GET https://api.binance.com/api/v3/klines
 *   参数：symbol=BTCUSDT & interval=1h & limit=500
 *
 * 返回每条 K 线：[开盘时间, 开盘价, 最高价, 最低价, 收盘价, 成交量, ...]
 *
 * 缓存策略：同一交易对+周期只拉一次，存到 data/ 目录
 */

import * as fs from 'fs';
import * as path from 'path';

/**
 * K 线数据结构
 */
export interface Candle {
  time: number;   // 开盘时间戳（毫秒）
  open: number;   // 开盘价
  high: number;   // 最高价
  low: number;    // 最低价
  close: number;  // 收盘价
  volume: number; // 成交量
}

const API_BASE = 'https://api.binance.com/api/v3/klines';
const DATA_DIR = path.join(process.cwd(), 'data');

/**
 * 从 Binance 拉取历史 K 线数据
 *
 * @param symbol 交易对，如 "BTCUSDT"
 * @param interval K线周期，如 "1h", "4h", "1d"
 * @param limit 获取数量，最大 1000
 */
export async function fetchKlines(
  symbol: string = 'BTCUSDT',
  interval: string = '1h',
  limit: number = 500
): Promise<Candle[]> {
  // === 检查本地缓存 ===
  const cacheFile = path.join(
    DATA_DIR,
    `${symbol}_${interval}_${limit}.json`
  );

  if (fs.existsSync(cacheFile)) {
    console.log(`  读取缓存: ${cacheFile}`);
    const raw = fs.readFileSync(cacheFile, 'utf-8');
    return JSON.parse(raw) as Candle[];
  }

  // === 从 API 拉取 ===
  const url = `${API_BASE}?symbol=${symbol}&interval=${interval}&limit=${limit}`;
  console.log(`  请求: ${url}`);

  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`API 请求失败: ${response.status} ${response.statusText}`);
  }

  const rawData = (await response.json()) as any[];

  // === 解析数据 ===
  // Binance 返回格式：[开盘时间, 开, 高, 低, 收, 成交量, 收盘时间, 成交额, 交易数, 主动买量, 主动买额, 忽略]
  const candles: Candle[] = rawData.map((k) => ({
    time: k[0],
    open: parseFloat(k[1]),
    high: parseFloat(k[2]),
    low: parseFloat(k[3]),
    close: parseFloat(k[4]),
    volume: parseFloat(k[5]),
  }));

  // === 写入缓存 ===
  if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
  }
  fs.writeFileSync(cacheFile, JSON.stringify(candles, null, 2));
  console.log(`  已缓存 ${candles.length} 根 K 线 → ${cacheFile}`);

  return candles;
}
