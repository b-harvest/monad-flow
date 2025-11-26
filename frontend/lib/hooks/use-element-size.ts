"use client";

import { useCallback, useEffect, useRef, useState } from "react";

interface Size {
  width: number;
  height: number;
}

export function useElementSize<T extends HTMLElement>() {
  const ref = useRef<T | null>(null);
  const [size, setSize] = useState<Size>({ width: 0, height: 0 });

  const handleResize = useCallback((entries: ResizeObserverEntry[]) => {
    const entry = entries[0];
    if (!entry) return;
    const { width, height } = entry.contentRect;
    setSize({ width, height });
  }, []);

  useEffect(() => {
    if (!ref.current) {
      return;
    }
    const observer = new ResizeObserver(handleResize);
    observer.observe(ref.current);
    return () => observer.disconnect();
  }, [handleResize]);

  return [ref, size] as const;
}
