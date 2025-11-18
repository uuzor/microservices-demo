"use client"

import { useState, useEffect } from "react"
import { motion } from "framer-motion"
import {
  TrendingUp,
  TrendingDown,
  Clock,
  DollarSign,
  Users,
  Sparkles,
  ArrowLeft,
} from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { api, Market, Bet } from "@/lib/api"
import { formatCurrency, formatPercentage, formatDate } from "@/lib/utils"
import Link from "next/link"
import { useParams } from "next/navigation"

export default function MarketDetailPage() {
  const params = useParams()
  const marketId = params.id as string

  const [market, setMarket] = useState<Market | null>(null)
  const [aiSummary, setAiSummary] = useState<string>("")
  const [recentBets, setRecentBets] = useState<Bet[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [selectedSide, setSelectedSide] = useState<"YES" | "NO">("YES")
  const [betAmount, setBetAmount] = useState("")
  const [isPlacingBet, setIsPlacingBet] = useState(false)

  useEffect(() => {
    // Mock user ID
    const userId = "user123"

    api
      .getMarketDetail(marketId, userId)
      .then((data) => {
        setMarket(data.market)
        setAiSummary(data.aiSummary || "")
        setRecentBets(data.recentBets || [])
        setIsLoading(false)
      })
      .catch((error) => {
        console.error("Failed to load market:", error)
        setIsLoading(false)
      })
  }, [marketId])

  const handlePlaceBet = async () => {
    if (!betAmount || !market) return

    setIsPlacingBet(true)

    try {
      await api.placeBet("user123", market.id, selectedSide, parseFloat(betAmount))
      // Refresh market data
      const data = await api.getMarketDetail(marketId, "user123")
      setMarket(data.market)
      setRecentBets(data.recentBets || [])
      setBetAmount("")
      setIsPlacingBet(false)
    } catch (error) {
      console.error("Failed to place bet:", error)
      setIsPlacingBet(false)
    }
  }

  const calculateShares = () => {
    if (!betAmount || !market) return 0
    const amount = parseFloat(betAmount)
    const price = selectedSide === "YES" ? market.yesPrice : market.noPrice
    return amount / price
  }

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto space-y-6">
        <div className="skeleton h-12 w-48 rounded-lg" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            <div className="skeleton h-96 rounded-2xl" />
            <div className="skeleton h-64 rounded-2xl" />
          </div>
          <div className="skeleton h-[600px] rounded-2xl" />
        </div>
      </div>
    )
  }

  if (!market) {
    return (
      <div className="max-w-7xl mx-auto text-center py-16">
        <h2 className="text-2xl font-display font-bold text-text-primary mb-4">
          Market not found
        </h2>
        <Link href="/markets">
          <Button variant="primary">Back to Markets</Button>
        </Link>
      </div>
    )
  }

  const priceChange = market.yesPrice - 0.5
  const isIncreasing = priceChange > 0

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Back Button */}
      <Link href="/markets">
        <Button variant="outline" size="sm" className="gap-2">
          <ArrowLeft className="w-4 h-4" />
          Back to Markets
        </Button>
      </Link>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column */}
        <div className="lg:col-span-2 space-y-6">
          {/* Market Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6 }}
          >
            <Card>
              <CardContent className="p-8">
                <div className="flex items-start gap-4 mb-6">
                  <Badge variant="primary" className="text-base px-4 py-2">
                    {market.category}
                  </Badge>
                  <div className="flex items-center gap-2 text-text-tertiary">
                    <Clock className="w-4 h-4" />
                    <span className="text-sm">
                      Closes {formatDate(market.expirationDate)}
                    </span>
                  </div>
                </div>

                <h1 className="text-4xl font-display font-bold text-text-primary mb-4 leading-tight">
                  {market.question}
                </h1>

                <p className="text-lg text-text-secondary mb-8 leading-relaxed">
                  {market.description}
                </p>

                {/* Price Display */}
                <div className="grid grid-cols-2 gap-6 mb-6">
                  <div className="p-6 rounded-xl bg-gradient-to-br from-primary-50 to-accent-pink-50 border-2 border-primary-200">
                    <div className="text-sm text-text-tertiary mb-2">YES</div>
                    <div className="text-5xl font-bold font-display text-primary-600 tabular-nums mb-2">
                      {formatPercentage(market.yesPrice)}
                    </div>
                    <div className="text-sm text-text-secondary">
                      {formatCurrency(market.yesPool)} pool
                    </div>
                  </div>
                  <div className="p-6 rounded-xl bg-surface-secondary border-2 border-surface-tertiary">
                    <div className="text-sm text-text-tertiary mb-2">NO</div>
                    <div className="text-5xl font-bold font-display text-text-primary tabular-nums mb-2">
                      {formatPercentage(market.noPrice)}
                    </div>
                    <div className="text-sm text-text-secondary">
                      {formatCurrency(market.noPool)} pool
                    </div>
                  </div>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-3 gap-4 pt-6 border-t border-surface-tertiary">
                  <div>
                    <div className="flex items-center gap-2 text-text-tertiary mb-1">
                      <DollarSign className="w-4 h-4" />
                      <span className="text-sm">Total Volume</span>
                    </div>
                    <div className="text-2xl font-bold font-display text-text-primary tabular-nums">
                      {formatCurrency(market.totalVolume)}
                    </div>
                  </div>
                  <div>
                    <div className="flex items-center gap-2 text-text-tertiary mb-1">
                      {isIncreasing ? (
                        <TrendingUp className="w-4 h-4 text-success-dark" />
                      ) : (
                        <TrendingDown className="w-4 h-4 text-error-dark" />
                      )}
                      <span className="text-sm">24h Change</span>
                    </div>
                    <div
                      className={`text-2xl font-bold font-display tabular-nums ${
                        isIncreasing ? "text-success-dark" : "text-error-dark"
                      }`}
                    >
                      {isIncreasing ? "+" : ""}
                      {(priceChange * 100).toFixed(1)}%
                    </div>
                  </div>
                  <div>
                    <div className="flex items-center gap-2 text-text-tertiary mb-1">
                      <Users className="w-4 h-4" />
                      <span className="text-sm">Traders</span>
                    </div>
                    <div className="text-2xl font-bold font-display text-text-primary tabular-nums">
                      {Math.floor(market.totalVolume / 50)}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* AI Summary */}
          {aiSummary && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.2 }}
            >
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Sparkles className="w-5 h-5 text-accent-yellow-500" />
                    AI Market Summary
                  </CardTitle>
                  <CardDescription>
                    Powered by Google Gemini
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="prose prose-sm max-w-none">
                    <p className="text-text-secondary leading-relaxed">
                      {aiSummary}
                    </p>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {/* Recent Bets */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
          >
            <Card>
              <CardHeader>
                <CardTitle>Recent Activity</CardTitle>
                <CardDescription>Latest bets on this market</CardDescription>
              </CardHeader>
              <CardContent>
                {recentBets.length === 0 ? (
                  <div className="text-center py-8 text-text-tertiary">
                    No bets yet. Be the first to trade!
                  </div>
                ) : (
                  <div className="space-y-3">
                    {recentBets.slice(0, 10).map((bet) => (
                      <div
                        key={bet.id}
                        className="flex items-center justify-between p-4 rounded-lg bg-surface-secondary hover:bg-surface-tertiary transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary-500 to-accent-pink-500 flex items-center justify-center text-white font-bold text-sm">
                            {bet.userId.charAt(0).toUpperCase()}
                          </div>
                          <div>
                            <div className="text-sm font-semibold text-text-primary">
                              {bet.userId}
                            </div>
                            <div className="text-xs text-text-tertiary">
                              {new Date(bet.timestamp).toLocaleString()}
                            </div>
                          </div>
                        </div>
                        <div className="text-right">
                          <div
                            className={`text-sm font-bold ${
                              bet.side === "YES"
                                ? "text-primary-600"
                                : "text-text-primary"
                            }`}
                          >
                            {bet.side}
                          </div>
                          <div className="text-sm text-text-secondary tabular-nums">
                            {formatCurrency(bet.amount)}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        </div>

        {/* Right Column - Betting Panel */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <Card className="sticky top-24">
            <CardHeader>
              <CardTitle>Place Your Bet</CardTitle>
              <CardDescription>
                Choose your position and amount
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Side Selection */}
              <div className="grid grid-cols-2 gap-3">
                <button
                  onClick={() => setSelectedSide("YES")}
                  className={`p-6 rounded-xl border-2 transition-all duration-200 ${
                    selectedSide === "YES"
                      ? "border-primary-500 bg-primary-50"
                      : "border-surface-tertiary bg-surface-secondary hover:bg-surface-tertiary"
                  }`}
                >
                  <div className="text-sm text-text-tertiary mb-2">YES</div>
                  <div className="text-3xl font-bold font-display text-primary-600 tabular-nums">
                    {formatPercentage(market.yesPrice)}
                  </div>
                </button>
                <button
                  onClick={() => setSelectedSide("NO")}
                  className={`p-6 rounded-xl border-2 transition-all duration-200 ${
                    selectedSide === "NO"
                      ? "border-text-primary bg-surface-tertiary"
                      : "border-surface-tertiary bg-surface-secondary hover:bg-surface-tertiary"
                  }`}
                >
                  <div className="text-sm text-text-tertiary mb-2">NO</div>
                  <div className="text-3xl font-bold font-display text-text-primary tabular-nums">
                    {formatPercentage(market.noPrice)}
                  </div>
                </button>
              </div>

              {/* Amount Input */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">
                  Amount
                </label>
                <Input
                  type="number"
                  placeholder="0.00"
                  value={betAmount}
                  onChange={(e) => setBetAmount(e.target.value)}
                  min="0"
                  step="0.01"
                  className="text-lg"
                />
                <div className="flex items-center gap-2">
                  {[10, 25, 50, 100].map((amount) => (
                    <button
                      key={amount}
                      onClick={() => setBetAmount(amount.toString())}
                      className="flex-1 px-3 py-2 rounded-lg text-sm font-medium bg-surface-secondary hover:bg-surface-tertiary text-text-secondary hover:text-text-primary transition-colors"
                    >
                      ${amount}
                    </button>
                  ))}
                </div>
              </div>

              {/* Summary */}
              {betAmount && parseFloat(betAmount) > 0 && (
                <div className="p-4 rounded-xl bg-surface-secondary space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-tertiary">You pay</span>
                    <span className="font-semibold text-text-primary tabular-nums">
                      {formatCurrency(parseFloat(betAmount))}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-tertiary">You receive</span>
                    <span className="font-semibold text-text-primary tabular-nums">
                      {calculateShares().toFixed(2)} shares
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-tertiary">Average price</span>
                    <span className="font-semibold text-text-primary tabular-nums">
                      {formatPercentage(
                        selectedSide === "YES" ? market.yesPrice : market.noPrice
                      )}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm pt-2 border-t border-surface-tertiary">
                    <span className="text-text-tertiary">Potential return</span>
                    <span className="font-bold text-primary-600 tabular-nums">
                      {formatCurrency(calculateShares())}
                    </span>
                  </div>
                </div>
              )}

              {/* Place Bet Button */}
              <Button
                variant="primary"
                size="lg"
                className="w-full"
                onClick={handlePlaceBet}
                disabled={!betAmount || parseFloat(betAmount) <= 0 || isPlacingBet}
              >
                {isPlacingBet ? (
                  "Placing Bet..."
                ) : (
                  <>
                    Place Bet on {selectedSide}
                    <TrendingUp className="w-5 h-5" />
                  </>
                )}
              </Button>

              <p className="text-xs text-text-tertiary text-center">
                By placing a bet, you agree to our trading terms. All bets are final.
              </p>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  )
}
