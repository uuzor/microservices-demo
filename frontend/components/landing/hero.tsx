"use client"

import { motion } from "framer-motion"
import { TrendingUp, Sparkles, Users, BarChart3 } from "lucide-react"
import { Button } from "@/components/ui/button"
import Link from "next/link"

const floatingAnimation = {
  initial: { y: 0 },
  animate: {
    y: [-10, 10, -10],
    transition: {
      duration: 6,
      repeat: Infinity,
      ease: "easeInOut",
    },
  },
}

const stats = [
  { icon: TrendingUp, label: "Active Markets", value: "500+" },
  { icon: Users, label: "Active Traders", value: "10K+" },
  { icon: BarChart3, label: "Total Volume", value: "$2.5M" },
]

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden bg-gradient-to-br from-surface via-surface to-primary-50/30">
      {/* Animated Background Shapes */}
      <div className="absolute inset-0 overflow-hidden">
        <motion.div
          className="absolute top-20 left-10 w-72 h-72 bg-accent-pink-200/20 rounded-full blur-3xl"
          {...floatingAnimation}
        />
        <motion.div
          className="absolute bottom-20 right-10 w-96 h-96 bg-accent-yellow-200/20 rounded-full blur-3xl"
          animate={{
            y: [10, -10, 10],
            transition: {
              duration: 8,
              repeat: Infinity,
              ease: "easeInOut",
            },
          }}
        />
        <motion.div
          className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-primary-200/10 rounded-full blur-3xl"
          animate={{
            scale: [1, 1.1, 1],
            transition: {
              duration: 10,
              repeat: Infinity,
              ease: "easeInOut",
            },
          }}
        />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-6 py-24 text-center">
        {/* Badge */}
        <motion.div
          className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-surface-secondary/80 backdrop-blur-sm border border-surface-tertiary mb-8"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
        >
          <Sparkles className="w-4 h-4 text-accent-yellow-500" />
          <span className="text-sm font-medium text-text-secondary">
            Powered by AI and Real-Time Data
          </span>
        </motion.div>

        {/* Main Heading */}
        <motion.h1
          className="text-6xl md:text-7xl lg:text-8xl font-display font-bold mb-6 leading-tight"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <span className="gradient-text">Trade the Future</span>
          <br />
          <span className="text-text-primary">Before It Happens</span>
        </motion.h1>

        {/* Subheading */}
        <motion.p
          className="text-xl md:text-2xl text-text-secondary max-w-3xl mx-auto mb-12 leading-relaxed"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.4 }}
        >
          Join the social prediction market where AI meets collective intelligence.
          Bet on trends, earn rewards, and shape the conversation.
        </motion.p>

        {/* CTA Buttons */}
        <motion.div
          className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-20"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.6 }}
        >
          <Link href="/signup">
            <Button variant="primary" size="lg" className="min-w-[200px]">
              Start Trading
              <TrendingUp className="w-5 h-5" />
            </Button>
          </Link>
          <Link href="#how-it-works">
            <Button variant="outline" size="lg" className="min-w-[200px]">
              Learn More
            </Button>
          </Link>
        </motion.div>

        {/* Stats */}
        <motion.div
          className="grid grid-cols-1 sm:grid-cols-3 gap-8 max-w-4xl mx-auto"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.8 }}
        >
          {stats.map((stat, index) => (
            <motion.div
              key={stat.label}
              className="flex flex-col items-center gap-3 p-6 rounded-2xl bg-surface/50 backdrop-blur-sm border border-surface-tertiary shadow-soft"
              whileHover={{ scale: 1.05, y: -5 }}
              transition={{ duration: 0.2 }}
            >
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-500 to-accent-pink-500 flex items-center justify-center shadow-colored-pink">
                <stat.icon className="w-6 h-6 text-white" />
              </div>
              <div className="text-3xl font-bold font-display text-text-primary tabular-nums">
                {stat.value}
              </div>
              <div className="text-sm text-text-secondary font-medium">
                {stat.label}
              </div>
            </motion.div>
          ))}
        </motion.div>
      </div>

      {/* Scroll Indicator */}
      <motion.div
        className="absolute bottom-8 left-1/2 -translate-x-1/2"
        animate={{
          y: [0, 10, 0],
          transition: {
            duration: 2,
            repeat: Infinity,
            ease: "easeInOut",
          },
        }}
      >
        <div className="w-6 h-10 rounded-full border-2 border-text-tertiary flex items-start justify-center p-2">
          <motion.div
            className="w-1.5 h-1.5 rounded-full bg-text-tertiary"
            animate={{
              y: [0, 12, 0],
              transition: {
                duration: 2,
                repeat: Infinity,
                ease: "easeInOut",
              },
            }}
          />
        </div>
      </motion.div>
    </section>
  )
}
