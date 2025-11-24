# API Contracts - Frontend

> This document defines all API endpoints, WebSocket events, and data contracts between frontend and backend.

## Table of Contents

1. [Base Configuration](#base-configuration)
2. [Authentication](#authentication)
3. [REST API Endpoints](#rest-api-endpoints)
4. [WebSocket Events](#websocket-events)
5. [Error Handling](#error-handling)
6. [Type Definitions](#type-definitions)

---

## Base Configuration

### Environment Variables

```env
NEXT_PUBLIC_API_URL=https://api.monadflow.com
NEXT_PUBLIC_WS_URL=wss://ws.monadflow.com
NEXT_PUBLIC_API_VERSION=v1
```

### Base URLs

- **REST API**: `${NEXT_PUBLIC_API_URL}/${NEXT_PUBLIC_API_VERSION}`
- **WebSocket**: `${NEXT_PUBLIC_WS_URL}`

### Request Headers

```typescript
{
  'Content-Type': 'application/json',
  'Authorization': 'Bearer {token}',  // Optional, for authenticated requests
  'X-Client-Version': '1.0.0'
}
```

---

## Authentication

### POST /auth/login

**Description**: User login

**Request**:
```typescript
{
  email: string
  password: string
}
```

**Response**:
```typescript
{
  success: boolean
  data: {
    user: {
      id: string
      email: string
      name: string
      avatar?: string
    }
    accessToken: string
    refreshToken: string
    expiresIn: number  // seconds
  }
}
```

### POST /auth/refresh

**Description**: Refresh access token

**Request**:
```typescript
{
  refreshToken: string
}
```

**Response**:
```typescript
{
  success: boolean
  data: {
    accessToken: string
    expiresIn: number
  }
}
```

---

## REST API Endpoints

### Blockchain Data

#### GET /blockchain/stats

**Description**: Get overall blockchain statistics

**Query Parameters**:
```typescript
{
  timeRange?: '1h' | '24h' | '7d' | '30d'  // default: '24h'
}
```

**Response**:
```typescript
{
  success: boolean
  data: {
    blockHeight: number
    tps: number  // transactions per second
    averageBlockTime: number  // milliseconds
    totalTransactions: number
    activeAddresses: number
    gasPrice: {
      low: string
      medium: string
      high: string
    }
    timestamp: string  // ISO 8601
  }
}
```

#### GET /blockchain/blocks

**Description**: Get list of recent blocks

**Query Parameters**:
```typescript
{
  limit?: number  // default: 20, max: 100
  offset?: number  // default: 0
  sortBy?: 'height' | 'timestamp'  // default: 'height'
  order?: 'asc' | 'desc'  // default: 'desc'
}
```

**Response**:
```typescript
{
  success: boolean
  data: {
    blocks: Array<{
      height: number
      hash: string
      timestamp: string
      transactionCount: number
      size: number  // bytes
      gasUsed: string
      gasLimit: string
      validator?: string
    }>
    pagination: {
      total: number
      limit: number
      offset: number
      hasMore: boolean
    }
  }
}
```

#### GET /blockchain/blocks/:blockId

**Description**: Get detailed block information

**Path Parameters**:
- `blockId`: Block height (number) or block hash (string)

**Response**:
```typescript
{
  success: boolean
  data: {
    height: number
    hash: string
    parentHash: string
    timestamp: string
    transactionCount: number
    transactions: string[]  // transaction hashes
    size: number
    gasUsed: string
    gasLimit: string
    validator?: string
    extraData?: string
  }
}
```

### Transactions

#### GET /transactions

**Description**: Get list of transactions

**Query Parameters**:
```typescript
{
  limit?: number
  offset?: number
  address?: string  // filter by address (from or to)
  status?: 'pending' | 'confirmed' | 'failed'
  type?: 'transfer' | 'contract' | 'stake'
}
```

**Response**:
```typescript
{
  success: boolean
  data: {
    transactions: Array<{
      hash: string
      from: string
      to: string
      value: string  // in wei
      gasUsed: string
      gasPrice: string
      timestamp: string
      status: 'pending' | 'confirmed' | 'failed'
      blockHeight?: number
    }>
    pagination: {
      total: number
      limit: number
      offset: number
      hasMore: boolean
    }
  }
}
```

#### GET /transactions/:txHash

**Description**: Get detailed transaction information

**Path Parameters**:
- `txHash`: Transaction hash

**Response**:
```typescript
{
  success: boolean
  data: {
    hash: string
    from: string
    to: string
    value: string
    gasUsed: string
    gasPrice: string
    nonce: number
    timestamp: string
    status: 'pending' | 'confirmed' | 'failed'
    blockHeight?: number
    blockHash?: string
    input?: string
    logs?: Array<{
      address: string
      topics: string[]
      data: string
    }>
  }
}
```

### Addresses

#### GET /addresses/:address

**Description**: Get address information

**Path Parameters**:
- `address`: Wallet address

**Response**:
```typescript
{
  success: boolean
  data: {
    address: string
    balance: string  // in wei
    transactionCount: number
    firstSeen: string
    lastActive: string
    type: 'wallet' | 'contract'
    contractInfo?: {
      name?: string
      verified: boolean
      abi?: any
    }
  }
}
```

#### GET /addresses/:address/transactions

**Description**: Get transactions for specific address

**Query Parameters**: Same as GET /transactions

**Response**: Same as GET /transactions

### Network

#### GET /network/status

**Description**: Get current network status

**Response**:
```typescript
{
  success: boolean
  data: {
    status: 'healthy' | 'degraded' | 'down'
    latency: number  // milliseconds
    peers: number
    syncStatus: {
      isSyncing: boolean
      currentBlock: number
      highestBlock: number
      percentage: number
    }
  }
}
```

#### GET /network/gas-price

**Description**: Get current gas price recommendations

**Response**:
```typescript
{
  success: boolean
  data: {
    timestamp: string
    prices: {
      slow: {
        price: string  // in gwei
        estimatedTime: number  // seconds
      }
      standard: {
        price: string
        estimatedTime: number
      }
      fast: {
        price: string
        estimatedTime: number
      }
      instant: {
        price: string
        estimatedTime: number
      }
    }
  }
}
```

---

## WebSocket Events

### Connection

**URL**: `${NEXT_PUBLIC_WS_URL}`

**Connection Parameters**:
```typescript
{
  token?: string  // Optional authentication token
}
```

### Client → Server Events

#### subscribe

Subscribe to specific data streams

```typescript
{
  event: 'subscribe'
  data: {
    channels: Array<'blocks' | 'transactions' | 'gas_price'>
    filters?: {
      address?: string
      minValue?: string
    }
  }
}
```

#### unsubscribe

Unsubscribe from data streams

```typescript
{
  event: 'unsubscribe'
  data: {
    channels: Array<'blocks' | 'transactions' | 'gas_price'>
  }
}
```

#### ping

Heartbeat to keep connection alive

```typescript
{
  event: 'ping'
  timestamp: number
}
```

### Server → Client Events

#### connected

Sent when connection is established

```typescript
{
  event: 'connected'
  data: {
    connectionId: string
    timestamp: string
  }
}
```

#### pong

Response to ping

```typescript
{
  event: 'pong'
  timestamp: number
}
```

#### new_block

New block mined

```typescript
{
  event: 'new_block'
  data: {
    height: number
    hash: string
    timestamp: string
    transactionCount: number
    gasUsed: string
  }
}
```

#### new_transaction

New transaction detected

```typescript
{
  event: 'new_transaction'
  data: {
    hash: string
    from: string
    to: string
    value: string
    timestamp: string
  }
}
```

#### gas_price_update

Gas price update

```typescript
{
  event: 'gas_price_update'
  data: {
    timestamp: string
    prices: {
      slow: string
      standard: string
      fast: string
      instant: string
    }
  }
}
```

#### error

Error message

```typescript
{
  event: 'error'
  data: {
    code: string
    message: string
  }
}
```

---

## Error Handling

### Error Response Format

All error responses follow this format:

```typescript
{
  success: false
  error: {
    code: string
    message: string
    details?: any
    timestamp: string
  }
}
```

### HTTP Status Codes

- `200`: Success
- `201`: Created
- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Not Found
- `429`: Too Many Requests
- `500`: Internal Server Error
- `503`: Service Unavailable

### Error Codes

#### Authentication Errors (AUTH_*)

- `AUTH_INVALID_CREDENTIALS`: Invalid email or password
- `AUTH_TOKEN_EXPIRED`: Access token has expired
- `AUTH_TOKEN_INVALID`: Invalid or malformed token
- `AUTH_UNAUTHORIZED`: User not authorized for this resource

#### Validation Errors (VALIDATION_*)

- `VALIDATION_MISSING_FIELD`: Required field is missing
- `VALIDATION_INVALID_FORMAT`: Field format is invalid
- `VALIDATION_OUT_OF_RANGE`: Value is out of acceptable range

#### Resource Errors (RESOURCE_*)

- `RESOURCE_NOT_FOUND`: Requested resource not found
- `RESOURCE_ALREADY_EXISTS`: Resource already exists

#### Network Errors (NETWORK_*)

- `NETWORK_UNAVAILABLE`: Blockchain network is unavailable
- `NETWORK_TIMEOUT`: Request to blockchain network timed out

#### Rate Limiting (RATE_*)

- `RATE_LIMIT_EXCEEDED`: Too many requests, try again later

---

## Type Definitions

### Common Types

```typescript
// Pagination
interface Pagination {
  total: number
  limit: number
  offset: number
  hasMore: boolean
}

// API Response Wrapper
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
    details?: any
    timestamp: string
  }
}

// Time Range
type TimeRange = '1h' | '24h' | '7d' | '30d' | '1y' | 'all'

// Transaction Status
type TransactionStatus = 'pending' | 'confirmed' | 'failed'

// Network Status
type NetworkStatus = 'healthy' | 'degraded' | 'down'
```

### TypeScript Example

```typescript
// Example: Using the API with TypeScript

import axios from 'axios'

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Example function
async function getBlockchainStats(): Promise<ApiResponse<BlockchainStats>> {
  const response = await api.get('/blockchain/stats')
  return response.data
}
```

---

## Notes for Implementation

1. **Error Handling**: Always wrap API calls in try-catch blocks
2. **Loading States**: Show loading indicators during API calls
3. **Retry Logic**: Implement exponential backoff for failed requests
4. **Caching**: Use TanStack Query for automatic caching and refetching
5. **WebSocket Reconnection**: Implement automatic reconnection with exponential backoff
6. **Rate Limiting**: Handle 429 errors gracefully with retry after delay
7. **Token Refresh**: Automatically refresh tokens when they expire
8. **Optimistic Updates**: Use optimistic updates for better UX (with TanStack Query)

---

## Version

- Version: 1.0.0
- Last Updated: 2025-01-24
- API Compatibility: v1
