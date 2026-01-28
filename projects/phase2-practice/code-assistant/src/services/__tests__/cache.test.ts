import { describe, it, expect, beforeEach } from 'vitest';
import { RequestCache } from '@/services/cache';

describe('RequestCache', () => {
  let cache: RequestCache<string>;

  beforeEach(() => {
    cache = new RequestCache<string>({
      ttl: 1000, // 1秒
      maxSize: 3,
    });
  });

  it('should store and retrieve data', () => {
    cache.set('key1', 'value1');
    expect(cache.get('key1')).toBe('value1');
  });

  it('should return null for non-existent key', () => {
    expect(cache.get('nonexistent')).toBeNull();
  });

  it('should expire after TTL', async () => {
    cache.set('key1', 'value1', 100); // 100ms TTL
    expect(cache.get('key1')).toBe('value1');

    // Wait for expiration
    await new Promise((resolve) => setTimeout(resolve, 150));
    expect(cache.get('key1')).toBeNull();
  });

  it('should respect max size', () => {
    cache.set('key1', 'value1');
    cache.set('key2', 'value2');
    cache.set('key3', 'value3');
    expect(cache.size()).toBe(3);

    // Adding 4th item should remove the oldest
    cache.set('key4', 'value4');
    expect(cache.size()).toBe(3);
    expect(cache.get('key1')).toBeNull();
    expect(cache.get('key4')).toBe('value4');
  });

  it('should delete specific key', () => {
    cache.set('key1', 'value1');
    cache.delete('key1');
    expect(cache.get('key1')).toBeNull();
  });

  it('should clear all cache', () => {
    cache.set('key1', 'value1');
    cache.set('key2', 'value2');
    cache.clear();
    expect(cache.size()).toBe(0);
  });

  it('should use withCache correctly', async () => {
    let callCount = 0;
    const fetcher = async () => {
      callCount++;
      return 'result';
    };

    // First call should execute fetcher
    const result1 = await cache.withCache('test', fetcher);
    expect(result1).toBe('result');
    expect(callCount).toBe(1);

    // Second call should use cache
    const result2 = await cache.withCache('test', fetcher);
    expect(result2).toBe('result');
    expect(callCount).toBe(1); // Should not increment
  });

  it('should cleanup expired entries', async () => {
    cache.set('key1', 'value1', 100); // 100ms TTL
    cache.set('key2', 'value2', 1000); // 1s TTL

    await new Promise((resolve) => setTimeout(resolve, 150));

    cache.cleanup();
    expect(cache.get('key1')).toBeNull();
    expect(cache.get('key2')).toBe('value2');
    expect(cache.size()).toBe(1);
  });
});
