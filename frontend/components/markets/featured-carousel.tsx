"use client"

import { useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { ChevronLeft, ChevronRight } from "lucide-react"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Market } from "@/lib/api"
import { formatCurrency, formatPercentage } from "@/lib/utils"
import Link from "next/link"

interface FeaturedCarouselProps {
  markets: Market[]
}

export function FeaturedCarousel({ markets }: FeaturedCarouselProps) {
  const [currentIndex, setCurrentIndex] = useState(0)

  const next = () => {
    setCurrentIndex((prev) => (prev + 1) % markets.length)
  }

  const previous = () => {
    setCurrentIndex((prev) => (prev - 1 + markets.length) % markets.length)
  }

  if (markets.length === 0) {
    return null
  }

  const currentMarket = markets[currentIndex]

  return (
    <div className="relative">
      <div className="overflow-hidden rounded-2xl">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentIndex}
            initial={{ opacity: 0, x: 100 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -100 }}
            transition={{ duration: 0.3 }}
          >
            <div className="relative h-80 bg-gradient-to-br from-primary-500 via-accent-pink-500 to-accent-yellow-500 p-8 flex items-center">
              {/* Background Pattern */}
              <div className="absolute inset-0 opacity-10">
                <div className="absolute top-10 left-10 w-40 h-40 rounded-full bg-white blur-2xl" />
                <div className="absolute bottom-10 right-10 w-60 h-60 rounded-full bg-white blur-3xl" />
              </div>

              {/* Content */}
              <div className="relative z-10 max-w-3xl">
                <Badge className="mb-4 bg-white/20 text-white border-white/30">
                  Featured Market
                </Badge>
                <h2 className="text-4xl font-display font-bold text-white mb-4 leading-tight">
                  {currentMarket.question}
                </h2>
                <p className="text-lg text-white/90 mb-6 line-clamp-2">
                  {currentMarket.description}
                </p>
                <div className="flex items-center gap-6 mb-6">
                  <div>
                    <div className="text-sm text-white/80 mb-1">YES Price</div>
                    <div className="text-3xl font-bold font-display text-white tabular-nums">
                      {formatPercentage(currentMarket.yesPrice)}
                    </div>
                  </div>
                  <div>
                    <div className="text-sm text-white/80 mb-1">Total Volume</div>
                    <div className="text-2xl font-bold font-display text-white tabular-nums">
                      {formatCurrency(currentMarket.totalVolume)}
                    </div>
                  </div>
                </div>
                <Link href={`/markets/${currentMarket.id}`}>
                  <Button
                    size="lg"
                    className="bg-white text-primary-600 hover:bg-white/90"
                  >
                    Trade Now
                  </Button>
                </Link>
              </div>

              {/* Navigation Buttons */}
              <div className="absolute bottom-6 right-6 flex items-center gap-2">
                <button
                  onClick={previous}
                  className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm border border-white/30 flex items-center justify-center text-white hover:bg-white/30 transition-all"
                >
                  <ChevronLeft className="w-5 h-5" />
                </button>
                <div className="px-4 py-2 rounded-full bg-white/20 backdrop-blur-sm border border-white/30 text-white font-medium tabular-nums">
                  {currentIndex + 1} / {markets.length}
                </div>
                <button
                  onClick={next}
                  className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm border border-white/30 flex items-center justify-center text-white hover:bg-white/30 transition-all"
                >
                  <ChevronRight className="w-5 h-5" />
                </button>
              </div>
            </div>
          </motion.div>
        </AnimatePresence>
      </div>

      {/* Indicators */}
      <div className="flex items-center justify-center gap-2 mt-4">
        {markets.map((_, index) => (
          <button
            key={index}
            onClick={() => setCurrentIndex(index)}
            className={`h-2 rounded-full transition-all duration-300 ${
              index === currentIndex
                ? "w-8 bg-primary-500"
                : "w-2 bg-surface-tertiary hover:bg-text-tertiary"
            }`}
          />
        ))}
      </div>
    </div>
  )
}
