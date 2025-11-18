"use client"

import { useState, useEffect } from "react"
import { motion } from "framer-motion"
import {
  User,
  Award,
  TrendingUp,
  DollarSign,
  Edit,
  Users,
  Calendar,
} from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { api, Bet, User as UserType } from "@/lib/api"
import { formatCurrency, formatPercentage, formatDate } from "@/lib/utils"

const stats = [
  {
    title: "Total Volume",
    value: "$12,450",
    icon: DollarSign,
    gradient: "from-primary-500 to-accent-pink-500",
  },
  {
    title: "Total Bets",
    value: "87",
    icon: TrendingUp,
    gradient: "from-accent-pink-500 to-accent-yellow-500",
  },
  {
    title: "Win Rate",
    value: "68%",
    icon: Award,
    gradient: "from-accent-yellow-500 to-primary-500",
  },
  {
    title: "Followers",
    value: "234",
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

export default function ProfilePage() {
  const [bets, setBets] = useState<Bet[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<"active" | "history">("active")

  useEffect(() => {
    // Mock user ID
    const userId = "user123"

    api
      .getUserBets(userId)
      .then((data) => {
        setBets(data.bets || [])
        setIsLoading(false)
      })
      .catch((error) => {
        console.error("Failed to load bets:", error)
        setIsLoading(false)
      })
  }, [])

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Profile Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <Card>
          <CardContent className="p-8">
            <div className="flex items-start gap-6">
              {/* Avatar */}
              <div className="w-32 h-32 rounded-2xl bg-gradient-to-br from-primary-500 via-accent-pink-500 to-accent-yellow-500 flex items-center justify-center text-white text-5xl font-bold shadow-large">
                JD
              </div>

              {/* Info */}
              <div className="flex-1">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h1 className="text-4xl font-display font-bold text-text-primary mb-2">
                      John Doe
                    </h1>
                    <p className="text-lg text-text-secondary">@johndoe</p>
                  </div>
                  <Button variant="outline" className="gap-2">
                    <Edit className="w-4 h-4" />
                    Edit Profile
                  </Button>
                </div>

                <p className="text-text-secondary mb-6 max-w-2xl">
                  Prediction market enthusiast. Trading on politics, crypto, and tech markets.
                  Always learning and improving my strategy. 📈
                </p>

                <div className="flex items-center gap-6 text-sm text-text-tertiary">
                  <div className="flex items-center gap-2">
                    <Calendar className="w-4 h-4" />
                    <span>Joined January 2025</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    <span>234 followers • 189 following</span>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
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
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${stat.gradient} flex items-center justify-center shadow-colored-pink`}>
                    <stat.icon className="w-6 h-6 text-white" />
                  </div>
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

      {/* Bets Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.4 }}
      >
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>My Bets</CardTitle>
                <CardDescription>Track your predictions and positions</CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setActiveTab("active")}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                    activeTab === "active"
                      ? "bg-primary-500 text-white shadow-soft"
                      : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary"
                  }`}
                >
                  Active
                </button>
                <button
                  onClick={() => setActiveTab("history")}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                    activeTab === "history"
                      ? "bg-primary-500 text-white shadow-soft"
                      : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary"
                  }`}
                >
                  History
                </button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="skeleton h-24 rounded-xl" />
                ))}
              </div>
            ) : bets.length === 0 ? (
              <div className="text-center py-16">
                <div className="text-6xl mb-4">📊</div>
                <h3 className="text-2xl font-display font-bold text-text-primary mb-2">
                  No bets yet
                </h3>
                <p className="text-text-secondary mb-6">
                  Start trading to see your positions here
                </p>
                <Button variant="primary">Explore Markets</Button>
              </div>
            ) : (
              <div className="space-y-4">
                {bets.map((bet) => (
                  <div
                    key={bet.id}
                    className="p-6 rounded-xl bg-surface-secondary hover:bg-surface-tertiary transition-all duration-200 border border-surface-tertiary hover:border-text-tertiary"
                  >
                    <div className="flex items-start justify-between gap-6">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-3">
                          <Badge
                            variant={bet.side === "YES" ? "primary" : "default"}
                          >
                            {bet.side}
                          </Badge>
                          <span className="text-sm text-text-tertiary">
                            {formatDate(bet.timestamp)}
                          </span>
                        </div>
                        <h3 className="text-lg font-semibold text-text-primary mb-2">
                          Market #{bet.marketId}
                        </h3>
                        <div className="flex items-center gap-6 text-sm">
                          <div>
                            <span className="text-text-tertiary">Amount: </span>
                            <span className="text-text-primary font-semibold tabular-nums">
                              {formatCurrency(bet.amount)}
                            </span>
                          </div>
                          <div>
                            <span className="text-text-tertiary">Shares: </span>
                            <span className="text-text-primary font-semibold tabular-nums">
                              {bet.shares.toFixed(2)}
                            </span>
                          </div>
                          <div>
                            <span className="text-text-tertiary">Avg Price: </span>
                            <span className="text-text-primary font-semibold tabular-nums">
                              {formatPercentage(bet.price)}
                            </span>
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm text-text-tertiary mb-1">
                          Current Value
                        </div>
                        <div className="text-2xl font-bold font-display text-primary-600 tabular-nums">
                          {formatCurrency(bet.shares * bet.price)}
                        </div>
                        <div className="text-sm text-success-dark font-semibold">
                          +{((bet.shares * bet.price - bet.amount) / bet.amount * 100).toFixed(1)}%
                        </div>
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
  )
}
