const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9000'

export interface Market {
  id: string
  question: string
  description: string
  category: string
  creator: string
  yesPrice: number
  noPrice: number
  yesPool: number
  noPool: number
  totalVolume: number
  expirationDate: string
  status: string
  tags: string[]
}

export interface Bet {
  id: string
  marketId: string
  userId: string
  side: string
  amount: number
  shares: number
  price: number
  timestamp: string
}

export interface FeedEvent {
  id: string
  userId: string
  username: string
  type: string
  marketId: string
  marketQuestion: string
  side?: string
  amount?: number
  timestamp: string
}

export interface User {
  id: string
  username: string
  bio: string
  totalVolume: number
  totalBets: number
  winRate: number
  followers: number
  following: number
}

export interface LeaderboardEntry {
  userId: string
  username: string
  totalVolume: number
  totalBets: number
  winRate: number
}

class APIClient {
  private baseURL: string

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      })

      if (!response.ok) {
        throw new Error(`API error: ${response.statusText}`)
      }

      return response.json()
    } catch (error) {
      console.error('API request failed:', error)
      throw error
    }
  }

  // Market endpoints
  async getMarkets(
    category?: string,
    search?: string
  ): Promise<{ markets: Market[] }> {
    const params = new URLSearchParams()
    if (category) params.append('category', category)
    if (search) params.append('search', search)

    const query = params.toString()
    return this.request(`/api/markets${query ? `?${query}` : ''}`)
  }

  async getMarket(marketId: string): Promise<Market> {
    return this.request(`/api/markets/${marketId}`)
  }

  async getMarketDetail(
    marketId: string,
    userId: string
  ): Promise<{
    market: Market
    userPosition?: any
    aiSummary?: string
    recentBets: Bet[]
  }> {
    return this.request(`/api/market-detail/${marketId}?user_id=${userId}`)
  }

  // Betting endpoints
  async placeBet(
    userId: string,
    marketId: string,
    side: string,
    amount: number
  ): Promise<Bet> {
    return this.request(`/api/bets`, {
      method: 'POST',
      body: JSON.stringify({ userId, marketId, side, amount }),
    })
  }

  async getUserBets(userId: string): Promise<{ bets: Bet[] }> {
    return this.request(`/api/users/${userId}/bets`)
  }

  // Social feed endpoints
  async getFeed(limit?: number): Promise<{ events: FeedEvent[] }> {
    const query = limit ? `?limit=${limit}` : ''
    return this.request(`/api/feed${query}`)
  }

  async getUserProfile(userId: string): Promise<User> {
    return this.request(`/api/users/${userId}`)
  }

  async followUser(userId: string, targetUserId: string): Promise<void> {
    return this.request(`/api/users/${userId}/follow`, {
      method: 'POST',
      body: JSON.stringify({ targetUserId }),
    })
  }

  async unfollowUser(userId: string, targetUserId: string): Promise<void> {
    return this.request(`/api/users/${userId}/unfollow`, {
      method: 'POST',
      body: JSON.stringify({ targetUserId }),
    })
  }

  async getLeaderboard(
    sortBy: 'volume' | 'winrate' = 'volume',
    limit: number = 10
  ): Promise<{ leaderboard: LeaderboardEntry[] }> {
    return this.request(`/api/leaderboard?sort_by=${sortBy}&limit=${limit}`)
  }

  // AI endpoints
  async getRecommendations(
    userId: string,
    limit: number = 5
  ): Promise<{ recommendations: Market[] }> {
    return this.request(`/api/recommendations/${userId}?limit=${limit}`)
  }

  async getHomepage(userId: string): Promise<{
    recommendations: Market[]
    feed: FeedEvent[]
    leaderboard: LeaderboardEntry[]
  }> {
    return this.request(`/api/homepage/${userId}`)
  }

  // Search endpoint
  async searchMarkets(query: string): Promise<{ markets: Market[] }> {
    return this.request(`/api/markets/search?q=${encodeURIComponent(query)}`)
  }
}

export const api = new APIClient(API_BASE_URL)
