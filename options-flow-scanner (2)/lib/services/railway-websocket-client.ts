/**
 * ================================================
 * RAILWAY WEBSOCKET CLIENT - PRODUCTION READY
 * ================================================
 * Connects Next.js frontend to Railway WebSocket server
 * Handles automatic reconnection and subscriptions
 * Zero-latency real-time data streaming
 * ================================================
 */

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface WebSocketMessage {
  type: string
  channel?: string
  data: any
  timestamp: number
  symbols?: string[]
  metadata?: Record<string, any>
}

export interface Subscription {
  id: string
  channel: string
  symbols: string[]
  callback: (data: any) => void
}

export interface ConnectionConfig {
  url: string
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  debug?: boolean
}

/**
 * Professional WebSocket Client for Railway Backend
 */
export class RailwayWebSocketClient {
  private ws: WebSocket | null = null
  private status: WebSocketStatus = 'disconnected'
  private subscriptions: Map<string, Subscription> = new Map()
  private reconnectAttempts = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private heartbeatTimer: NodeJS.Timeout | null = null
  private messageQueue: WebSocketMessage[] = []
  private statusCallbacks: Set<(status: WebSocketStatus) => void> = new Set()
  private pendingSubscriptions: Set<string> = new Set()
  
  private config: Required<ConnectionConfig> = {
    url: '',
    reconnectInterval: 3000,
    maxReconnectAttempts: 10,
    heartbeatInterval: 30000,
    debug: false,
  }

  constructor(config: ConnectionConfig) {
    this.config = { ...this.config, ...config }
    
    if (this.config.debug) {
      console.log('🔧 Railway WebSocket Client initialized:', this.config.url)
    }
  }

  /**
   * Connect to Railway WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      if (this.config.debug) {
        console.log('✅ WebSocket already connected')
      }
      return
    }

    this.updateStatus('connecting')

    try {
      this.ws = new WebSocket(this.config.url)
      this.setupEventHandlers()
    } catch (error) {
      console.error('❌ Failed to create WebSocket:', error)
      this.updateStatus('error')
      this.scheduleReconnect()
    }
  }

  /**
   * Setup WebSocket event handlers
   */
  private setupEventHandlers(): void {
    if (!this.ws) return

    this.ws.onopen = () => {
      if (this.config.debug) {
        console.log('✅ Connected to Railway WebSocket server')
      }
      
      this.updateStatus('connected')
      this.reconnectAttempts = 0
      this.startHeartbeat()
      this.resubscribeAll()
      this.flushMessageQueue()
    }

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('❌ Failed to parse WebSocket message:', error)
      }
    }

    this.ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error)
      this.updateStatus('error')
    }

    this.ws.onclose = (event) => {
      if (this.config.debug) {
        console.log('🔌 WebSocket disconnected:', event.code, event.reason)
      }
      
      this.updateStatus('disconnected')
      this.stopHeartbeat()
      this.scheduleReconnect()
    }
  }

  /**
   * Handle incoming message from server
   */
  private handleMessage(message: WebSocketMessage): void {
    if (this.config.debug) {
      console.log('📨 Received:', message.type, message.channel)
    }

    // Handle system messages
    if (message.type === 'pong') {
      // Heartbeat response
      return
    }

    if (message.type === 'subscribed') {
      this.pendingSubscriptions.delete(message.channel || '')
      if (this.config.debug) {
        console.log('✅ Subscription confirmed:', message.channel)
      }
      return
    }

    if (message.type === 'unsubscribed') {
      if (this.config.debug) {
        console.log('✅ Unsubscription confirmed:', message.channel)
      }
      return
    }

    // Route message to appropriate subscriptions
    this.subscriptions.forEach(sub => {
      if (this.messageMatchesSubscription(message, sub)) {
        try {
          sub.callback(message.data)
        } catch (error) {
          console.error('❌ Subscription callback error:', error)
        }
      }
    })
  }

  /**
   * Check if message matches subscription
   */
  private messageMatchesSubscription(message: WebSocketMessage, sub: Subscription): boolean {
    // Check if message channel matches subscription channel
    if (message.channel !== sub.channel) {
      return false
    }

    // If subscription has specific symbols, check if message matches
    if (sub.symbols.length > 0 && message.symbols) {
      const hasMatch = sub.symbols.some(symbol => 
        message.symbols?.includes(symbol)
      )
      if (!hasMatch) {
        return false
      }
    }

    return true
  }

  /**
   * Subscribe to a channel
   */
  subscribe(channel: string, symbols: string[], callback: (data: any) => void): string {
    const id = `${channel}-${Date.now()}-${Math.random()}`
    
    const subscription: Subscription = {
      id,
      channel,
      symbols,
      callback,
    }

    this.subscriptions.set(id, subscription)
    this.pendingSubscriptions.add(channel)

    // Send subscription message if connected
    if (this.status === 'connected') {
      this.sendSubscription(subscription)
    }

    if (this.config.debug) {
      console.log('📥 Subscribed:', channel, symbols)
    }

    return id
  }

  /**
   * Unsubscribe from a channel
   */
  unsubscribe(subscriptionId: string): void {
    const subscription = this.subscriptions.get(subscriptionId)
    if (!subscription) return

    // Send unsubscription message if connected
    if (this.status === 'connected') {
      this.sendUnsubscription(subscription)
    }

    this.subscriptions.delete(subscriptionId)

    if (this.config.debug) {
      console.log('📤 Unsubscribed:', subscription.channel)
    }
  }

  /**
   * Send subscription message to server
   */
  private sendSubscription(subscription: Subscription): void {
    this.send({
      type: 'subscribe',
      channel: subscription.channel,
      symbols: subscription.symbols,
      data: {},
      timestamp: Date.now(),
    })
  }

  /**
   * Send unsubscription message to server
   */
  private sendUnsubscription(subscription: Subscription): void {
    this.send({
      type: 'unsubscribe',
      channel: subscription.channel,
      data: {},
      timestamp: Date.now(),
    })
  }

  /**
   * Resubscribe to all channels after reconnection
   */
  private resubscribeAll(): void {
    if (this.config.debug) {
      console.log('🔄 Resubscribing to', this.subscriptions.size, 'channels')
    }

    this.subscriptions.forEach(sub => {
      this.sendSubscription(sub)
    })
  }

  /**
   * Send message through WebSocket
   */
  private send(message: WebSocketMessage): void {
    if (this.status !== 'connected' || !this.ws) {
      // Queue message for later
      this.messageQueue.push(message)
      return
    }

    try {
      this.ws.send(JSON.stringify(message))
    } catch (error) {
      console.error('❌ Failed to send WebSocket message:', error)
      this.messageQueue.push(message)
    }
  }

  /**
   * Flush queued messages
   */
  private flushMessageQueue(): void {
    if (this.config.debug && this.messageQueue.length > 0) {
      console.log('📬 Flushing', this.messageQueue.length, 'queued messages')
    }

    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift()
      if (message) {
        this.send(message)
      }
    }
  }

  /**
   * Start heartbeat to keep connection alive
   */
  private startHeartbeat(): void {
    this.stopHeartbeat()
    
    this.heartbeatTimer = setInterval(() => {
      this.send({
        type: 'ping',
        data: {},
        timestamp: Date.now(),
      })
    }, this.config.heartbeatInterval)
  }

  /**
   * Stop heartbeat
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  /**
   * Schedule reconnection attempt
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.error('❌ Max reconnection attempts reached')
      return
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    this.reconnectAttempts++
    const delay = Math.min(
      this.config.reconnectInterval * Math.pow(1.5, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    )

    if (this.config.debug) {
      console.log(`🔄 Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`)
    }

    this.reconnectTimer = setTimeout(() => {
      this.connect()
    }, delay)
  }

  /**
   * Update connection status
   */
  private updateStatus(status: WebSocketStatus): void {
    this.status = status
    this.statusCallbacks.forEach(callback => {
      try {
        callback(status)
      } catch (error) {
        console.error('❌ Status callback error:', error)
      }
    })
  }

  /**
   * Register status change callback
   */
  onStatusChange(callback: (status: WebSocketStatus) => void): () => void {
    this.statusCallbacks.add(callback)
    
    // Return unsubscribe function
    return () => {
      this.statusCallbacks.delete(callback)
    }
  }

  /**
   * Get current connection status
   */
  getStatus(): WebSocketStatus {
    return this.status
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.status === 'connected'
  }

  /**
   * Disconnect WebSocket
   */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    this.stopHeartbeat()

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }

    this.subscriptions.clear()
    this.messageQueue = []
    this.updateStatus('disconnected')

    if (this.config.debug) {
      console.log('👋 Disconnected from Railway WebSocket server')
    }
  }

  /**
   * Get active subscriptions
   */
  getSubscriptions(): Subscription[] {
    return Array.from(this.subscriptions.values())
  }

  /**
   * Get subscription count
   */
  getSubscriptionCount(): number {
    return this.subscriptions.size
  }
}

/**
 * Create and export singleton instance
 */
let railwayWSClient: RailwayWebSocketClient | null = null

export function getRailwayWebSocketClient(): RailwayWebSocketClient {
  if (!railwayWSClient) {
    // Get WebSocket URL from environment
    const wsUrl = process.env.NEXT_PUBLIC_RAILWAY_WS_URL || 'ws://localhost:8080/ws'
    const debug = process.env.NODE_ENV === 'development'

    railwayWSClient = new RailwayWebSocketClient({
      url: wsUrl,
      debug,
    })

    // Auto-connect
    railwayWSClient.connect()
  }

  return railwayWSClient
}

/**
 * Export convenience functions
 */
export function subscribeToChannel(
  channel: string,
  symbols: string[],
  callback: (data: any) => void
): string {
  const client = getRailwayWebSocketClient()
  return client.subscribe(channel, symbols, callback)
}

export function unsubscribeFromChannel(subscriptionId: string): void {
  const client = getRailwayWebSocketClient()
  client.unsubscribe(subscriptionId)
}

export function getConnectionStatus(): WebSocketStatus {
  const client = getRailwayWebSocketClient()
  return client.getStatus()
}

export function isWebSocketConnected(): boolean {
  const client = getRailwayWebSocketClient()
  return client.isConnected()
}
