"use client"

import { motion } from "framer-motion"
import { UserPlus, Search, TrendingUp, Trophy } from "lucide-react"

const steps = [
  {
    icon: UserPlus,
    title: "Create Your Account",
    description: "Sign up in seconds and get instant access to hundreds of prediction markets.",
    step: "01",
  },
  {
    icon: Search,
    title: "Discover Markets",
    description: "Browse markets across politics, sports, crypto, entertainment, and more. Filter by category or let AI recommend the best opportunities.",
    step: "02",
  },
  {
    icon: TrendingUp,
    title: "Place Your Bets",
    description: "Buy YES or NO shares on outcomes you believe in. Prices adjust dynamically based on market activity.",
    step: "03",
  },
  {
    icon: Trophy,
    title: "Earn Rewards",
    description: "Win when your predictions come true. Build your reputation and climb the leaderboard.",
    step: "04",
  },
]

export function HowItWorks() {
  return (
    <section id="how-it-works" className="py-24 bg-gradient-to-b from-surface to-primary-50/20">
      <div className="max-w-7xl mx-auto px-6">
        {/* Section Header */}
        <motion.div
          className="text-center mb-20"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <h2 className="text-5xl md:text-6xl font-display font-bold mb-6">
            How It <span className="gradient-text">Works</span>
          </h2>
          <p className="text-xl text-text-secondary max-w-2xl mx-auto">
            Start trading in minutes. It's that simple.
          </p>
        </motion.div>

        {/* Steps */}
        <div className="relative">
          {/* Connection Line */}
          <div className="hidden lg:block absolute top-1/2 left-0 right-0 h-0.5 bg-gradient-to-r from-primary-200 via-accent-pink-200 to-accent-yellow-200 -translate-y-1/2" />

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 lg:gap-4">
            {steps.map((step, index) => (
              <motion.div
                key={step.title}
                className="relative"
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: index * 0.1 }}
              >
                <div className="relative z-10 bg-surface rounded-2xl p-8 shadow-soft border border-surface-tertiary hover:shadow-large transition-all duration-300">
                  {/* Step Number */}
                  <div className="absolute -top-4 -right-4 w-12 h-12 rounded-full bg-gradient-to-br from-primary-500 via-accent-pink-500 to-accent-yellow-500 flex items-center justify-center shadow-colored-pink">
                    <span className="text-white font-bold font-display">
                      {step.step}
                    </span>
                  </div>

                  {/* Icon */}
                  <div className="w-16 h-16 rounded-xl bg-gradient-to-br from-primary-100 to-accent-pink-100 flex items-center justify-center mb-6">
                    <step.icon className="w-8 h-8 text-primary-600" />
                  </div>

                  {/* Content */}
                  <h3 className="text-2xl font-display font-bold text-text-primary mb-3">
                    {step.title}
                  </h3>
                  <p className="text-text-secondary leading-relaxed">
                    {step.description}
                  </p>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
