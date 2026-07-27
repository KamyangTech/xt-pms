import { request } from '@/utils/request';

// ========== 类型定义 ==========

export interface CheckUpdateResult {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  changelog: string;
  download_url: string;
  sha256_url: string;
  sha256: string;
  file_size: number;
  published_at: string;
}

export interface InstallInfo {
  install_dir: string;
  service_name: string;
  update_mode: string;      // systemd / supervisor / manual
  can_auto_update: boolean;
}

// ========== API ==========

/** 检查新版本 */
export function checkUpdate(currentVersion: string): Promise<CheckUpdateResult> {
  return request.get({ url: '/system/update/check', params: { current: currentVersion } });
}

/** 下载更新包（SSE 流式） */
export function downloadUpdate(downloadURL: string): EventSource {
  // POST 请求的 SSE：使用 EventSource 不支持 POST，改用 fetch + ReadableStream
  // 这里通过一个辅助方式实现
  // 注意：localStorage 中存储的 key 是 'token'，'pms_token' 是 WebSocket 的 Cookie 名称
  const token = localStorage.getItem('token') || '';
  const url = `/api/system/update/download`;
  
  // 用 fetch 建立 SSE 连接
  const controller = new AbortController();
  
  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ download_url: downloadURL }),
    signal: controller.signal,
  });
  
  // 返回 AbortController 用于取消
  return controller as unknown as EventSource;
}

/** 获取 SSE 下载流（返回 Promise<ReadableStream>） */
export function downloadUpdateStream(downloadURL: string, onProgress: (data: any) => void, onDone: (data: any) => void, onError: (err: string) => void): AbortController {
  // 注意：localStorage 中存储的 key 是 'token'，'pms_token' 是 WebSocket 的 Cookie 名称
  const token = localStorage.getItem('token') || '';
  const controller = new AbortController();
  
  fetch('/api/system/update/download', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({ download_url: downloadURL }),
    signal: controller.signal,
  }).then(async (response) => {
    if (!response.ok) {
      const text = await response.text();
      onError(text || `HTTP ${response.status}`);
      return;
    }
    
    const reader = response.body?.getReader();
    if (!reader) {
      onError('浏览器不支持流式读取');
      return;
    }
    
    const decoder = new TextDecoder();
    let buffer = '';
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6));
            if (data.status === 'completed') {
              onDone(data);
            } else if (data.status === 'failed') {
              onError(data.message || '下载失败');
            } else {
              onProgress(data);
            }
          } catch {
            // ignore parse errors
          }
        }
      }
    }
  }).catch((err) => {
    if (err.name !== 'AbortError') {
      onError(err.message || '网络错误');
    }
  });
  
  return controller;
}

/** 执行更新 */
export function applyUpdate(data: {
  from_version: string;
  to_version: string;
  package_path: string;
  sha256: string;
  changelog: string;
  file_size: number;
}): Promise<{ message: string }> {
  return request.post({ url: '/system/update/apply', data });
}

/** 重启服务（主进程 Exit，supervisor 自动拉起，通杀所有部署方式） */
export function restartService(): Promise<{ message: string }> {
  return request.post({ url: '/system/update/restart' });
}

/** 获取部署信息 */
export function getInstallInfo(): Promise<InstallInfo> {
  return request.get({ url: '/system/update/install-info' });
}
