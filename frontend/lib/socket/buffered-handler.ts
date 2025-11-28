"use client";

type Handler<T> = (payload: T) => void;

/**
 * Buffers high-frequency events and flushes them on the next microtask.
 * Reduces React state churn when websocket bursts deliver many payloads.
 */
export function createBufferedHandler<T>(
  handler: Handler<T>,
  flushStrategy: "microtask" | "animationFrame" = "microtask",
) {
  const queue: T[] = [];
  let scheduled = false;

  const flush = () => {
    scheduled = false;
    const batch = queue.splice(0);
    for (let index = 0; index < batch.length; index += 1) {
      handler(batch[index]);
    }
  };

  return (payload: T) => {
    queue.push(payload);
    if (scheduled) {
      return;
    }
    scheduled = true;
    if (flushStrategy === "animationFrame" && typeof window !== "undefined") {
      window.requestAnimationFrame(flush);
      return;
    }
    if (typeof queueMicrotask === "function") {
      queueMicrotask(flush);
    } else {
      Promise.resolve().then(flush);
    }
  };
}
