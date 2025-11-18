"use client"

import { Search, Bell } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { motion } from "framer-motion"

export function DashboardHeader() {
  return (
    <motion.header
      className="sticky top-0 z-30 bg-surface/95 backdrop-blur-sm border-b border-surface-tertiary"
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
    >
      <div className="flex items-center justify-between px-8 py-4">
        {/* Search */}
        <div className="flex-1 max-w-xl">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-text-tertiary" />
            <Input
              type="search"
              placeholder="Search markets..."
              className="pl-11 bg-surface-secondary"
            />
          </div>
        </div>

        {/* Right Side */}
        <div className="flex items-center gap-4">
          {/* Notifications */}
          <button className="relative p-2 rounded-xl hover:bg-surface-secondary transition-colors">
            <Bell className="w-5 h-5 text-text-secondary" />
            <div className="absolute top-1 right-1 w-2 h-2 rounded-full bg-accent-pink-500" />
          </button>

          {/* Balance */}
          <div className="px-4 py-2 rounded-xl bg-gradient-to-br from-primary-100 to-accent-pink-100 border border-primary-200">
            <div className="text-xs text-text-tertiary font-medium">Balance</div>
            <div className="text-lg font-bold font-display text-primary-700 tabular-nums">
              $5,250.00
            </div>
          </div>
        </div>
      </div>
    </motion.header>
  )
}
