"use client"

import { useState, useEffect } from "react"
import { motion } from "framer-motion"
import { Trophy, TrendingUp, DollarSign, Award, Medal } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { api, LeaderboardEntry } from "@/lib/api"
import { formatCurrency, formatPercentage } from "@/lib/utils"

const tabs = [
  { value: "volume" as const, label: "Total Volume", icon: DollarSign },
  { value: "winrate" as const, label: "Win Rate", icon: Trophy },
]

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.05,
    },
  },
}

const itemVariants = {
  hidden: { opacity: 0, x: -20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: {
      duration: 0.4,
    },
  },
}

export default function LeaderboardPage() {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [sortBy, setSortBy] = useState<"volume" | "winrate">("volume")

  useEffect(() => {
    api
      .getLeaderboard(sortBy, 50)
      .then((data) => {
        setLeaderboard(data.leaderboard || [])
        setIsLoading(false)
      })
      .catch((error) => {
        console.error("Failed to load leaderboard:", error)
        setIsLoading(false)
      })
  }, [sortBy])

  const getRankBadge = (rank: number) => {
    if (rank === 1) {
      return (
        <div className="w-12 h-12 rounded-full bg-gradient-to-br from-accent-yellow-400 to-accent-yellow-600 flex items-center justify-center shadow-colored-yellow">
          <Trophy className="w-6 h-6 text-white" />
        </div>
      )
    }
    if (rank === 2) {
      return (
        <div className="w-12 h-12 rounded-full bg-gradient-to-br from-gray-300 to-gray-500 flex items-center justify-center shadow-soft">
          <Medal className="w-6 h-6 text-white" />
        </div>
      )
    }
    if (rank === 3) {
      return (
        <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary-300 to-primary-500 flex items-center justify-center shadow-soft">
          <Award className="w-6 h-6 text-white" />
        </div>
      )
    }
    return (
      <div className="w-12 h-12 rounded-full bg-surface-tertiary flex items-center justify-center border-2 border-surface-tertiary">
        <span className="text-lg font-bold text-text-secondary">#{rank}</span>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-4xl font-display font-bold text-text-primary mb-2">
              Leaderboard
            </h1>
            <p className="text-lg text-text-secondary">
              Top traders ranked by performance
            </p>
          </div>
          <div className="flex items-center gap-2">
            {tabs.map((tab) => (
              <button
                key={tab.value}
                onClick={() => setSortBy(tab.value)}
                className={`flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-medium transition-all duration-200 ${
                  sortBy === tab.value
                    ? "bg-primary-500 text-white shadow-soft"
                    : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary hover:text-text-primary"
                }`}
              >
                <tab.icon className="w-4 h-4" />
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </motion.div>

      {/* Top 3 Podium */}
      {!isLoading && leaderboard.length >= 3 && (
        <motion.div
          className="grid grid-cols-3 gap-6"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          {/* 2nd Place */}
          <motion.div
            className="order-1"
            whileHover={{ y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <Card className="text-center h-full">
              <CardContent className="p-8 pt-12">
                <div className="flex justify-center mb-4">
                  {getRankBadge(2)}
                </div>
                <div className="text-2xl font-bold text-text-secondary mb-1">
                  #2
                </div>
                <div className="w-20 h-20 mx-auto mb-4 rounded-full bg-gradient-to-br from-gray-300 to-gray-500 flex items-center justify-center text-white text-2xl font-bold shadow-soft">
                  {leaderboard[1].username.charAt(0).toUpperCase()}
                </div>
                <h3 className="text-xl font-display font-bold text-text-primary mb-2">
                  {leaderboard[1].username}
                </h3>
                <div className="text-3xl font-bold font-display text-text-primary tabular-nums mb-2">
                  {sortBy === "volume"
                    ? formatCurrency(leaderboard[1].totalVolume)
                    : formatPercentage(leaderboard[1].winRate)}
                </div>
                <div className="text-sm text-text-tertiary">
                  {leaderboard[1].totalBets} bets
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* 1st Place */}
          <motion.div
            className="order-2"
            whileHover={{ y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <Card className="text-center h-full bg-gradient-to-br from-accent-yellow-50 to-primary-50 border-2 border-accent-yellow-300 shadow-large">
              <CardContent className="p-8 pt-16">
                <div className="flex justify-center mb-4">
                  {getRankBadge(1)}
                </div>
                <div className="text-3xl font-bold gradient-text mb-2">
                  #1
                </div>
                <div className="w-24 h-24 mx-auto mb-4 rounded-full bg-gradient-to-br from-accent-yellow-400 to-accent-yellow-600 flex items-center justify-center text-white text-3xl font-bold shadow-colored-yellow">
                  {leaderboard[0].username.charAt(0).toUpperCase()}
                </div>
                <h3 className="text-2xl font-display font-bold text-text-primary mb-3">
                  {leaderboard[0].username}
                </h3>
                <div className="text-4xl font-bold font-display text-primary-600 tabular-nums mb-3">
                  {sortBy === "volume"
                    ? formatCurrency(leaderboard[0].totalVolume)
                    : formatPercentage(leaderboard[0].winRate)}
                </div>
                <div className="text-sm text-text-secondary font-medium">
                  {leaderboard[0].totalBets} bets
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* 3rd Place */}
          <motion.div
            className="order-3"
            whileHover={{ y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <Card className="text-center h-full">
              <CardContent className="p-8 pt-12">
                <div className="flex justify-center mb-4">
                  {getRankBadge(3)}
                </div>
                <div className="text-2xl font-bold text-primary-600 mb-1">
                  #3
                </div>
                <div className="w-20 h-20 mx-auto mb-4 rounded-full bg-gradient-to-br from-primary-300 to-primary-500 flex items-center justify-center text-white text-2xl font-bold shadow-soft">
                  {leaderboard[2].username.charAt(0).toUpperCase()}
                </div>
                <h3 className="text-xl font-display font-bold text-text-primary mb-2">
                  {leaderboard[2].username}
                </h3>
                <div className="text-3xl font-bold font-display text-text-primary tabular-nums mb-2">
                  {sortBy === "volume"
                    ? formatCurrency(leaderboard[2].totalVolume)
                    : formatPercentage(leaderboard[2].winRate)}
                </div>
                <div className="text-sm text-text-tertiary">
                  {leaderboard[2].totalBets} bets
                </div>
              </CardContent>
            </Card>
          </motion.div>
        </motion.div>
      )}

      {/* Full Leaderboard */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.4 }}
      >
        <Card>
          <CardHeader>
            <CardTitle>All Rankings</CardTitle>
            <CardDescription>
              Complete leaderboard sorted by {sortBy === "volume" ? "total volume" : "win rate"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-3">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div key={i} className="skeleton h-20 rounded-xl" />
                ))}
              </div>
            ) : (
              <motion.div
                className="space-y-2"
                variants={containerVariants}
                initial="hidden"
                animate="visible"
              >
                {leaderboard.map((entry, index) => (
                  <motion.div
                    key={entry.userId}
                    variants={itemVariants}
                    className={`flex items-center gap-6 p-4 rounded-xl transition-all duration-200 ${
                      index < 3
                        ? "bg-gradient-to-r from-primary-50 to-accent-pink-50 border-2 border-primary-200"
                        : "bg-surface-secondary hover:bg-surface-tertiary border border-transparent hover:border-surface-tertiary"
                    }`}
                  >
                    {/* Rank */}
                    <div className="flex-shrink-0">
                      {getRankBadge(index + 1)}
                    </div>

                    {/* User Info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-1">
                        <h3 className="text-lg font-semibold text-text-primary truncate">
                          {entry.username}
                        </h3>
                        {index < 3 && (
                          <Badge variant="primary" className="flex-shrink-0">
                            Top {index + 1}
                          </Badge>
                        )}
                      </div>
                      <div className="text-sm text-text-tertiary">
                        {entry.totalBets} total bets
                      </div>
                    </div>

                    {/* Stats */}
                    <div className="flex items-center gap-8">
                      <div className="text-right">
                        <div className="text-sm text-text-tertiary mb-1">
                          Volume
                        </div>
                        <div className="text-lg font-bold font-display text-text-primary tabular-nums">
                          {formatCurrency(entry.totalVolume)}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm text-text-tertiary mb-1">
                          Win Rate
                        </div>
                        <div className="text-lg font-bold font-display text-primary-600 tabular-nums">
                          {formatPercentage(entry.winRate)}
                        </div>
                      </div>
                    </div>
                  </motion.div>
                ))}
              </motion.div>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}
