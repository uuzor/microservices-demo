"use client"

import { motion } from "framer-motion"
import { Filter, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

const categories = [
  "All",
  "Politics",
  "Sports",
  "Crypto",
  "Technology",
  "Entertainment",
  "Business",
  "Science",
]

const statuses = [
  { value: "all", label: "All Markets" },
  { value: "active", label: "Active" },
  { value: "closing_soon", label: "Closing Soon" },
  { value: "resolved", label: "Resolved" },
]

interface FiltersProps {
  selectedCategory: string
  selectedStatus: string
  onCategoryChange: (category: string) => void
  onStatusChange: (status: string) => void
  onClear: () => void
}

export function Filters({
  selectedCategory,
  selectedStatus,
  onCategoryChange,
  onStatusChange,
  onClear,
}: FiltersProps) {
  const hasFilters = selectedCategory !== "All" || selectedStatus !== "all"

  return (
    <motion.div
      className="space-y-6"
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Filter className="w-5 h-5 text-text-secondary" />
          <h3 className="text-lg font-semibold text-text-primary">Filters</h3>
        </div>
        {hasFilters && (
          <Button
            variant="outline"
            size="sm"
            onClick={onClear}
            className="text-sm"
          >
            <X className="w-4 h-4" />
            Clear All
          </Button>
        )}
      </div>

      {/* Categories */}
      <div>
        <h4 className="text-sm font-medium text-text-secondary mb-3">Category</h4>
        <div className="flex flex-wrap gap-2">
          {categories.map((category) => (
            <button
              key={category}
              onClick={() => onCategoryChange(category)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                selectedCategory === category
                  ? "bg-primary-500 text-white shadow-soft"
                  : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary hover:text-text-primary"
              }`}
            >
              {category}
            </button>
          ))}
        </div>
      </div>

      {/* Status */}
      <div>
        <h4 className="text-sm font-medium text-text-secondary mb-3">Status</h4>
        <div className="flex flex-wrap gap-2">
          {statuses.map((status) => (
            <button
              key={status.value}
              onClick={() => onStatusChange(status.value)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                selectedStatus === status.value
                  ? "bg-accent-pink-500 text-white shadow-soft"
                  : "bg-surface-secondary text-text-secondary hover:bg-surface-tertiary hover:text-text-primary"
              }`}
            >
              {status.label}
            </button>
          ))}
        </div>
      </div>

      {/* Active Filters */}
      {hasFilters && (
        <div className="pt-4 border-t border-surface-tertiary">
          <h4 className="text-sm font-medium text-text-secondary mb-3">
            Active Filters
          </h4>
          <div className="flex flex-wrap gap-2">
            {selectedCategory !== "All" && (
              <Badge variant="primary" className="gap-2">
                {selectedCategory}
                <button
                  onClick={() => onCategoryChange("All")}
                  className="hover:opacity-70"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            )}
            {selectedStatus !== "all" && (
              <Badge variant="primary" className="gap-2">
                {statuses.find((s) => s.value === selectedStatus)?.label}
                <button
                  onClick={() => onStatusChange("all")}
                  className="hover:opacity-70"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            )}
          </div>
        </div>
      )}
    </motion.div>
  )
}
