# Design System - Monad Flow

## Design Philosophy

**Monad Flow** is a professional blockchain monitoring platform with core values of **clarity**, **efficiency**, and **reliability**.

### Core Principles

1. **Data-First**: Information must be communicated clearly
2. **Speed**: Intuitive UI for fast decision-making
3. **Trust**: Stable and reliable visual language
4. **Accessibility**: Interface accessible to all users
5. **Elegance**: Sophisticated and modern design

---

## Color System

### Brand Colors

```css
/* Primary - Monad Purple/Blue */
--color-primary-50: #f0f4ff
--color-primary-100: #e0e9ff
--color-primary-200: #c7d7fe
--color-primary-300: #a5bbfc
--color-primary-400: #8195f8
--color-primary-500: #6366f1  /* Main brand color */
--color-primary-600: #4f46e5
--color-primary-700: #4338ca
--color-primary-800: #3730a3
--color-primary-900: #312e81

/* Accent - Cyan */
--color-accent-50: #ecfeff
--color-accent-100: #cffafe
--color-accent-200: #a5f3fc
--color-accent-300: #67e8f9
--color-accent-400: #22d3ee
--color-accent-500: #06b6d4  /* Main accent */
--color-accent-600: #0891b2
--color-accent-700: #0e7490
--color-accent-800: #155e75
--color-accent-900: #164e63
```

### Semantic Colors

```css
/* Success - Green (Price Up) */
--color-success-50: #f0fdf4
--color-success-100: #dcfce7
--color-success-200: #bbf7d0
--color-success-300: #86efac
--color-success-400: #4ade80
--color-success-500: #22c55e  /* Main success */
--color-success-600: #16a34a
--color-success-700: #15803d
--color-success-800: #166534
--color-success-900: #14532d

/* Danger - Red (Price Down) */
--color-danger-50: #fef2f2
--color-danger-100: #fee2e2
--color-danger-200: #fecaca
--color-danger-300: #fca5a5
--color-danger-400: #f87171
--color-danger-500: #ef4444  /* Main danger */
--color-danger-600: #dc2626
--color-danger-700: #b91c1c
--color-danger-800: #991b1b
--color-danger-900: #7f1d1d

/* Warning - Orange/Yellow */
--color-warning-50: #fffbeb
--color-warning-100: #fef3c7
--color-warning-200: #fde68a
--color-warning-300: #fcd34d
--color-warning-400: #fbbf24
--color-warning-500: #f59e0b  /* Main warning */
--color-warning-600: #d97706
--color-warning-700: #b45309
--color-warning-800: #92400e
--color-warning-900: #78350f

/* Info - Blue */
--color-info-500: #3b82f6
--color-info-600: #2563eb
```

### Neutral Colors (Dark Theme Primary)

```css
/* Background & Surface */
--color-bg-primary: #0a0a0f      /* Main background */
--color-bg-secondary: #13131a    /* Cards, panels */
--color-bg-tertiary: #1a1a24     /* Hover states */
--color-bg-elevated: #22222e     /* Modals, dropdowns */

/* Borders */
--color-border-primary: #2a2a3a
--color-border-secondary: #3a3a4a
--color-border-accent: #4a4a5a

/* Text */
--color-text-primary: #f5f5f7     /* Main text */
--color-text-secondary: #a1a1aa   /* Secondary text */
--color-text-tertiary: #71717a    /* Disabled, placeholder */
--color-text-inverse: #0a0a0f     /* On light backgrounds */
```

### Light Theme (Optional)

```css
/* Background & Surface */
--color-bg-primary-light: #ffffff
--color-bg-secondary-light: #f9fafb
--color-bg-tertiary-light: #f3f4f6
--color-bg-elevated-light: #ffffff

/* Borders */
--color-border-primary-light: #e5e7eb
--color-border-secondary-light: #d1d5db
--color-border-accent-light: #9ca3af

/* Text */
--color-text-primary-light: #111827
--color-text-secondary-light: #6b7280
--color-text-tertiary-light: #9ca3af
```

### Chart Colors

```css
/* For multiple data series */
--chart-color-1: #6366f1  /* Primary */
--chart-color-2: #06b6d4  /* Accent */
--chart-color-3: #8b5cf6  /* Purple */
--chart-color-4: #ec4899  /* Pink */
--chart-color-5: #f59e0b  /* Orange */
--chart-color-6: #10b981  /* Green */
```

### Gradient Overlays

```css
/* For hero sections, cards */
--gradient-primary: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)
--gradient-accent: linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%)
--gradient-success: linear-gradient(135deg, #10b981 0%, #059669 100%)
--gradient-danger: linear-gradient(135deg, #ef4444 0%, #dc2626 100%)

/* Subtle overlays */
--gradient-overlay: linear-gradient(180deg, rgba(10,10,15,0) 0%, rgba(10,10,15,0.8) 100%)
--gradient-glass: linear-gradient(135deg, rgba(255,255,255,0.05) 0%, rgba(255,255,255,0.02) 100%)
```

---

## Typography

### Font Family

```css
/* Primary */
--font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif

/* Monospace (for numbers, code) */
--font-mono: 'JetBrains Mono', 'Fira Code', 'Monaco', 'Courier New', monospace

/* Display (optional, for headings) */
--font-display: 'Inter', sans-serif
```

### Font Sizes

```css
--text-xs: 0.75rem      /* 12px */
--text-sm: 0.875rem     /* 14px */
--text-base: 1rem       /* 16px */
--text-lg: 1.125rem     /* 18px */
--text-xl: 1.25rem      /* 20px */
--text-2xl: 1.5rem      /* 24px */
--text-3xl: 1.875rem    /* 30px */
--text-4xl: 2.25rem     /* 36px */
--text-5xl: 3rem        /* 48px */
--text-6xl: 3.75rem     /* 60px */
```

### Font Weights

```css
--font-light: 300
--font-normal: 400
--font-medium: 500
--font-semibold: 600
--font-bold: 700
--font-extrabold: 800
```

### Line Heights

```css
--leading-none: 1
--leading-tight: 1.25
--leading-snug: 1.375
--leading-normal: 1.5
--leading-relaxed: 1.625
--leading-loose: 2
```

### Typography Scale

```css
/* Display */
.text-display-large {
  font-size: var(--text-6xl);
  font-weight: var(--font-bold);
  line-height: var(--leading-tight);
  letter-spacing: -0.02em;
}

.text-display {
  font-size: var(--text-5xl);
  font-weight: var(--font-bold);
  line-height: var(--leading-tight);
  letter-spacing: -0.015em;
}

/* Headings */
.text-h1 {
  font-size: var(--text-4xl);
  font-weight: var(--font-bold);
  line-height: var(--leading-tight);
}

.text-h2 {
  font-size: var(--text-3xl);
  font-weight: var(--font-semibold);
  line-height: var(--leading-snug);
}

.text-h3 {
  font-size: var(--text-2xl);
  font-weight: var(--font-semibold);
  line-height: var(--leading-snug);
}

.text-h4 {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  line-height: var(--leading-normal);
}

/* Body */
.text-body-large {
  font-size: var(--text-lg);
  font-weight: var(--font-normal);
  line-height: var(--leading-relaxed);
}

.text-body {
  font-size: var(--text-base);
  font-weight: var(--font-normal);
  line-height: var(--leading-normal);
}

.text-body-small {
  font-size: var(--text-sm);
  font-weight: var(--font-normal);
  line-height: var(--leading-normal);
}

/* Numbers (for prices, metrics) */
.text-number {
  font-family: var(--font-mono);
  font-weight: var(--font-semibold);
  font-variant-numeric: tabular-nums;
}

/* Labels */
.text-label {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.text-caption {
  font-size: var(--text-xs);
  font-weight: var(--font-normal);
  line-height: var(--leading-normal);
}
```

---

## Spacing System

### Base Unit: 4px

```css
--space-0: 0
--space-1: 0.25rem    /* 4px */
--space-2: 0.5rem     /* 8px */
--space-3: 0.75rem    /* 12px */
--space-4: 1rem       /* 16px */
--space-5: 1.25rem    /* 20px */
--space-6: 1.5rem     /* 24px */
--space-8: 2rem       /* 32px */
--space-10: 2.5rem    /* 40px */
--space-12: 3rem      /* 48px */
--space-16: 4rem      /* 64px */
--space-20: 5rem      /* 80px */
--space-24: 6rem      /* 96px */
--space-32: 8rem      /* 128px */
```

### Component-Specific Spacing

```css
/* Padding */
--padding-xs: var(--space-2) var(--space-3)
--padding-sm: var(--space-3) var(--space-4)
--padding-md: var(--space-4) var(--space-6)
--padding-lg: var(--space-6) var(--space-8)
--padding-xl: var(--space-8) var(--space-12)

/* Gap */
--gap-xs: var(--space-2)
--gap-sm: var(--space-3)
--gap-md: var(--space-4)
--gap-lg: var(--space-6)
--gap-xl: var(--space-8)
```

---

## Border Radius

```css
--radius-none: 0
--radius-sm: 0.25rem    /* 4px */
--radius-md: 0.5rem     /* 8px */
--radius-lg: 0.75rem    /* 12px */
--radius-xl: 1rem       /* 16px */
--radius-2xl: 1.5rem    /* 24px */
--radius-full: 9999px   /* Fully rounded */
```

### Component Radius

```css
--radius-button: var(--radius-md)
--radius-input: var(--radius-md)
--radius-card: var(--radius-lg)
--radius-modal: var(--radius-xl)
--radius-badge: var(--radius-full)
```

---

## Shadows

```css
/* Elevation */
--shadow-xs: 0 1px 2px 0 rgba(0, 0, 0, 0.05)
--shadow-sm: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)
--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)
--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.2), 0 4px 6px -4px rgba(0, 0, 0, 0.1)
--shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.2), 0 8px 10px -6px rgba(0, 0, 0, 0.1)
--shadow-2xl: 0 25px 50px -12px rgba(0, 0, 0, 0.25)

/* Glow effects (for dark theme) */
--shadow-glow-primary: 0 0 20px rgba(99, 102, 241, 0.3)
--shadow-glow-success: 0 0 20px rgba(34, 197, 94, 0.3)
--shadow-glow-danger: 0 0 20px rgba(239, 68, 68, 0.3)

/* Inner shadow */
--shadow-inner: inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)
```

---

## Animation & Transitions

### Duration

```css
--duration-instant: 100ms
--duration-fast: 150ms
--duration-normal: 250ms
--duration-slow: 350ms
--duration-slower: 500ms
```

### Easing

```css
--ease-in: cubic-bezier(0.4, 0, 1, 1)
--ease-out: cubic-bezier(0, 0, 0.2, 1)
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1)
--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1)
--ease-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55)
```

### Transitions

```css
/* Standard transitions */
--transition-all: all var(--duration-normal) var(--ease-in-out)
--transition-color: color var(--duration-fast) var(--ease-in-out)
--transition-bg: background-color var(--duration-fast) var(--ease-in-out)
--transition-transform: transform var(--duration-normal) var(--ease-out)
--transition-opacity: opacity var(--duration-fast) var(--ease-in-out)
```

### Animation Presets

```javascript
// Framer Motion variants
export const fadeIn = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
  transition: { duration: 0.25 }
}

export const slideUp = {
  initial: { y: 20, opacity: 0 },
  animate: { y: 0, opacity: 1 },
  exit: { y: -20, opacity: 0 },
  transition: { duration: 0.3, ease: 'easeOut' }
}

export const scaleIn = {
  initial: { scale: 0.95, opacity: 0 },
  animate: { scale: 1, opacity: 1 },
  exit: { scale: 0.95, opacity: 0 },
  transition: { duration: 0.2 }
}

export const slideInFromRight = {
  initial: { x: 100, opacity: 0 },
  animate: { x: 0, opacity: 1 },
  exit: { x: 100, opacity: 0 },
  transition: { type: 'spring', stiffness: 300, damping: 30 }
}
```

---

## Component Guidelines

### Buttons

```tsx
/* Primary Button */
- Background: var(--color-primary-500)
- Text: white
- Hover: var(--color-primary-600)
- Active: var(--color-primary-700)
- Padding: var(--padding-sm)
- Border Radius: var(--radius-button)
- Font Weight: var(--font-semibold)
- Transition: all 150ms

/* Secondary Button */
- Background: transparent
- Border: 1px solid var(--color-border-secondary)
- Text: var(--color-text-primary)
- Hover: var(--color-bg-tertiary)

/* Success/Danger Buttons */
- Use semantic colors (green/red)
- Same structure as primary

/* Ghost Button */
- Background: transparent
- No border
- Hover: var(--color-bg-tertiary)

/* Icon Button */
- Size: 36px × 36px (md), 32px × 32px (sm), 40px × 40px (lg)
- Padding: var(--space-2)
- Border Radius: var(--radius-md)
```

### Cards

```tsx
/* Standard Card */
- Background: var(--color-bg-secondary)
- Border: 1px solid var(--color-border-primary)
- Border Radius: var(--radius-card)
- Padding: var(--space-6)
- Shadow: var(--shadow-sm)
- Hover: subtle shadow increase

/* Elevated Card (hover state) */
- Shadow: var(--shadow-lg)
- Transform: translateY(-2px)
- Transition: all 250ms

/* Glass Card (for overlays) */
- Background: rgba(255, 255, 255, 0.05)
- Backdrop Filter: blur(10px)
- Border: 1px solid rgba(255, 255, 255, 0.1)
```

### Inputs

```tsx
/* Text Input */
- Height: 40px (md), 36px (sm), 44px (lg)
- Background: var(--color-bg-tertiary)
- Border: 1px solid var(--color-border-primary)
- Border Radius: var(--radius-input)
- Padding: 0 var(--space-4)
- Font Size: var(--text-sm)
- Focus: border color → primary, shadow glow

/* Label */
- Font Size: var(--text-sm)
- Font Weight: var(--font-medium)
- Margin Bottom: var(--space-2)
- Color: var(--color-text-secondary)

/* Error State */
- Border: var(--color-danger-500)
- Error Text: var(--text-xs), var(--color-danger-500)
```

### Tables

```tsx
/* Table Header */
- Background: var(--color-bg-tertiary)
- Font Size: var(--text-xs)
- Font Weight: var(--font-semibold)
- Text Transform: uppercase
- Letter Spacing: 0.05em
- Color: var(--color-text-secondary)
- Padding: var(--space-3) var(--space-4)

/* Table Row */
- Border Bottom: 1px solid var(--color-border-primary)
- Hover: var(--color-bg-tertiary)
- Padding: var(--space-4)

/* Table Cell (Numbers) */
- Font Family: var(--font-mono)
- Text Align: right (for numbers)
- Font Variant Numeric: tabular-nums
```

### Badges

```tsx
/* Status Badge */
- Padding: var(--space-1) var(--space-3)
- Border Radius: var(--radius-badge)
- Font Size: var(--text-xs)
- Font Weight: var(--font-semibold)

/* Success Badge (Price Up) */
- Background: rgba(34, 197, 94, 0.1)
- Color: var(--color-success-500)

/* Danger Badge (Price Down) */
- Background: rgba(239, 68, 68, 0.1)
- Color: var(--color-danger-500)
```

### Tooltips

```tsx
- Background: var(--color-bg-elevated)
- Border: 1px solid var(--color-border-secondary)
- Border Radius: var(--radius-md)
- Padding: var(--space-2) var(--space-3)
- Font Size: var(--text-sm)
- Shadow: var(--shadow-lg)
- Max Width: 200px
- Animation: fade in 150ms
```

### Modals

```tsx
/* Overlay */
- Background: rgba(0, 0, 0, 0.7)
- Backdrop Filter: blur(4px)

/* Modal Container */
- Background: var(--color-bg-elevated)
- Border: 1px solid var(--color-border-secondary)
- Border Radius: var(--radius-modal)
- Shadow: var(--shadow-2xl)
- Max Width: 600px (md), 800px (lg), 400px (sm)
- Padding: var(--space-8)
- Animation: scale in + fade in
```

---

## Layout

### Container

```css
--container-xs: 640px
--container-sm: 768px
--container-md: 1024px
--container-lg: 1280px
--container-xl: 1536px
```

### Grid

```css
/* Standard Grid */
- Gap: var(--space-6)
- Columns: 12 (desktop), 4 (mobile)
- Max Width: var(--container-xl)
- Padding: var(--space-6) (mobile), var(--space-8) (desktop)
```

### Breakpoints

```css
--breakpoint-sm: 640px
--breakpoint-md: 768px
--breakpoint-lg: 1024px
--breakpoint-xl: 1280px
--breakpoint-2xl: 1536px
```

---

## Data Visualization

### Price Display

```tsx
/* Price Up */
- Color: var(--color-success-500)
- Icon: ▲ or ↑
- Font: var(--font-mono)

/* Price Down */
- Color: var(--color-danger-500)
- Icon: ▼ or ↓
- Font: var(--font-mono)

/* Percentage Change */
- Font Size: var(--text-sm)
- Background: rgba(color, 0.1)
- Padding: var(--space-1) var(--space-2)
- Border Radius: var(--radius-sm)
```

### Charts

```tsx
/* TradingView Chart Theme */
- Background: var(--color-bg-primary)
- Grid Lines: var(--color-border-primary)
- Text: var(--color-text-secondary)
- Candle Up: var(--color-success-500)
- Candle Down: var(--color-danger-500)
- Volume Bars: 40% opacity of candle color

/* Line Chart */
- Line Width: 2px
- Line Color: var(--color-primary-500)
- Area Fill: gradient (primary, 0.2 opacity → 0)
- Tooltip: var(--color-bg-elevated), shadow-lg
```

---

## Icons

### Library
- **Lucide React** (recommended): `npm install lucide-react`
- **Heroicons**: Alternative option

### Sizes

```css
--icon-xs: 12px
--icon-sm: 16px
--icon-md: 20px
--icon-lg: 24px
--icon-xl: 32px
```

### Usage

```tsx
<Icon
  size={20}
  strokeWidth={2}
  color="currentColor"
/>
```

---

## Accessibility

### Contrast Ratios
- Normal Text: 4.5:1 minimum
- Large Text: 3:1 minimum
- UI Components: 3:1 minimum

### Focus States
- Outline: 2px solid var(--color-primary-500)
- Outline Offset: 2px
- Border Radius: inherit from component

### ARIA Labels
- Always include for icon-only buttons
- Use semantic HTML when possible
- Provide alt text for images

### Keyboard Navigation
- Tab order: logical flow
- Enter/Space: activate buttons
- Escape: close modals/dropdowns
- Arrow keys: navigate lists/menus

---

## Dark/Light Mode Toggle

```tsx
/* Implementation */
- Use CSS variables for all colors
- Toggle class on <html>: 'dark' / 'light'
- Persist preference in localStorage
- System preference detection: prefers-color-scheme

/* Transition */
transition: background-color 200ms, color 200ms, border-color 200ms;
```

---

## Mobile Considerations

### Touch Targets
- Minimum: 44px × 44px
- Recommended: 48px × 48px

### Typography
- Increase base font size on mobile: 16px (prevents zoom)
- Adjust line height for readability

### Spacing
- Increase padding/margin on mobile
- Reduce data density for better usability

### Navigation
- Bottom tab bar for primary navigation
- Hamburger menu for secondary items

---

## Performance

### Image Optimization
- Use Next.js Image component
- WebP format with fallback
- Lazy loading for below-fold images
- Responsive images (srcset)

### Font Loading
- Font Display: swap
- Preload critical fonts
- Subset fonts (Latin only if applicable)

### CSS
- Critical CSS inline
- Non-critical CSS deferred
- Minimize unused CSS (Tailwind purge)

---

## Examples

### Price Card Component

```tsx
<Card>
  <div className="flex items-center justify-between">
    <div>
      <Label>BTC/USD</Label>
      <div className="text-3xl font-mono font-semibold">
        $42,358.21
      </div>
    </div>
    <Badge variant="success">
      ▲ 2.34%
    </Badge>
  </div>
</Card>
```

### Data Table Row

```tsx
<tr className="border-b border-border-primary hover:bg-bg-tertiary transition-colors">
  <td className="py-3 px-4">BTC</td>
  <td className="py-3 px-4 text-right font-mono">$42,358.21</td>
  <td className="py-3 px-4 text-right font-mono text-success-500">
    +2.34%
  </td>
</tr>
```

---

## Resources

### Inspiration
- Binance
- Coinbase Pro
- Upbit
- TradingView
- Stripe Dashboard
- Vercel Dashboard

### Tools
- Figma (design mockups)
- Tailwind CSS Playground
- Coolors (color palettes)
- Type Scale (typography)

---

## Notes for AI Agents

1. **Always use design tokens**: Never hardcode colors/spacing
2. **Dark mode first**: Design for dark theme, then adapt to light
3. **Semantic HTML**: Use proper tags (button, nav, main, etc.)
4. **Consistent spacing**: Stick to 4px grid system
5. **Accessibility**: Always include ARIA labels, proper contrast
6. **Performance**: Lazy load images, optimize animations
7. **Mobile responsive**: Test on all breakpoints
8. **Brand consistency**: Use primary colors sparingly for emphasis
9. **Data clarity**: Use monospace font for numbers, tabular-nums
10. **Visual hierarchy**: Clear heading structure, proper spacing

---

## Version

- Version: 1.0.0
- Last Updated: 2025-01-24
- Next Review: When adding new components or major features
