"use client";

import { useMemo } from "react";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { useElementSize } from "@/lib/hooks/use-element-size";

const projectPoint = (
  latitude: number,
  longitude: number,
  width: number,
  height: number,
) => {
  const x = ((longitude + 180) / 360) * width;
  const y = ((90 - latitude) / 180) * height;
  return { x, y };
};

interface PositionedNode {
  id: string;
  name: string;
  isLocal?: boolean;
  x: number;
  y: number;
  isInactive?: boolean;
}

export function NodePulseMap() {
  const nodes = useNodePulseStore((state) => state.nodes);
  const effects = useNodePulseStore((state) => state.visualEffects);
  const setSelectedNode = useNodePulseStore((state) => state.setSelectedNode);
  const [ref, size] = useElementSize<HTMLDivElement>();

  const positionedNodes = useMemo(() => {
    if (!size.width || !size.height) {
      return [] as Array<
        PositionedNode & { isInactive: boolean; lastActivity?: string }
      >;
    }
    const now = Date.now();
    return nodes
      .filter(
        (node) =>
          typeof node.latitude === "number" &&
          typeof node.longitude === "number",
      )
      .map((node) => {
        const { x, y } = projectPoint(
          node.latitude ?? 0,
          node.longitude ?? 0,
          size.width,
          size.height,
        );
        const last = node.lastActivity
          ? new Date(node.lastActivity).getTime()
          : 0;
        const isInactive = !node.isLocal && now - last > 60_000;
        return {
          id: node.id,
          name: node.name,
          isLocal: node.isLocal,
          x,
          y,
          isInactive,
          lastActivity: node.lastActivity,
        };
      });
  }, [nodes, size.width, size.height]);

  const nodePositionMap = useMemo(() => {
    const map = new Map<string, { x: number; y: number }>();
    positionedNodes.forEach((node) => {
      map.set(node.id, { x: node.x, y: node.y });
    });
    return map;
  }, [positionedNodes]);

  const chunkFlows = useMemo(() => {
    if (!size.width || !size.height) {
      return [];
    }
    return effects
      .filter((effect) => effect.type === "chunk" && effect.toNodeId)
      .map((effect) => {
        const from = effect.fromNodeId
          ? nodePositionMap.get(effect.fromNodeId)
          : undefined;
        const to = effect.toNodeId
          ? nodePositionMap.get(effect.toNodeId)
          : undefined;
        if (!from || !to) {
          return null;
        }
        const dx = to.x - from.x;
        const dy = to.y - from.y;
        const length = Math.hypot(dx, dy);
        const angle = (Math.atan2(dy, dx) * 180) / Math.PI;
        return {
          id: effect.id,
          createdAt: effect.createdAt,
          ttl: effect.ttl,
          from,
          to,
          length,
          angle,
          direction: effect.direction ?? "inbound",
        };
      })
      .filter(Boolean) as Array<{
      id: string;
      from: { x: number; y: number };
      to: { x: number; y: number };
      length: number;
      angle: number;
      direction: "inbound" | "outbound";
    }>;
  }, [effects, nodePositionMap, size.width, size.height]);

  return (
    <div className="visualization-map" ref={ref}>
      <div className="world-map-layer" aria-hidden>
        <img src="/world-map-dark.svg" alt="" />
      </div>
      <div className="world-map-overlay">
        {positionedNodes.map((node) => (
          <button
            key={node.id}
            type="button"
            className={`map-node ${node.isLocal ? "local" : ""} ${node.isInactive ? "faded" : ""}`}
            style={{
              left: `${node.x}px`,
              top: `${node.y}px`,
            }}
            onClick={() => {
              if (node.isLocal) {
                setSelectedNode(node.id);
              }
            }}
            aria-label={node.isLocal ? "Open node telemetry" : undefined}
          >
            <span className="map-node-dot" />
            <span className="map-node-label">{node.name}</span>
          </button>
        ))}
      </div>
      <div className="chunk-flow-layer">
        {chunkFlows.map((flow) => (
          <div
            key={flow.id}
            className={`chunk-flow ${flow.direction}`}
            style={{
              left: `${flow.from.x}px`,
              top: `${flow.from.y - 1}px`,
              width: `${flow.length}px`,
              transform: `rotate(${flow.angle}deg)`,
            }}
          >
            <span className="chunk-flow-line" />
            <span className="chunk-flow-head" />
          </div>
        ))}
      </div>
    </div>
  );
}
