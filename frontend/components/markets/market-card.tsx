"use client"

import { motion } from "framer-motion"
import { TrendingUp, TrendingDown, Clock } from "lucide-react"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Market } from "@/lib/api"
import { formatCurrency, formatPercentage, formatRelativeTime } from "@/lib/utils"
import Link from "next/link"

interface MarketCardProps {
  market: Market
  index?: number
}

export function MarketCard({ market, index = 0 }: MarketCardProps) {
  const priceChange = market.yesPrice - 0.5
  const isIncreasing = priceChange > 0

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.05 }}
    >
      <Link href={`/markets/${market.id}`}>
        <Card interactive>
          <CardContent className="p-6">
            {/* Header */}
            <div className="flex items-start justify-between gap-4 mb-4">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <Badge variant="primary">{market.category}</Badge>
                  <div className="flex items-center gap-1 text-xs text-text-tertiary">
                    <Clock className="w-3 h-3" />
                    {formatRelativeTime(market.expirationDate)}
                  </div>
                </div>
                <h3 className="text-lg font-semibold text-text-primary mb-2 line-clamp-2">
                  {market.question}
                </h3>
                <p className="text-sm text-text-secondary line-clamp-2">
                  {market.description}
                </p>
              </div>
            </div>

            {/* Price Section */}
            <div className="mb-4">
              <div className="flex items-end gap-4">
                <div className="flex-1">
                  <div className="text-xs text-text-tertiary mb-1">YES</div>
                  <div className="text-3xl font-bold font-display text-primary-600 tabular-nums">
                    {formatPercentage(market.yesPrice)}
                  </div>
                </div>
                <div className="flex-1">
                  <div className="text-xs text-text-tertiary mb-1">NO</div>
                  <div className="text-3xl font-bold font-display text-text-primary tabular-nums">
                    {formatPercentage(market.noPrice)}
                  </div>
                </div>
                <div
                  className={`flex items-center gap-1 px-3 py-2 rounded-lg ${
                    isIncreasing
                      ? "bg-success-light text-success-dark"
                      : "bg-error-light text-error-dark"
                  }`}
                >
                  {isIncreasing ? (
                    <TrendingUp className="w-4 h-4" />
                  ) : (
                    <TrendingDown className="w-4 h-4" />
                  )}
                  <span className="text-sm font-semibold tabular-nums">
                    {Math.abs(priceChange * 100).toFixed(1)}%
                  </span>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="pt-4 border-t border-surface-tertiary flex items-center justify-between">
              <div className="text-sm">
                <span className="text-text-tertiary">Volume: </span>
                <span className="text-text-primary font-semibold tabular-nums">
                  {formatCurrency(market.totalVolume)}
                </span>
              </div>
              <Button size="sm" variant="primary">
                Trade
              </Button>
            </div>
          </CardContent>
        </Card>
      </Link>
    </motion.div>
  )
}
