import { describe, it, expect, vi } from 'vitest';
import {
  debounce,
  throttle,
  retry,
  sleep,
  withTimeout,
  formatFileSize,
  generateId,
} from '@/utils/helpers';

describe('Helpers', () => {
  describe('debounce', () => {
    it('should debounce function calls', async () => {
      let callCount = 0;
      const fn = () => callCount++;
      const debouncedFn = debounce(fn, 100);

      debouncedFn();
      debouncedFn();
      debouncedFn();

      expect(callCount).toBe(0);

      await sleep(150);
      expect(callCount).toBe(1);
    });
  });

  describe('throttle', () => {
    it('should throttle function calls', async () => {
      let callCount = 0;
      const fn = () => callCount++;
      const throttledFn = throttle(fn, 100);

      throttledFn();
      throttledFn();
      throttledFn();

      expect(callCount).toBe(1);

      await sleep(150);
      throttledFn();
      expect(callCount).toBe(2);
    });
  });

  describe('retry', () => {
    it('should retry on failure', async () => {
      let attempts = 0;
      const fn = async () => {
        attempts++;
        if (attempts < 3) {
          throw new Error('Fail');
        }
        return 'success';
      };

      const result = await retry(fn, { retries: 3, delay: 10 });
      expect(result).toBe('success');
      expect(attempts).toBe(3);
    });

    it('should throw after max retries', async () => {
      const fn = async () => {
        throw new Error('Always fail');
      };

      await expect(retry(fn, { retries: 2, delay: 10 })).rejects.toThrow(
        'Always fail'
      );
    });
  });

  describe('withTimeout', () => {
    it('should resolve if promise completes in time', async () => {
      const promise = sleep(50).then(() => 'success');
      const result = await withTimeout(promise, 100);
      expect(result).toBe('success');
    });

    it('should reject if promise times out', async () => {
      const promise = sleep(200).then(() => 'success');
      await expect(withTimeout(promise, 100)).rejects.toThrow(
        'Operation timed out'
      );
    });
  });

  describe('formatFileSize', () => {
    it('should format bytes correctly', () => {
      expect(formatFileSize(0)).toBe('0 Bytes');
      expect(formatFileSize(1024)).toBe('1 KB');
      expect(formatFileSize(1024 * 1024)).toBe('1 MB');
      expect(formatFileSize(1536)).toBe('1.5 KB');
    });
  });

  describe('generateId', () => {
    it('should generate unique IDs', () => {
      const id1 = generateId();
      const id2 = generateId();
      expect(id1).not.toBe(id2);
      expect(id1).toMatch(/^\d+-[a-z0-9]+$/);
    });
  });
});
