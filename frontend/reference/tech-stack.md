# Tech Stack - Frontend

## Core Framework & Language

### Next.js 16 (App Router)
- **Purpose**: React framework with Server Components and optimized routing
- **Why**: SEO optimization, fast initial load, server/client component separation
- **Key Features**:
  - App Router (src/app directory structure)
  - Server Components (RSC)
  - Server Actions
  - Streaming SSR
  - Image & Font Optimization

### TypeScript 5
- **Purpose**: Type safety and enhanced developer experience
- **Configuration**: strict mode enabled
- **Usage**:
  - All components and utilities written in TypeScript
  - API response types must be defined
  - Leverage generic types actively

### React 19
- **Purpose**: UI library
- **Key Features**:
  - React Compiler (automatic optimization)
  - Actions & Form Actions
  - use() Hook
  - Optimistic Updates

---

## Styling & UI

### Tailwind CSS 4
- **Purpose**: Utility-first CSS framework
- **Why**: Fast development speed, consistent design, small bundle size
- **Configuration**:
  - Custom color palette (see design-system.md)
  - Custom spacing, typography
  - Dark mode support

### Framer Motion
- **Purpose**: Animation library
- **Use Cases**:
  - Page transition animations
  - Chart data change animations
  - Micro interactions
  - Scroll-based animations
- **Installation**: `npm install framer-motion`

### Radix UI
- **Purpose**: Headless UI component library
- **Why**: Perfect accessibility (A11Y) support, high customization freedom
- **Components**:
  - Dialog, Dropdown, Tooltip, Select
  - Popover, Toast, Context Menu
  - Accordion, Tabs, Slider
- **Installation**: `npm install @radix-ui/react-*` (per component)

---

## State Management

### Zustand
- **Purpose**: Lightweight global state management
- **Why**:
  - Simpler than Redux with less boilerplate
  - Excellent compatibility with React 19
  - DevTools support
- **Use Cases**:
  - User authentication state
  - Theme settings (light/dark mode)
  - Global UI state (sidebar, modals)
- **Installation**: `npm install zustand`

### TanStack Query (React Query) v5
- **Purpose**: Server state management and caching
- **Why**:
  - API data caching and synchronization
  - Automatic refetch, polling
  - Optimistic updates
  - Automatic loading/error state management
- **Use Cases**:
  - Blockchain data fetching
  - Real-time price updates
  - Transaction status queries
- **Installation**: `npm install @tanstack/react-query`

---

## Data Visualization

### TradingView Lightweight Charts
- **Purpose**: High-performance financial charts
- **Why**: Industry standard for crypto exchanges, lightweight and fast
- **Features**:
  - Candlestick, Line, Area, Bar charts
  - Real-time data streaming
  - Highly customizable
- **Installation**: `npm install lightweight-charts`

### Recharts
- **Purpose**: React-based chart library
- **Why**: React-friendly, declarative API
- **Use Cases**:
  - Dashboard statistics charts
  - Portfolio distribution
  - Simple data visualization
- **Installation**: `npm install recharts`

### D3.js (Optional)
- **Purpose**: Advanced data visualization
- **When to use**: When custom visualizations are needed
- **Installation**: `npm install d3`

---

## Form & Validation

### React Hook Form
- **Purpose**: Form state management and validation
- **Why**:
  - Minimizes re-renders (excellent performance)
  - Intuitive API
  - Perfect TypeScript support
- **Installation**: `npm install react-hook-form`

### Zod
- **Purpose**: TypeScript-first schema validation
- **Why**:
  - Type safety
  - API response validation
  - Form validation (integrates with React Hook Form)
- **Installation**: `npm install zod`

---

## API Communication

### Axios
- **Purpose**: HTTP client
- **Why**:
  - Interceptor support (auto-add auth tokens)
  - Request/response transformation
  - Timeout configuration
  - Error handling
- **Installation**: `npm install axios`
- **Alternative**: Native fetch with enhanced wrapper

### WebSocket (Socket.io Client)
- **Purpose**: Real-time bidirectional communication
- **Use Cases**:
  - Real-time price updates
  - Transaction notifications
  - Real-time chart data
- **Installation**: `npm install socket.io-client`

---

## Utilities & Tools

### date-fns
- **Purpose**: Date/time handling
- **Why**: Lightweight and modular, Tree-shakeable
- **Installation**: `npm install date-fns`

### numeral.js / numbro
- **Purpose**: Number formatting
- **Use Cases**: Price, quantity, percentage display
- **Installation**: `npm install numbro`

### clsx / classnames
- **Purpose**: Conditional className composition
- **Why**: Convenient when used with Tailwind
- **Installation**: `npm install clsx`

### tailwind-merge
- **Purpose**: Resolve Tailwind class conflicts
- **Installation**: `npm install tailwind-merge`

---

## Testing

### Vitest
- **Purpose**: Unit testing
- **Why**: Vite-based, fast, Jest compatible
- **Installation**: `npm install -D vitest @vitejs/plugin-react`

### React Testing Library
- **Purpose**: Component testing
- **Installation**: `npm install -D @testing-library/react @testing-library/jest-dom`

### Playwright (Optional)
- **Purpose**: E2E testing
- **Installation**: `npm install -D @playwright/test`

---

## Code Quality & Linting

### ESLint 9
- **Config**: Next.js default config + custom rules
- **Plugins**:
  - `eslint-plugin-react-hooks`
  - `eslint-plugin-jsx-a11y` (accessibility)

### Prettier
- **Purpose**: Code formatting
- **Installation**: `npm install -D prettier prettier-plugin-tailwindcss`
- **Config**: Includes Tailwind class sorting

### TypeScript ESLint
- **Purpose**: TypeScript-specific lint rules
- **Installation**: `npm install -D @typescript-eslint/parser @typescript-eslint/eslint-plugin`

---

## Performance & Monitoring

### React DevTools
- **Purpose**: Component tree debugging

### TanStack Query DevTools
- **Purpose**: Query state monitoring
- **Installation**: `@tanstack/react-query-devtools` (dev only)

### Lighthouse
- **Purpose**: Performance, accessibility, SEO audit

### Web Vitals
- **Purpose**: Core Web Vitals measurement
- **Installation**: `npm install web-vitals`

---

## Deployment & Build

### Vercel (Recommended)
- **Why**: Optimal Next.js hosting, automatic deployment, Edge Functions

### Docker (Alternative)
- **For**: Self-hosting

---

## Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── (auth)/            # Auth route group
│   ├── (dashboard)/       # Dashboard route group
│   ├── layout.tsx         # Root layout
│   └── page.tsx           # Home page
├── components/            # React components
│   ├── ui/               # UI primitives (Radix + Tailwind)
│   ├── charts/           # Chart components
│   ├── forms/            # Form components
│   └── layouts/          # Layout components
├── lib/                  # Utilities & configs
│   ├── api/             # API clients
│   ├── hooks/           # Custom hooks
│   ├── utils/           # Helper functions
│   └── stores/          # Zustand stores
├── types/               # TypeScript types
└── styles/              # Global styles
```

---

## Environment Variables

```env
# API
NEXT_PUBLIC_API_URL=
NEXT_PUBLIC_WS_URL=

# Feature Flags
NEXT_PUBLIC_ENABLE_TESTNET=

# Analytics (Optional)
NEXT_PUBLIC_GA_ID=
```

---

## Installation Commands

### Essential Packages
```bash
# UI & Styling
npm install framer-motion clsx tailwind-merge
npm install @radix-ui/react-dialog @radix-ui/react-dropdown-menu @radix-ui/react-tooltip

# State Management
npm install zustand @tanstack/react-query

# Forms & Validation
npm install react-hook-form zod @hookform/resolvers

# API & WebSocket
npm install axios socket.io-client

# Charts
npm install lightweight-charts recharts

# Utilities
npm install date-fns numbro

# Dev Tools
npm install -D @tanstack/react-query-devtools
npm install -D prettier prettier-plugin-tailwindcss
```

### Optional Packages
```bash
# Testing
npm install -D vitest @vitejs/plugin-react @testing-library/react

# Advanced visualization
npm install d3

# Internationalization
npm install next-intl
```

---

## Development Workflow

1. **Local Development**: `npm run dev` (http://localhost:3000)
2. **Type Checking**: `tsc --noEmit`
3. **Linting**: `npm run lint`
4. **Build**: `npm run build`
5. **Production**: `npm run start`

---

## Performance Goals

- **LCP**: < 2.5s
- **FID**: < 100ms
- **CLS**: < 0.1
- **Bundle Size**: < 200KB (initial load)
- **Lighthouse Score**: 90+ (all categories)

---

## Browser Support

- Chrome/Edge (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)
- Mobile browsers (iOS Safari, Chrome Mobile)

---

## Notes for AI Agents

1. **Always use TypeScript**: All new files should use .tsx/.ts extension
2. **Server vs Client**: When 'use client' is needed:
   - Using Hooks (useState, useEffect, etc.)
   - Event handlers
   - Browser APIs
   - Third-party libraries (most)
3. **Performance**: Consider React.memo, useMemo, useCallback for large components
4. **Accessibility**: Always consider ARIA labels, keyboard navigation
5. **Naming Convention**:
   - Components: PascalCase
   - Files: kebab-case or PascalCase
   - Functions: camelCase
   - Constants: UPPER_SNAKE_CASE
