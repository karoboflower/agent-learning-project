/**
 * 请求缓存服务
 * 用于缓存API请求结果，减少重复调用
 */

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  expiresAt: number;
}

interface CacheOptions {
  ttl?: number; // 缓存时间（毫秒），默认5分钟
  maxSize?: number; // 最大缓存条目数，默认100
}

export class RequestCache<T = any> {
  private cache: Map<string, CacheEntry<T>>;
  private readonly ttl: number;
  private readonly maxSize: number;

  constructor(options: CacheOptions = {}) {
    this.cache = new Map();
    this.ttl = options.ttl || 5 * 60 * 1000; // 默认5分钟
    this.maxSize = options.maxSize || 100; // 默认100条
  }

  /**
   * 生成缓存键
   */
  private generateKey(prefix: string, params: any): string {
    return `${prefix}:${JSON.stringify(params)}`;
  }

  /**
   * 获取缓存
   */
  get(key: string): T | null {
    const entry = this.cache.get(key);

    if (!entry) {
      return null;
    }

    // 检查是否过期
    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  /**
   * 设置缓存
   */
  set(key: string, data: T, ttl?: number): void {
    // 如果缓存已满，删除最早的条目
    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value;
      if (firstKey) {
        this.cache.delete(firstKey);
      }
    }

    const now = Date.now();
    const expiresAt = now + (ttl || this.ttl);

    this.cache.set(key, {
      data,
      timestamp: now,
      expiresAt,
    });
  }

  /**
   * 删除缓存
   */
  delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * 清空所有缓存
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * 获取缓存大小
   */
  size(): number {
    return this.cache.size;
  }

  /**
   * 带缓存的请求包装器
   */
  async withCache<R>(
    key: string,
    fetcher: () => Promise<R>,
    ttl?: number
  ): Promise<R> {
    // 尝试从缓存获取
    const cached = this.get(key) as R | null;
    if (cached !== null) {
      console.log(`[Cache Hit] ${key}`);
      return cached;
    }

    // 执行请求
    console.log(`[Cache Miss] ${key}`);
    const result = await fetcher();

    // 存入缓存
    this.set(key, result as any, ttl);

    return result;
  }

  /**
   * 清理过期缓存
   */
  cleanup(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];

    this.cache.forEach((entry, key) => {
      if (now > entry.expiresAt) {
        keysToDelete.push(key);
      }
    });

    keysToDelete.forEach((key) => this.cache.delete(key));

    if (keysToDelete.length > 0) {
      console.log(`[Cache Cleanup] Removed ${keysToDelete.length} expired entries`);
    }
  }
}

// 创建全局缓存实例
export const agentCache = new RequestCache({
  ttl: 10 * 60 * 1000, // 10分钟
  maxSize: 50, // 最多50条
});

// 定期清理过期缓存（每分钟）
if (typeof window !== 'undefined') {
  setInterval(() => {
    agentCache.cleanup();
  }, 60 * 1000);
}
