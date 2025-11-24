# Design System - Monad Flow

## Brand North Star

Monad Flow의 인터페이스는 **시각적 정확성**, **실시간 반응성**, **신뢰성 있는 데이터 표현**을 최우선으로 한다. 모든 화면은 모니터링 전문가가 단 3초 만에 핵심 이상 징후를 발견할 수 있도록 설계된다.

### 브랜드 원칙
1. **Precision UI**: 모든 수치·상태는 명확한 대비와 모노스페이스 타입으로 노이즈 없이 전달한다.
2. **Kinetic Feedback**: 인터랙션과 애니메이션은 실시간 상태 변화를 직관적으로 보여주는 도구다.
3. **Operational Calm**: 다크 모드 기반의 깊이감 있는 팔레트로 장시간 관찰에도 피로를 최소화한다.
4. **Accessibility First**: 키보드, 스크린리더, 고대비 요구 사항을 기본값으로 둔다.
5. **Performance Conscious**: 60FPS 유지와 빠른 첫 렌더가 모든 UI 토큰 결정의 기준이다.

---

## Identity Snapshot

| Asset | Primary Use | Notes |
|-------|-------------|-------|
| **Britti Sans** | Headlines, hero titles, KPI 스포트라이트 | 대문자 대비가 크므로 타이틀에만 사용, 서브픽스는 Inter로 보완 |
| **Inter** | 본문, 설명 텍스트, 표 기본값 | 다국어 렌더링 안정성 확보, 숫자/단위 표기 시 Roboto Mono와 섞지 않는다 |
| **Roboto Mono** | 라벨, 버튼, 링크, 코드, 데코 수치 | `font-variant-numeric: tabular-nums` 기본 적용 |
| **#6E54FF** | 핵심 브랜드 컬러 | CTA, 선택 상태, 주 그래프 라인 |
| **#85E6FF · #FF8EE4 · #FFAE45** | 보조 액센트 | 실시간 이벤트 유형 별 컬러 태깅 |
| **#0E091C** | 기본 배경 | 3D 시각화, 패널, HUD의 공통 배경 |
| **#FFFFFF / #000000** | 대비용 텍스트 | 다크 모드에서는 #FFFFFF 80% 투명도로 사용 |

---

## Color System

### Brand Palette
```css
:root {
  /* Core Purple */
  --color-primary-900: #0e091c;
  --color-primary-700: #6e54ff; /* Signature Britti highlight */
  --color-primary-500: #8f7bff;
  --color-primary-300: #b8a7ff;
  --color-primary-100: #ddd7fe;
  --color-primary-050: #f3f1ff;

  /* Secondary Accents */
  --color-secondary-cyan: #85e6ff;
  --color-secondary-icy: #b9e3f9;
  --color-secondary-magenta: #ff8ee4;
  --color-secondary-amber: #ffae45;
}
```

### Neutrals & Surfaces (Dark-first)
```css
:root {
  --color-bg-primary: #0e091c;
  --color-bg-secondary: #141128;
  --color-bg-tertiary: #1c1934;
  --color-bg-elevated: #242043;

  --color-border-primary: #312c56;
  --color-border-secondary: #3f3970;
  --color-border-faint: rgba(255, 255, 255, 0.08);

  --color-text-primary: #f6f4ff;
  --color-text-secondary: rgba(255, 255, 255, 0.7);
  --color-text-tertiary: rgba(255, 255, 255, 0.45);
  --color-text-inverse: #000000;
}
```

### Semantic Palette
```css
:root {
  --color-success-500: #22c55e;
  --color-danger-500: #ef4444;
  --color-warning-500: #f59e0b;
  --color-info-500: #3b82f6;

  --color-success-bg: rgba(34, 197, 94, 0.12);
  --color-danger-bg: rgba(239, 68, 68, 0.16);
  --color-warning-bg: rgba(245, 158, 11, 0.14);
  --color-info-bg: rgba(59, 130, 246, 0.16);
}
```

### Gradients & Glow
```css
:root {
  --gradient-primary: linear-gradient(140deg, #6e54ff 0%, #8f7bff 45%, #ff8ee4 100%);
  --gradient-cyan: linear-gradient(135deg, #85e6ff 0%, #6bd3ff 100%);
  --gradient-amber: linear-gradient(135deg, #ffae45 0%, #ff8f45 100%);
  --gradient-overlay: linear-gradient(180deg, rgba(14, 9, 28, 0) 0%, rgba(14, 9, 28, 0.9) 100%);

  --shadow-glow-primary: 0 0 30px rgba(110, 84, 255, 0.45);
  --shadow-glow-cyan: 0 0 28px rgba(133, 230, 255, 0.4);
  --shadow-glow-danger: 0 0 32px rgba(239, 68, 68, 0.45);
}
```

Color usage rules:
- 기본 배경(캔버스, HUD)은 항상 `--color-bg-primary` 또는 overlay 변형을 사용한다.
- CTA, 강조 KPI, 3D 리더 노드에는 `--color-primary-700`을 사용하고, hover 시 gradient를 적용한다.
- WebSocket 이벤트 태그는 secondary 컬러 4종을 순환 배정해 유형 식별을 돕는다.
- 실패/에러 시각화는 danger 팔레트 + primary glow로 이중 대비를 확보한다.

---

## Typography System

### Font Families
```css
:root {
  --font-display: 'Britti Sans', 'Inter', sans-serif;
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  --font-mono: 'Roboto Mono', 'JetBrains Mono', 'Fira Code', monospace;
}
```

Usage notes:
- Britti Sans는 제목과 KPI에만 사용한다. 본문이나 설명 텍스트에서는 Inter를 사용한다.
- Roboto Mono는 버튼, 라벨, 링크, 코드 스니펫, 수치 텍스트에 적용한다. `letter-spacing: 0.02em`으로 가독성을 높인다.

### Type Scale
```css
:root {
  --text-xs: 0.75rem;
  --text-sm: 0.875rem;
  --text-base: 1rem;
  --text-lg: 1.125rem;
  --text-xl: 1.25rem;
  --text-2xl: 1.5rem;
  --text-3xl: 1.875rem;
  --text-4xl: 2.25rem;
  --text-5xl: 3rem;
  --text-6xl: 3.75rem;

  --leading-tight: 1.2;
  --leading-snug: 1.35;
  --leading-normal: 1.5;
  --leading-relaxed: 1.65;
}
```

Component classes:
```css
.text-display-hero {
  font-family: var(--font-display);
  font-size: var(--text-5xl);
  font-weight: 700;
  line-height: var(--leading-tight);
  letter-spacing: -0.02em;
}

.text-title {
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 600;
  line-height: var(--leading-snug);
}

.text-body {
  font-family: var(--font-sans);
  font-size: var(--text-base);
  font-weight: 400;
  line-height: var(--leading-normal);
}

.text-label {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.text-number {
  font-family: var(--font-mono);
  font-size: var(--text-lg);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
```

### Motion & Micro-interaction Typography
- 인터랙션 중 제목이 변할 때는 `0.2s ease-out`으로 scale/opacity를 동기화한다.
- 실시간 수치 업데이트는 `font-variation-settings`/`Opacity` 조정 대신 색상과 weight를 이용한다 (화면 떨림 방지).

---

## Spacing & Layout

- 기본 단위: 4px (`--space-1 = 0.25rem`).
- 여백은 4, 8, 12, 16, 24, 32, 40, 64px 계열을 준수한다.
- 카드/패널 내부는 `var(--space-6)` 이상의 padding을 유지한다 (데이터 밀도를 확보하면서도 시각적 호흡 확보).

```css
:root {
  --space-0: 0;
  --space-1: 0.25rem;
  --space-2: 0.5rem;
  --space-3: 0.75rem;
  --space-4: 1rem;
  --space-5: 1.25rem;
  --space-6: 1.5rem;
  --space-8: 2rem;
  --space-10: 2.5rem;
  --space-12: 3rem;
  --space-16: 4rem;
  --space-20: 5rem;
  --space-24: 6rem;
}
```

### Grid & Breakpoints
```css
:root {
  --container-sm: 640px;
  --container-md: 1024px;
  --container-lg: 1280px;
  --container-xl: 1440px;

  --breakpoint-sm: 640px;
  --breakpoint-md: 768px;
  --breakpoint-lg: 1024px;
  --breakpoint-xl: 1440px;
}
```
- Desktop: 12-column grid, `gap: var(--space-6)`.
- Tablet: 8-column grid, `gap: var(--space-4)`.
- Mobile: 4-column grid, `gap: var(--space-3)`.

---

## Elevation & Radius

```css
:root {
  --radius-sm: 0.25rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
  --radius-xl: 1rem;
  --radius-full: 9999px;

  --shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.12);
  --shadow-sm: 0 3px 8px rgba(0, 0, 0, 0.18);
  --shadow-md: 0 10px 20px rgba(8, 5, 20, 0.55);
  --shadow-lg: 0 18px 35px rgba(8, 5, 20, 0.65);
}
```
- 카드: `--radius-lg`, `--shadow-sm`.
- 모달/HUD: `--radius-xl`, `--shadow-lg` + glow.
- 배지/토글: `--radius-full`.

---

## Motion

```css
:root {
  --duration-instant: 100ms;
  --duration-fast: 150ms;
  --duration-normal: 250ms;
  --duration-slow: 400ms;

  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);
  --ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);
}
```
- 상태 변화(성공/실패)에는 `--ease-spring`을 적용해 실시간 반응을 강조한다.
- 데이터 리프레시 애니메이션은 `frameloop='demand'`와 동기화하여 60FPS를 유지한다.

---

## Component Guidelines

### Buttons
- Primary: `--color-primary-700` 배경, Britti Sans 대신 Roboto Mono 라벨, hover 시 gradient/outline glow.
- Secondary: 투명 배경, `--color-border-secondary` 라인, 텍스트는 Inter Medium.
- Ghost: 배경 없음, hover 시 `--color-bg-tertiary` 적용.
- Icon Buttons: 40px (lg) / 36px (md) / 32px (sm). padding `var(--space-2)`.

### Cards & Panels
- 배경: `--color-bg-secondary`, 보더: `--color-border-primary`.
- Padding: 최소 `var(--space-6)`.
- Hover 시 `transform: translateY(-2px)` + `--shadow-md`.
- 그래프 카드에는 `--gradient-overlay`를 사용해 데이터 시선을 모은다.

### Inputs
- 높이: 40px (md), 44px (lg), 36px (sm).
- 배경: `--color-bg-tertiary`, 포커스 시 border → `--color-primary-700`, glow 적용.
- 라벨: Roboto Mono 라벨, Inter 서브카피.

### Tables
- Head: Britti Sans 600, `text-transform: uppercase`, 배경 `--color-bg-tertiary`.
- Body: Roboto Mono for numbers, Inter for labels.
- Hover: `background-color: rgba(255, 255, 255, 0.03)`.

### Badges & Status Chips
- 컬러 매핑: success/danger/warning/info + secondary palette.
- Typography: Roboto Mono uppercase.
- Padding: `var(--space-1) var(--space-3)`.

### Tooltips & Popovers
- 배경 `--color-bg-elevated`, border `--color-border-secondary`, shadow `--shadow-lg`.
- 화살표는 `linear-gradient(135deg, rgba(255,255,255,0.05), rgba(255,255,255,0.02))` 사용.

### Modals
- overlay `rgba(0, 0, 0, 0.7)` + blur(6px).
- 내부 배경 `--color-bg-elevated`, Padding `var(--space-8)`.
- Show/hide는 scale + opacity + glow 동시 적용.

---

## Data Visualization

- 메인 라인/영역 그래프: `--color-primary-700` + gradient.
- WebSocket 이벤트/노드 상태: secondary 팔레트 순환 (Cyan → Icy → Magenta → Amber).
- 상승/하강 지표는 success/danger 팔레트를 그대로 사용한다.
- 3D 시각화 노드:
  - Leader: `#6E54FF` + halo glow
  - Active Validator: `#85E6FF`
  - Idle: `#71717a`
  - Failed: `#ef4444` + pulsating halo

### Price Card Example
```tsx
<Card className="space-y-4">
  <span className="text-label text-secondary">BTC / USD</span>
  <div className="flex items-end gap-4">
    <span className="text-display-hero text-primary-foreground">$42,358.21</span>
    <Badge variant="success">▲ 2.34%</Badge>
  </div>
  <small className="text-body text-muted">Updated 2s ago</small>
</Card>
```

---

## Accessibility & Performance Guardrails
- 텍스트 대비: 최소 4.5:1, Britti Sans 사용 시 굵기 600 이상을 유지한다.
- 포커스 링: 2px `#85E6FF`, offset 2px, radius inherit.
- 모션 최소화 옵션 (`prefers-reduced-motion`) 시 particle, halo, auto-rotation 모두 비활성화.
- 모든 실시간 데이터 컴포넌트는 16ms 이하 업데이트 루프를 준수한다.

---

## Versioning
- Version: 2.0.0
- Palette Revision: Britti Sans Era (2025-02-02)
- 담당: Frontend/Design Systems
- 다음 점검: 신규 시각화 컴포넌트 도입 시
