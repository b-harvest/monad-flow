"use client";

import { useMemo, useRef, useState } from "react";
import * as THREE from "three";
import { Canvas, useFrame } from "@react-three/fiber";
import {
  Billboard,
  Line,
  OrbitControls,
  PerspectiveCamera,
  Stars,
  useCursor,
} from "@react-three/drei";
import type { Line2, LineMaterial } from "three-stdlib";
import { NODE_STATE_VISUALS } from "@/lib/monad/node-colors";
import { useNodePulseStore } from "@/lib/monad/node-pulse-store";
import { usePrefersReducedMotion } from "@/lib/hooks/use-prefers-reduced-motion";

interface NodeSphereProps {
  id: string;
  position: [number, number, number];
  state: keyof typeof NODE_STATE_VISUALS;
  reducedMotion: boolean;
  isSelected: boolean;
  isLocal: boolean;
  onSelect: (nodeId: string) => void;
}

interface ProposalBeamProps {
  start: [number, number, number];
  end: [number, number, number];
  createdAt: number;
  ttl: number;
}

interface VoteRippleProps {
  origin: [number, number, number];
  createdAt: number;
  ttl: number;
}

interface ConsensusPulseProps {
  origin: [number, number, number];
  createdAt: number;
  ttl: number;
}

export function NodePulseScene() {
  const nodes = useNodePulseStore((state) => state.nodes);
  const effects = useNodePulseStore((state) => state.visualEffects);
  const selectedNodeId = useNodePulseStore((state) => state.selectedNodeId);
  const setSelectedNode = useNodePulseStore((state) => state.setSelectedNode);
  const autoRotate = useNodePulseStore(
    (state) => state.preferences.autoRotate,
  );
  const showParticles = useNodePulseStore(
    (state) => state.preferences.showParticles,
  );
  const prefersReducedMotion = usePrefersReducedMotion();

  const nodePositionMap = useMemo(() => {
    const map = new Map<string, [number, number, number]>();
    nodes.forEach((node) => map.set(node.id, node.position));
    return map;
  }, [nodes]);

  const rootNode =
    nodes.find((node) => node.isLocal) ?? nodes.find((node) => node.role === "leader");

  const connectors = useMemo(() => {
    const segments: Array<{
      id: string;
      start: [number, number, number];
      end: [number, number, number];
      color: string;
      width: number;
    }> = [];
    if (rootNode) {
      nodes
        .filter((node) => node.id !== rootNode.id)
        .forEach((node) => {
          segments.push({
            id: `${node.id}-connector`,
            start: rootNode.position,
            end: node.position,
            color: "#6E54FF",
            width: 1.2,
          });
        });
    }
    return segments;
  }, [nodes, rootNode]);

  return (
    <div className="visualization-canvas">
      <Canvas
        dpr={[1, 2]}
        shadows
        style={{ width: "100%", height: "100%" }}
        onPointerMissed={() => setSelectedNode(null)}
      >
        <PerspectiveCamera makeDefault position={[0, 5.5, 14]} fov={48} />
        <color attach="background" args={["#05030d"]} />
        <fog attach="fog" args={["#05030d", 18, 55]} />
        <ambientLight intensity={0.35} />
        <pointLight position={[10, 12, 10]} intensity={1} />
        <pointLight position={[-8, 6, -8]} intensity={0.5} color="#85E6FF" />
        <pointLight position={[0, -6, 0]} intensity={0.35} color="#6E54FF" />

        {showParticles ? (
          <>
            <Stars
              radius={70}
              depth={60}
              factor={3}
              saturation={0.8}
              fade
              speed={prefersReducedMotion ? 0 : 0.3}
            />
            <Stars
              radius={40}
              depth={25}
              factor={1.5}
              saturation={0.6}
              fade
              speed={prefersReducedMotion ? 0 : 0.15}
            />
          </>
        ) : null}

        <GridFloor />

        <group>
          {connectors.map((segment) => (
            <Line
              key={segment.id}
              points={[
                new THREE.Vector3(...segment.start),
                new THREE.Vector3(...segment.end),
              ]}
              color={segment.color}
              lineWidth={segment.width}
              transparent
              opacity={0.65}
            />
          ))}
        </group>

        <group>
          {nodes.map((node) => (
            <NodeSphere
              key={node.id}
              id={node.id}
              state={node.state}
              position={node.position}
              reducedMotion={prefersReducedMotion}
              isSelected={selectedNodeId === node.id}
              isLocal={Boolean(node.isLocal)}
              onSelect={(nodeId) =>
                setSelectedNode(selectedNodeId === nodeId ? null : nodeId)
              }
            />
          ))}
        </group>

        <group>
          {effects.map((effect) => {
            if (effect.type === "proposal" && effect.toNodeId) {
              const start =
                nodePositionMap.get(effect.fromNodeId) ?? ([0, 0, 0] as const);
              const end =
                nodePositionMap.get(effect.toNodeId) ?? ([0, 0, 0] as const);
              return (
                <ProposalBeam
                  key={effect.id}
                  start={start}
                  end={end}
                  createdAt={effect.createdAt}
                  ttl={effect.ttl}
                />
              );
            }
            if (effect.type === "vote") {
              const origin =
                nodePositionMap.get(effect.fromNodeId) ?? ([0, 0, 0] as const);
              return (
                <VoteRipple
                  key={effect.id}
                  origin={origin}
                  createdAt={effect.createdAt}
                  ttl={effect.ttl}
                />
              );
            }
            if (effect.type === "pulse") {
              const origin =
                nodePositionMap.get(effect.fromNodeId) ?? ([0, 0, 0] as const);
              return (
                <ConsensusPulse
                  key={effect.id}
                  origin={origin}
                  createdAt={effect.createdAt}
                  ttl={effect.ttl}
                />
              );
            }
            return null;
          })}
        </group>

        <OrbitControls
          enablePan={false}
          enableZoom
          zoomSpeed={0.5}
          minDistance={9}
          maxDistance={28}
          maxPolarAngle={Math.PI / 2}
          autoRotate={autoRotate && !prefersReducedMotion}
          autoRotateSpeed={0.45}
          enableDamping
          dampingFactor={0.08}
        />
      </Canvas>
    </div>
  );
}

function NodeSphere({
  id,
  state,
  position,
  reducedMotion,
  isSelected,
  isLocal,
  onSelect,
}: NodeSphereProps) {
  const bodyRef = useRef<THREE.Mesh>(null);
  const haloRef = useRef<THREE.Mesh>(null);
  const shellGeometry = useMemo(
    () =>
      createExtrudedRoundedRect({
        width: 1.4,
        height: 1.4,
        radius: 0.35,
        depth: 0.45,
        rotationDeg: 25,
        hole: { width: 0.9, height: 0.9, radius: 0.28 },
      }),
    [],
  );
  const [hovered, setHovered] = useState(false);
  useCursor(hovered && isLocal, "pointer", "auto");
  const visuals = NODE_STATE_VISUALS[state];
  const breathing = visuals.breathingScale;
  const rotationSpeed = reducedMotion ? 0 : visuals.rotationSpeed;
  const baseRotation = THREE.MathUtils.degToRad(40);

  useFrame((stateObj) => {
    const body = bodyRef.current;
    if (!body) return;
    const elapsed = stateObj.clock.getElapsedTime();
    body.rotation.y = baseRotation + rotationSpeed * elapsed;
    body.rotation.x = THREE.MathUtils.degToRad(8);
    const minScale = breathing[0];
    const maxScale = breathing[1];
    const osc = Math.sin(elapsed * (reducedMotion ? 0.2 : 0.7) + position[0]);
    const scale = THREE.MathUtils.lerp(minScale, maxScale, (osc + 1) / 2);
    const finalScale = isSelected ? scale * 1.06 : scale;
    body.scale.set(finalScale, finalScale, finalScale);
    if (haloRef.current) {
      haloRef.current.scale.setScalar(finalScale * 1.4);
      const haloMaterial = haloRef.current.material as THREE.MeshBasicMaterial;
      haloMaterial.opacity =
        visuals.haloOpacity + 0.08 * Math.sin(elapsed * 1.4) + (isSelected ? 0.2 : 0);
    }
  });

  return (
    <group
      position={position}
      name={id}
      onPointerOver={(event) => {
        if (!isLocal) return;
        event.stopPropagation();
        setHovered(true);
      }}
      onPointerOut={() => setHovered(false)}
      onClick={(event) => {
        if (!isLocal) return;
        event.stopPropagation();
        onSelect(id);
      }}
    >
      {isLocal ? (
        <>
          <Billboard>
            <mesh ref={haloRef}>
              <circleGeometry args={[1.2, 64]} />
              <meshBasicMaterial
                color={visuals.haloColor}
                transparent
                opacity={visuals.haloOpacity}
                blending={THREE.AdditiveBlending}
              />
            </mesh>
          </Billboard>
          {state === "leader" ? (
            <Billboard>
              <mesh scale={isSelected ? 2 : 1.7}>
                <circleGeometry args={[1.4, 64]} />
                <meshBasicMaterial color="#6E54FF" transparent opacity={0.35} />
              </mesh>
            </Billboard>
          ) : null}
        </>
      ) : null}
      <mesh ref={bodyRef} geometry={shellGeometry} castShadow receiveShadow>
        <meshPhysicalMaterial
          color={visuals.color}
          emissive={visuals.emissive}
          emissiveIntensity={visuals.emissiveIntensity}
          roughness={0.18}
          metalness={0.55}
          clearcoat={0.6}
          clearcoatRoughness={0.2}
          side={THREE.DoubleSide}
        />
      </mesh>
    </group>
  );
}

function ProposalBeam({ start, end, createdAt, ttl }: ProposalBeamProps) {
  const lineRef = useRef<Line2>(null);
  const points = useMemo(
    () => [new THREE.Vector3(...start), new THREE.Vector3(...end)],
    [start, end],
  );

  useFrame(() => {
    const progress = (Date.now() - createdAt) / ttl;
    const opacity = Math.max(0, 1 - progress);
    if (lineRef.current) {
      const material = lineRef.current.material as LineMaterial;
      material.opacity = opacity;
      material.dashOffset = progress * 2;
    }
  });

  return (
    <Line
      ref={lineRef}
      points={points}
      color="#6E54FF"
      lineWidth={1.6}
      dashed
      dashSize={0.45}
      gapSize={0.25}
      transparent
      opacity={0.9}
    />
  );
}

function VoteRipple({ origin, createdAt, ttl }: VoteRippleProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  useFrame(() => {
    const progress = (Date.now() - createdAt) / ttl;
    const opacity = Math.max(0, 0.6 - progress);
    const scale = THREE.MathUtils.lerp(1.2, 3.2, progress);
    if (meshRef.current) {
      meshRef.current.scale.setScalar(scale);
      const material = meshRef.current.material as THREE.MeshBasicMaterial;
      material.opacity = opacity;
    }
  });

  return (
    <mesh ref={meshRef} position={origin} rotation={[-Math.PI / 2, 0, 0]}>
      <ringGeometry args={[1, 1.3, 64]} />
      <meshBasicMaterial
        color="#85E6FF"
        transparent
        opacity={0.6}
        side={THREE.DoubleSide}
      />
    </mesh>
  );
}

function ConsensusPulse({ origin, createdAt, ttl }: ConsensusPulseProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  useFrame(() => {
    const progress = (Date.now() - createdAt) / ttl;
    const opacity = Math.max(0, 0.6 - progress * 0.6);
    const scale = THREE.MathUtils.lerp(2, 15, progress);
    if (meshRef.current) {
      meshRef.current.scale.setScalar(scale);
      const material = meshRef.current.material as THREE.MeshBasicMaterial;
      material.opacity = opacity;
    }
  });

  return (
    <mesh ref={meshRef} position={origin}>
      <sphereGeometry args={[1, 32, 32]} />
      <meshBasicMaterial
        color="#22C55E"
        transparent
        opacity={0.5}
        side={THREE.DoubleSide}
        blending={THREE.AdditiveBlending}
      />
    </mesh>
  );
}

function GridFloor() {
  return (
    <group>
      <mesh
        rotation={[-Math.PI / 2, 0, 0]}
        position={[0, -1.2, 0]}
        receiveShadow
      >
        <planeGeometry args={[60, 60, 1, 1]} />
        <meshStandardMaterial
          color="#090512"
          metalness={0.15}
          roughness={0.9}
          opacity={0.75}
          transparent
        />
      </mesh>
      <gridHelper
        args={[40, 40, "#312c56", "#1c1934"]}
        position={[0, -1.19, 0]}
      />
    </group>
  );
}

function createExtrudedRoundedRect({
  width,
  height,
  radius,
  depth,
  rotationDeg = 0,
  hole,
}: {
  width: number;
  height: number;
  radius: number;
  depth: number;
  rotationDeg?: number;
  hole?: { width: number; height: number; radius: number };
}) {
  const shape = createRoundedRectShape(width, height, radius);
  if (hole) {
    const inner = createRoundedRectShape(hole.width, hole.height, hole.radius);
    shape.holes.push(inner);
  }

  const geometry = new THREE.ExtrudeGeometry(shape, {
    depth,
    bevelEnabled: true,
    bevelThickness: Math.min(0.2, depth * 0.25),
    bevelSize: rClamp(radius * 0.4),
    bevelSegments: 4,
  });
  geometry.center();
  geometry.rotateZ(THREE.MathUtils.degToRad(rotationDeg));
  return geometry;
}

function createRoundedRectShape(
  width: number,
  height: number,
  radius: number,
): THREE.Shape {
  const hw = width / 2;
  const hh = height / 2;
  const r = Math.min(radius, hw, hh);
  const shape = new THREE.Shape();
  shape.moveTo(-hw + r, -hh);
  shape.lineTo(hw - r, -hh);
  shape.quadraticCurveTo(hw, -hh, hw, -hh + r);
  shape.lineTo(hw, hh - r);
  shape.quadraticCurveTo(hw, hh, hw - r, hh);
  shape.lineTo(-hw + r, hh);
  shape.quadraticCurveTo(-hw, hh, -hw, hh - r);
  shape.lineTo(-hw, -hh + r);
  shape.quadraticCurveTo(-hw, -hh, -hw + r, -hh);
  shape.closePath();
  return shape;
}

function rClamp(value: number) {
  return Math.max(0.02, Math.min(value, 0.3));
}
