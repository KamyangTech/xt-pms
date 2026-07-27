import { request } from '@/utils/request';

// ------ 类型定义 ------

/** 公开健康检查 */
export interface SimpleHealth {
  status: string; // "ok" | "degraded" | "down"
}

/** DB 连接池统计 */
export interface PoolStats {
  max_open_connections: number;
  open_connections: number;
  idle_connections: number;
  in_use_connections: number;
  waiting_count: number;
  wait_duration: string;
  max_idle_closed: number;
  max_lifetime_closed: number;
}

/** DB 健康状态 */
export interface DBHealth {
  status: string;
  error?: string;
  pool?: PoolStats;
}

/** 运行时指标 */
export interface RuntimeHealth {
  goroutines: number;
  heap_alloc_mb: string;
  heap_sys_mb: string;
  num_gc: number;
  num_cpu: number;
  go_version: string;
}

/** 完整健康检查 */
export interface FullHealth {
  status: string;
  timestamp: string;
  uptime: string;
  database: DBHealth;
  runtime: RuntimeHealth;
}

/** 延迟指标 */
export interface LatencyMetrics {
  avg: number;
  p50: number;
  p95: number;
  p99: number;
  max: number;
  count: number;
}

/** 请求指标 */
export interface RequestMetrics {
  uptime_seconds: string;
  total_requests: number;
  requests_by_status: Record<string, number>;
  in_flight: number;
  latency_ms: LatencyMetrics;
  slow_requests_total: number;
}

// ------ API 函数 ------

const DebugApi = '/debug';

/** 公开健康检查（无需认证） */
export function getPublicHealth() {
  return request.get<SimpleHealth>({ url: '/health' });
}

/** 完整健康检查（需认证） */
export function getFullHealth() {
  return request.get<FullHealth>({ url: `${DebugApi}/health` });
}

/** 请求指标（需认证） */
export function getMetrics() {
  return request.get<RequestMetrics>({ url: `${DebugApi}/metrics` });
}

/** DB 连接池详情（需认证） */
export function getPoolStats() {
  return request.get<PoolStats>({ url: `${DebugApi}/pool` });
}
