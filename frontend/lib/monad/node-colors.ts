import type { MonadNodeState } from "@/types/monad";

type NodeStateVisual = {
  color: string;
  emissive: string;
  emissiveIntensity: number;
  haloColor: string;
  haloOpacity: number;
  breathingScale: [number, number];
  rotationSpeed: number;
};

export const NODE_STATE_VISUALS: Record<MonadNodeState, NodeStateVisual> = {
  leader: {
    color: "#6E54FF",
    emissive: "#6E54FF",
    emissiveIntensity: 0.8,
    haloColor: "#6E54FF",
    haloOpacity: 0.35,
    breathingScale: [1, 1.08],
    rotationSpeed: 0.25,
  },
  active: {
    color: "#85E6FF",
    emissive: "#85E6FF",
    emissiveIntensity: 0.4,
    haloColor: "#85E6FF",
    haloOpacity: 0.25,
    breathingScale: [0.98, 1.04],
    rotationSpeed: 0.18,
  },
  idle: {
    color: "#71717A",
    emissive: "#71717A",
    emissiveIntensity: 0.15,
    haloColor: "#71717A",
    haloOpacity: 0.12,
    breathingScale: [0.96, 1.02],
    rotationSpeed: 0.12,
  },
  failed: {
    color: "#EF4444",
    emissive: "#EF4444",
    emissiveIntensity: 1,
    haloColor: "#FF8EE4",
    haloOpacity: 0.55,
    breathingScale: [1, 1.12],
    rotationSpeed: 0.35,
  },
  syncing: {
    color: "#F59E0B",
    emissive: "#F59E0B",
    emissiveIntensity: 0.5,
    haloColor: "#FFAE45",
    haloOpacity: 0.3,
    breathingScale: [0.97, 1.05],
    rotationSpeed: 0.2,
  },
};

export const NODE_RING_RADIUS = 5;
