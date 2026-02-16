import { useEffect, useRef } from 'react';

/**
 * Custom hook for automatic data refresh with configurable interval
 * @param callback - Function to call on each refresh
 * @param interval - Refresh interval in milliseconds (default: 30000ms = 30s)
 * @param enabled - Whether auto-refresh is enabled (default: true)
 */
export const useAutoRefresh = (
  callback: () => void | Promise<void>,
  interval: number = 30000,
  enabled: boolean = true
) => {
  const savedCallback = useRef(callback);
  const intervalRef = useRef<ReturnType<typeof setInterval>>();

  // Update callback ref when it changes
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up and clean up interval
  useEffect(() => {
    if (!enabled) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
      return;
    }

    const tick = async () => {
      await savedCallback.current();
    };

    intervalRef.current = setInterval(tick, interval);

    // Cleanup on unmount or when dependencies change
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [interval, enabled]);
};

export default useAutoRefresh;
