"use client"

import { useState, useEffect } from "react"
import { motion } from "framer-motion"
import { SlidersHorizontal, ArrowUpDown } from "lucide-react"
import { FeaturedCarousel } from "@/components/markets/featured-carousel"
import { Filters } from "@/components/markets/filters"
import { MarketCard } from "@/components/markets/market-card"
import { Button } from "@/components/ui/button"
import { api, Market } from "@/lib/api"

type SortOption = "volume" | "price" | "ending" | "newest"

const sortOptions = [
  { value: "volume" as const, label: "Highest Volume" },
  { value: "price" as const, label: "Price Change" },
  { value: "ending" as const, label: "Ending Soon" },
  { value: "newest" as const, label: "Newest" },
]

export default function MarketsPage() {
  const [markets, setMarkets] = useState<Market[]>([])
  const [filteredMarkets, setFilteredMarkets] = useState<Market[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [selectedCategory, setSelectedCategory] = useState("All")
  const [selectedStatus, setSelectedStatus] = useState("all")
  const [sortBy, setSortBy] = useState<SortOption>("volume")
  const [showFilters, setShowFilters] = useState(true)

  useEffect(() => {
    api
      .getMarkets()
      .then((data) => {
        setMarkets(data.markets)
        setFilteredMarkets(data.markets)
        setIsLoading(false)
      })
      .catch((error) => {
        console.error("Failed to load markets:", error)
        setIsLoading(false)
      })
  }, [])

  useEffect(() => {
    let filtered = [...markets]

    // Apply category filter
    if (selectedCategory !== "All") {
      filtered = filtered.filter(
        (m) => m.category.toLowerCase() === selectedCategory.toLowerCase()
      )
    }

    // Apply status filter
    if (selectedStatus === "active") {
      filtered = filtered.filter((m) => m.status === "active")
    } else if (selectedStatus === "closing_soon") {
      // Filter markets closing within 24 hours
      const oneDayFromNow = new Date()
      oneDayFromNow.setDate(oneDayFromNow.getDate() + 1)
      filtered = filtered.filter(
        (m) => new Date(m.expirationDate) <= oneDayFromNow
      )
    } else if (selectedStatus === "resolved") {
      filtered = filtered.filter((m) => m.status === "resolved")
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case "volume":
          return b.totalVolume - a.totalVolume
        case "price":
          return Math.abs(b.yesPrice - 0.5) - Math.abs(a.yesPrice - 0.5)
        case "ending":
          return (
            new Date(a.expirationDate).getTime() -
            new Date(b.expirationDate).getTime()
          )
        case "newest":
          return b.id.localeCompare(a.id)
        default:
          return 0
      }
    })

    setFilteredMarkets(filtered)
  }, [markets, selectedCategory, selectedStatus, sortBy])

  const handleClearFilters = () => {
    setSelectedCategory("All")
    setSelectedStatus("all")
  }

  const featuredMarkets = markets.slice(0, 5)

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <h1 className="text-4xl font-display font-bold text-text-primary mb-2">
          Explore Markets
        </h1>
        <p className="text-lg text-text-secondary">
          Discover and trade on hundreds of prediction markets
        </p>
      </motion.div>

      {/* Featured Carousel */}
      {!isLoading && featuredMarkets.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <FeaturedCarousel markets={featuredMarkets} />
        </motion.div>
      )}

      {/* Controls */}
      <motion.div
        className="flex items-center justify-between gap-4 flex-wrap"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.3 }}
      >
        <Button
          variant="outline"
          onClick={() => setShowFilters(!showFilters)}
          className="gap-2"
        >
          <SlidersHorizontal className="w-4 h-4" />
          {showFilters ? "Hide" : "Show"} Filters
        </Button>

        <div className="flex items-center gap-2">
          <ArrowUpDown className="w-4 h-4 text-text-tertiary" />
          <span className="text-sm text-text-secondary">Sort by:</span>
          <div className="flex items-center gap-2">
            {sortOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => setSortBy(option.value)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                  sortBy === option.value
                    ? "bg-primary-500 text-white shadow-soft"
                    : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary hover:text-text-primary"
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>
      </motion.div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        {/* Filters Sidebar */}
        {showFilters && (
          <motion.div
            className="lg:col-span-1"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.4 }}
          >
            <div className="sticky top-24">
              <div className="p-6 rounded-2xl bg-surface border border-surface-tertiary shadow-soft">
                <Filters
                  selectedCategory={selectedCategory}
                  selectedStatus={selectedStatus}
                  onCategoryChange={setSelectedCategory}
                  onStatusChange={setSelectedStatus}
                  onClear={handleClearFilters}
                />
              </div>
            </div>
          </motion.div>
        )}

        {/* Markets Grid */}
        <div className={showFilters ? "lg:col-span-3" : "lg:col-span-4"}>
          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <div key={i} className="skeleton h-80 rounded-2xl" />
              ))}
            </div>
          ) : filteredMarkets.length === 0 ? (
            <div className="text-center py-16">
              <div className="text-6xl mb-4">📊</div>
              <h3 className="text-2xl font-display font-bold text-text-primary mb-2">
                No markets found
              </h3>
              <p className="text-text-secondary mb-6">
                Try adjusting your filters to see more markets
              </p>
              <Button variant="primary" onClick={handleClearFilters}>
                Clear Filters
              </Button>
            </div>
          ) : (
            <>
              <div className="mb-6 text-sm text-text-secondary">
                Showing {filteredMarkets.length} market{filteredMarkets.length !== 1 ? "s" : ""}
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {filteredMarkets.map((market, index) => (
                  <MarketCard key={market.id} market={market} index={index} />
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
