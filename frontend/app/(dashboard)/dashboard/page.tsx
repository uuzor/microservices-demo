"use client"

import { useState, useEffect } from "react"
import { motion } from "framer-motion"
import { TrendingUp, Users, Award, DollarSign, ArrowRight } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import Link from "next/link"
import { api, Market, FeedEvent, LeaderboardEntry } from "@/lib/api"
import { formatCurrency, formatPercentage } from "@/lib/utils"

const stats = [
  {
    title: "Portfolio Value",
    value: "$5,250.00",
    change: "+12.5%",
    icon: DollarSign,
    gradient: "from-primary-500 to-accent-pink-500",
  },
  {
    title: "Active Positions",
    value: "8",
    change: "+2",
    icon: TrendingUp,
    gradient: "from-accent-pink-500 to-accent-yellow-500",
  },
  {
    title: "Win Rate",
    value: "68%",
    change: "+5%",
    icon: Award,
    gradient: "from-accent-yellow-500 to-primary-500",
  },
  {
    title: "Rank",
    value: "#42",
    change: "↑8",
    icon: Users,
    gradient: "from-primary-600 to-accent-pink-600",
  },
]

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
    },
  },
}

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.4,
    },
  },
}

export default function DashboardPage() {
  const [recommendations, setRecommendations] = useState<Market[]>([])
  const [feed, setFeed] = useState<FeedEvent[]>([])
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Mock user ID - in real app, get from auth context
    const userId = "user123"

    api
      .getHomepage(userId)
      .then((data) => {
        setRecommendations(data.recommendations)
        setFeed(data.feed)
        setLeaderboard(data.leaderboard)
        setIsLoading(false)
      })
      .catch((error) => {
        console.error("Failed to load dashboard:", error)
        setIsLoading(false)
      })
  }, [])

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Welcome Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <h1 className="text-4xl font-display font-bold text-text-primary mb-2">
          Welcome back, John! 👋
        </h1>
        <p className="text-lg text-text-secondary">
          Here's what's happening with your predictions today.
        </p>
      </motion.div>

      {/* Stats Grid */}
      <motion.div
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {stats.map((stat) => (
          <motion.div key={stat.title} variants={itemVariants}>
            <Card>
              <CardContent className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${stat.gradient} flex items-center justify-center shadow-colored-pink`}>
                    <stat.icon className="w-6 h-6 text-white" />
                  </div>
                  <Badge variant="success" className="text-xs">
                    {stat.change}
                  </Badge>
                </div>
                <div className="text-3xl font-bold font-display text-text-primary tabular-nums mb-1">
                  {stat.value}
                </div>
                <div className="text-sm text-text-secondary">{stat.title}</div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Recommended Markets */}
        <motion.div
          className="lg:col-span-2 space-y-6"
          variants={itemVariants}
          initial="hidden"
          animate="visible"
        >
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-display font-bold text-text-primary">
              Recommended for You
            </h2>
            <Link href="/markets">
              <Button variant="outline" size="sm">
                View All
                <ArrowRight className="w-4 h-4" />
              </Button>
            </Link>
          </div>

          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="skeleton h-32 rounded-2xl" />
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              {recommendations.slice(0, 3).map((market) => (
                <Link key={market.id} href={`/markets/${market.id}`}>
                  <Card interactive>
                    <CardContent className="p-6">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-2">
                            <Badge variant="primary">{market.category}</Badge>
                            <span className="text-xs text-text-tertiary">
                              {market.expirationDate}
                            </span>
                          </div>
                          <h3 className="text-lg font-semibold text-text-primary mb-2">
                            {market.question}
                          </h3>
                          <p className="text-sm text-text-secondary line-clamp-2">
                            {market.description}
                          </p>
                        </div>
                        <div className="text-right">
                          <div className="text-2xl font-bold font-display text-primary-600 tabular-nums">
                            {formatPercentage(market.yesPrice)}
                          </div>
                          <div className="text-xs text-text-tertiary">YES</div>
                        </div>
                      </div>
                      <div className="mt-4 pt-4 border-t border-surface-tertiary flex items-center justify-between text-sm">
                        <span className="text-text-tertiary">
                          Volume: {formatCurrency(market.totalVolume)}
                        </span>
                        <Button size="sm" variant="primary">
                          Trade
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          )}
        </motion.div>

        {/* Sidebar */}
        <motion.div
          className="space-y-6"
          variants={itemVariants}
          initial="hidden"
          animate="visible"
        >
          {/* Leaderboard */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Award className="w-5 h-5 text-primary-600" />
                Top Traders
              </CardTitle>
              <CardDescription>This week's leaderboard</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-3">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="skeleton h-12 rounded-lg" />
                  ))}
                </div>
              ) : (
                <div className="space-y-3">
                  {leaderboard.slice(0, 5).map((entry, index) => (
                    <div
                      key={entry.userId}
                      className="flex items-center gap-3 p-3 rounded-lg hover:bg-surface-secondary transition-colors"
                    >
                      <div
                        className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm ${
                          index === 0
                            ? "bg-accent-yellow-500 text-white"
                            : index === 1
                            ? "bg-surface-tertiary text-text-secondary"
                            : index === 2
                            ? "bg-primary-200 text-primary-700"
                            : "bg-surface-tertiary text-text-tertiary"
                        }`}
                      >
                        {index + 1}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-semibold text-text-primary truncate">
                          {entry.username}
                        </div>
                        <div className="text-xs text-text-tertiary">
                          {formatCurrency(entry.totalVolume)}
                        </div>
                      </div>
                      <div className="text-sm font-semibold text-primary-600">
                        {formatPercentage(entry.winRate)}
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <Link href="/leaderboard" className="block mt-4">
                <Button variant="outline" size="sm" className="w-full">
                  View Full Leaderboard
                </Button>
              </Link>
            </CardContent>
          </Card>

          {/* Recent Activity */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-primary-600" />
                Recent Activity
              </CardTitle>
              <CardDescription>Live market activity</CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-3">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="skeleton h-16 rounded-lg" />
                  ))}
                </div>
              ) : (
                <div className="space-y-3">
                  {feed.slice(0, 4).map((event) => (
                    <div
                      key={event.id}
                      className="p-3 rounded-lg bg-surface-secondary hover:bg-surface-tertiary transition-colors"
                    >
                      <div className="flex items-start gap-2">
                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary-500 to-accent-pink-500 flex items-center justify-center text-white font-bold text-xs flex-shrink-0">
                          {event.username.charAt(0)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="text-xs text-text-tertiary mb-1">
                            {event.username} • {event.type}
                          </div>
                          <div className="text-sm text-text-primary font-medium line-clamp-2">
                            {event.marketQuestion}
                          </div>
                          {event.amount && (
                            <div className="text-xs text-primary-600 font-semibold mt-1">
                              {event.side} • {formatCurrency(event.amount)}
                            </div>
                          )}
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
    </div>
  )
}
