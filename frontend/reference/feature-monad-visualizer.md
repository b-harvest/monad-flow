# Monad Node Pulse - Visualization Feature Specification

## Overview

**Monad Node Pulse** is the flagship real-time visualization for Monad Flow. It turns raw consensus/network telemetry into a cinematic 3D experience so SREs and protocol researchers can see node health, consensus rounds, and failure scenarios without deciphering logs. The experience is powered by **Socket.IO** streams emitted by the backend, stored via Zustand, and rendered through React + Three.js.

### Core Vision
- **Cinematic 3D Experience**: Feel like watching a sci-fi control room—nodes glow, pulses traverse space, and the environment reacts to consensus progress.
- **Real-Time Monitoring**: Every visual element is driven by live Socket.IO events (HTTP fallback exists only for diagnostics). Latency budget < 250 ms end-to-end.
- **Failure Detection**: Leader stalls, timeout chains, and node crashes should be immediately visible via color shifts, particles, and HUD alerts.

---

## Main Features

### 1. 3D Node Visualization

#### 1.1 Node Sphere Representation
Each Monad validator is rendered as a metallic sphere.

- **Geometry**: Radius 1.0, 32 segments (LOD reduces segments when zoomed out).
- **Material**: Standard material with `roughness: 0.3`, `metalness: 0.8`, optional noise/distortion normal map (intensity 0.2–0.4) to create a living surface.
- **Layout**: Nodes form a circular orbit at radius 5. Leader stays at the front-center; validators are evenly distributed. Y-offset jitter ±0.5 keeps the ring organic.
- **Animation**:
  - Idle rotation: 0.2 rad/s on Y axis.
  - Breathing scale: 1.0 ↔ 1.05 on a 2 s loop.
  - Hover: scale 1.1, emissive boost to highlight details.

#### 1.2 Node State Colors

| State | Color | Emissive | Notes |
|-------|-------|----------|-------|
| Leader | `#6E54FF` | 0.5 | Main brand purple, stronger halo |
| Validator (Active) | `#85E6FF` | 0.3 | Cyan accent |
| Validator (Idle) | `#71717A` | 0.1 | Muted gray, low activity |
| Failed | `#EF4444` | 0.8 | Bright red + glitch effect |
| Syncing | `#F59E0B` | 0.4 | Amber to signal catch-up |

#### 1.3 Glow Halo Effect
- Secondary mesh scaled to 1.2 × node size, rendered with `THREE.BackSide` for inverted normals.
- Halo opacity: 0.15 normally, 0.3 for failed nodes.
- Halo pulsing: scale 1.2 → 1.3 over 2 s; failed nodes pulse faster.

---

### 2. Consensus Animation System

#### 2.1 Proposal Beam
Visualizes leader proposals traveling to validators.

- **Effect**: Cylindrical beam connecting leader → validator.
- **Color**: Leader purple.
- **Width**: 0.05 units; fades to zero alpha near receiver.
- **Lifetime**: 1.5 s with ease-out fade.
- **Trigger**: `ProtocolMessage` where `messageType === 1 (Proposal)` from `/api/outbound-message` or `MONAD_CHUNK_EVENT`.

#### 2.2 Vote Ripple
Validators replying with votes produce expanding rings.

- **Effect**: Expanding circle from node center.
- **Color**: Cyan.
- **Radius**: 1.2 → 3.0 units.
- **Opacity**: 0.6 → 0, 1 s duration.
- **Trigger**: `VoteMessage` (messageType 2).

#### 2.3 Consensus Pulse
When a quorum certificate arrives, broadcast a global pulse.

- **Effect**: Spherical shockwave from leader.
- **Color**: Success green `#22C55E`.
- **Speed**: 5 units/s, max radius 15.
- **Duration**: 3 s with bloom.
- **Trigger**: `QuorumCertificate` field resolved (3f+1 votes collected).

---

### 3. Leader Failure Scenario

When timeouts accumulate or a leader falls behind, users must see a dramatic shift.

#### 3.1 Detection Logic
- Timeout: leader fails to issue proposal for ≥5 s after round start.
- `TimeoutMessage` (type 3) or `RoundRecoveryMessage` (type 4) from events stream.
- Backend emits `LEADER_FAILURE` Socket.IO event for clarity.

#### 3.2 Visual Effects

**A. Node Appearance**
- Color snaps to `#EF4444`, emissive 0.8.
- Surface distortion intensity 0.4.
- Rotation becomes erratic: base 0.3 rad/s + `0.1 * sin(time * 0.5)` noise + random wobble on X/Z (±0.05 rad/s).

**B. Warning Particles**
- 50 red particles orbiting the node.
- Size 0.05–0.2, velocity 0.5–1.5 units/s upward spiral.
- Lifetime 2 s with opacity fade.

**C. Halo Intensification**
- Halo opacity 0.5, color oscillates between `#EF4444` and `#FF8EE4`.

**D. Failed Proposal Beam**
- Leader attempts turn into glitchy red beams that break mid-flight with particle bursts.

**E. Emergency Alert UI**
- Toast in top-right: background `rgba(239,68,68,0.1)`, border `2px solid #EF4444`, text “Leader Failure Detected – Round Recovery Initiated”.
- Animation: slide-in → shake → stay 5 s → fade.

#### 3.3 Recovery Sequence
1. **Old Leader Fade Out**: color fades from red to gray, emissive 0.1, drift backwards, particles stop.
2. **New Leader Transition**: highlight next leader with cyan → purple gradient, scale 1.0 → 1.2 → 1.0 bounce, golden flash overlay `#FFAE45` plus mini shockwave.
3. **Stabilization**: all nodes revert to normal breathing animation, success toast “New Leader Elected – Consensus Resumed”.

---

### 4. Background & Environment

1. **Particle Background**: 1000 star particles in a 30×30×30 unit cube. Gentle parallax motion, depth fog fade.
2. **Grid Floor (optional)**: Infinite grid using `rgba(110,84,255,0.1)` lines, fade at 20 units.
3. **Lighting**: Three-point lights (key `[10,10,10]` intensity 1.0, fill `[-10,5,-5]` 0.5, rim `[0,-5,-10]` 0.3) plus ambient 0.3.
4. **Post-Processing**: Bloom threshold 0.2, intensity 0.5, radius 0.9. Chromatic aberration optional but avoid motion sickness.

---

### 5. Camera System

- Default camera: position `[0,5,15]`, FOV 50°, near 0.1, far 1000.
- OrbitControls: pan disabled, zoom allowed (min 10, max 30), max polar angle `Math.PI / 2`, damping 0.05.
- Auto-rotation optional (speed 0.5) with toggle in HUD.

---

### 6. Side Panels & HUD

#### 6.1 Metrics Panel
- **Data**: round, epoch, leader ID, active validator count, network health, TPS, block height, avg block time.
- **Style**: `--color-bg-secondary`, border `--color-border-primary`, radius `--radius-lg`, padding `--space-6`, width 280 px, fixed top-right, slide-in animation.

#### 6.2 Event Log Panel
- Last 50 events: proposal sent, vote received, QC reached, timeout, leader change, node failure, plus raw Socket.IO events.
- Scrollable, auto-scroll to bottom, timestamp in Roboto Mono, colored badges.

#### 6.3 Status Bar
- **Left**: Logo + connection indicator (Socket.IO connected/disconnected) + last event time.
- **Center**: Current UTC time, uptime.
- **Right**: Theme toggle, auto-rotation toggle, fullscreen button.
- Background with blur, height 60 px.

#### 6.4 Node Info Tooltip
- Trigger on hover/focus.
- Data: node ID, role, IP, uptime, last activity, participation rate.
- Style: `--color-bg-elevated`, border `--color-border-secondary`, shadow `--shadow-lg`, 150 ms fade.

---

### 7. Performance Considerations

1. **React.memo/useMemo** for all heavy components (nodes, halos, HUD sections).
2. **InstancedMesh** for nodes/particles to reduce draw calls.
3. **LOD** for spheres and particle counts depending on zoom.
4. **WebSocket throttling**: keep buffer ≤1000 events; aggregate bursty PERF_STAT events; UI updates at 60 FPS max.
5. **Renderer**: `dpr={[1, 2]}`, `frameloop="demand"`, disable shadow auto-update unless needed.

---

### 8. Responsive Design

- **Desktop ≥1024px**: full 3D canvas with side panels.
- **Tablet 768–1023**: stack panels below canvas, enable touch gestures.
- **Mobile <768**: simplified controls, condensed HUD, single-finger orbit + pinch zoom, hide heavy particles.

---

### 9. Accessibility

- Keyboard mapping: `Tab` cycle nodes, `Enter/Space` focus details, arrows orbit camera, `+/-` zoom, `Esc` exit modals.
- Screen readers: key regions labeled (`role="region" aria-label="Visualization Canvas"`). Node tooltip content mirrored in DOM for SRs.
- Reduced motion: disable auto-rotation, slow animations, turn off particles.

---

### 10. Error Handling

- WebSocket failures: banner “Connection lost”, auto-retry (exponential backoff), fallback to cached data with timestamp overlay.
- Invalid data: Zod validation; log + toast “Invalid data received – skipped”.
- Three.js errors: ErrorBoundary renders 2D fallback graph with explanation.

---

### 11. Testing

- **Visual Regression**: snapshots for normal, leader failure, recovery, no nodes, single node.
- **Performance**: maintain ≥50 FPS with 20 nodes; memory <500 MB; handle 1000 events/s.
- **Accessibility**: Lighthouse ≥95, WAVE no errors, keyboard-only navigation passes.

---

### 12. Historical Forensics Mode (Timebox Playback)

Incident-response users need to rewind the network and visualize exact moments when anomalies happened. This mode complements live Socket.IO streaming.

#### 12.1 Data Source & API
- Reuse `/api/logs/:type?from=&to=` for all historical data. Frontend issues parallel requests per log bucket (chunk/router/offcpu/etc.) and stitches them client-side.
- Aggregate metrics (TPS, node uptime, CPU) are computed client-side from these logs; no dedicated summary endpoint required.
- Cached snapshots should be stored in browser cache/IndexedDB to enable offline replay once downloaded.

#### 12.2 UI/UX Requirements
- Date/time range picker (last 5m/1h/24h + custom absolute range). Dual calendar with timezone selection.
- Timeline scrubber showing stacked bars for proposals/votes/timeouts. Drag to scrub; click markers to jump to exact timestamp.
- Playback controls: play/pause, step forward/back (±1s, ±1 round), speed (0.25×, 0.5×, 1×, 2×, 4×). When playback active, live feed pauses but badge shows "LIVE AVAILABLE".
- Overlay event list pinned under the scrubber summarizing spikes (timeouts clusters, leader changes). Clicking an item jumps camera to relevant node and timestamp.
- Visual distinction between historical and live states (e.g., desaturated palette until returning to LIVE).

#### 12.3 Technical Workflow
```
Range selection → fetch snapshots → cache in indexedDB/local store
    ↓
Buffered playback controller (zustand store) → scheduler emits events at playback rate
    ↓
Visualization layer consumes playback stream (same interface as Socket.IO)
```
- Playback controller must coalesce duplicate events (e.g., rapid PERF_STAT updates) to maintain 60 FPS.
- Supports offline mode: once snapshots fetched they can be replayed without network.
- Provide REST fallback for critical metrics if the historical API fails (toast + degrade gracefully).

#### 12.4 Analytics Overlays
- Heatmap layer highlighting nodes with most failures in selected window.
- Metrics panel adds "range summary" subsection: total proposals, consensus success rate, average latency, node restart count.
- Event log filters switch to "Historic" mode, showing only events in the selected timeframe.

#### 12.5 Testing & Acceptance
- Validate playback accuracy with reference fixtures (expected order of events). Unit tests comparing recorded vs. replayed sequences.
- Performance test: 10-minute window with 1000 events should playback smoothly at 1×.
- UX test script ensures switching between LIVE ↔ Historical is intuitive and no stale state leaks.

#### 12.6 Roadmap Alignment
- Historical mode is feature-complete for V1.1 (post-initial release) but dependencies (snapshot API, storage) must be planned now to avoid rework.
- Future V2 ideas remain:
  - VR mode (Quest)
  - 2D topology graph overlay
  - Custom color themes per workspace
  - Spatial audio cues for consensus events
  - Multi-chain switcher
  - Screenshot/video export

---

## Visual Reference / Mood Board

Inspiration sources: Blockchain.com explorer (flow), Cyberpunk 2077 UI (neon glow), Tron Legacy (grid floor), Stellaris (space strategy aesthetic), Grafana dashboards (professional metrics).

Color mood: dark space backgrounds, neon cyan/purple, intense red alerts, vibrant green successes.

---

## Data Flow Summary

```
Socket.IO connection
    ↓
Socket.IO events (backend)
    ↓
useWebSocket hook → Zustand store (events-store)
    ↓
useMonadEvents hook → React components
    ↓
Three.js scene update → Render
```

---

## Implementation Phases

1. **Phase 1 – Core Visualization (Week 1)**
   - Canvas, camera, lighting
   - Node spheres + base colors
   - Socket.IO connection + Zustand event store
2. **Phase 2 – Animations (Week 2)**
   - Proposal beam, vote ripple, consensus pulse
   - Node rotation/glow micro-interactions
3. **Phase 3 – Leader Failure (Week 3)**
   - Detection logic
   - Failure visuals + UI alerts
   - Recovery sequence
4. **Phase 4 – Polish & Optimization (Week 4)**
   - HUD/side panels
   - Performance tuning + responsive design
   - Accessibility + error handling

---

## Technical Notes

### Dependencies
```bash
npm install three @react-three/fiber @react-three/drei @react-three/postprocessing
npm install socket.io-client zustand @tanstack/react-query
npm install framer-motion lucide-react
npm install zod date-fns numbro
```

### Environment Variables
```env
NEXT_PUBLIC_WS_URL=ws://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3000/api
NEXT_PUBLIC_ENABLE_3D=true
NEXT_PUBLIC_MAX_NODES=20
```

---

## Definition of Done

1. Nodes render with real-time data + consistent states.
2. Leader failure flow demonstrably works end-to-end.
3. Socket.IO stream powers visualization without stale data.
4. 60 FPS sustained on reference hardware.
5. Zod validation + error surfacing implemented.
6. Accessibility checks pass (keyboard + SR).
7. Responsive layouts working down to mobile.
8. Fallback 2D graph renders when 3D fails.
9. QA sign-off on performance + UX polish.
10. Demo-ready script where stakeholders say “This is insane.”

---

**Let’s build the most stunning blockchain visualization ever created.**
