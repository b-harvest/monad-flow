import type { MonadNode } from "@/types/monad";

const BASE_NODE_NAMES = [
  "Orion",
  "Vega",
  "Lyra",
  "Atlas",
  "Cypher",
  "Zenith",
  "Quasar",
  "Aurora",
  "Draco",
  "Helix",
  "Krypton",
  "Nova",
];

const ROOT_ID = "validator-1";
const ROOT_HEIGHT = 2.8;
const RING_RADIUS = 5.6;

const nameByIndex = (index: number) =>
  BASE_NODE_NAMES[index % BASE_NODE_NAMES.length] ?? `Node-${index + 1}`;

const createLeaderNode = (): MonadNode => ({
  id: ROOT_ID,
  name: BASE_NODE_NAMES[0],
  role: "leader",
  ip: "10.0.1.100",
  uptimePct: 99.4,
  participationRate: 98.1,
  lastActivity: new Date().toISOString(),
  state: "leader",
  position: [0, ROOT_HEIGHT, 0],
  isLocal: true,
});

export const createValidatorNode = (
  order: number,
  parentId = ROOT_ID,
): MonadNode => ({
  id: `validator-${order}`,
  name: nameByIndex(order - 1),
  role: "validator",
  ip: `10.0.1.${100 + order}`,
  uptimePct: 97 - Math.random() * 2,
  participationRate: 92 + Math.random() * 6,
  lastActivity: new Date(Date.now() - order * 2500).toISOString(),
  state: "active",
  position: [0, 1, 0],
  cluster: "primary",
  parentId,
});

export function arrangeNodeOrbit(nodes: MonadNode[]): MonadNode[] {
  const root =
    nodes.find((node) => node.isLocal) ?? nodes.find((node) => node.role === "leader");
  if (!root) {
    return nodes;
  }

  const validators = nodes
    .filter((node) => node.id !== root.id)
    .sort((a, b) => a.id.localeCompare(b.id));
  const total = Math.max(validators.length, 1);

  return nodes.map((node) => {
    if (node.id === root.id) {
      return { ...node, position: [0, ROOT_HEIGHT, 0] };
    }
    const index = validators.findIndex((validator) => validator.id === node.id);
    const angle = (index / total) * Math.PI * 2 - Math.PI / 2;
    const y = 1.1 + Math.sin(angle * 2) * 0.35;
    return {
      ...node,
      position: [RING_RADIUS * Math.cos(angle), y, RING_RADIUS * Math.sin(angle)],
      parentId: root.id,
      cluster: "primary",
    };
  });
}

export function generateInitialNodes(count = 12): MonadNode[] {
  const total = Math.max(3, count);
  const nodes: MonadNode[] = [createLeaderNode()];
  for (let i = 2; i <= total; i += 1) {
    nodes.push(createValidatorNode(i));
  }
  return arrangeNodeOrbit(nodes);
}
