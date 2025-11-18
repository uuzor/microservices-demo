"use client"

import { motion } from "framer-motion"
import { Brain, Zap, Shield, TrendingUp, Users, Award } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

const features = [
  {
    icon: Brain,
    title: "AI-Powered Insights",
    description: "Get personalized market recommendations powered by Google's Gemini AI and advanced analytics.",
    gradient: "from-primary-500 to-accent-pink-500",
  },
  {
    icon: Zap,
    title: "Real-Time Trading",
    description: "Trade with instant execution powered by Kafka event streaming and automated market makers.",
    gradient: "from-accent-pink-500 to-accent-yellow-500",
  },
  {
    icon: Shield,
    title: "Fraud Detection",
    description: "Advanced pattern recognition protects against wash trading, bot activity, and market manipulation.",
    gradient: "from-accent-yellow-500 to-primary-500",
  },
  {
    icon: TrendingUp,
    title: "Dynamic Pricing",
    description: "Fair prices set by automated market makers (AMM) that adjust based on real-time supply and demand.",
    gradient: "from-primary-600 to-accent-pink-600",
  },
  {
    icon: Users,
    title: "Social Feed",
    description: "Follow top traders, see live market activity, and discover trending predictions from the community.",
    gradient: "from-accent-pink-600 to-accent-yellow-600",
  },
  {
    icon: Award,
    title: "Leaderboards & Rewards",
    description: "Compete with traders worldwide. Climb the rankings based on volume, win rate, and profit.",
    gradient: "from-accent-yellow-600 to-primary-600",
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
      duration: 0.5,
    },
  },
}

export function Features() {
  return (
    <section id="features" className="py-24 bg-surface">
      <div className="max-w-7xl mx-auto px-6">
        {/* Section Header */}
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <h2 className="text-5xl md:text-6xl font-display font-bold mb-6">
            <span className="gradient-text">Why VibeCast?</span>
          </h2>
          <p className="text-xl text-text-secondary max-w-2xl mx-auto">
            Advanced technology meets intuitive design. Trade smarter with AI-powered predictions.
          </p>
        </motion.div>

        {/* Features Grid */}
        <motion.div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8"
          variants={containerVariants}
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, margin: "-100px" }}
        >
          {features.map((feature, index) => (
            <motion.div key={feature.title} variants={itemVariants}>
              <Card className="h-full group hover:shadow-large transition-all duration-300">
                <CardHeader>
                  <div className={`w-14 h-14 rounded-xl bg-gradient-to-br ${feature.gradient} flex items-center justify-center mb-4 shadow-colored-pink group-hover:scale-110 transition-transform duration-300`}>
                    <feature.icon className="w-7 h-7 text-white" />
                  </div>
                  <CardTitle className="text-2xl mb-2">{feature.title}</CardTitle>
                  <CardDescription className="text-base leading-relaxed">
                    {feature.description}
                  </CardDescription>
                </CardHeader>
              </Card>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
