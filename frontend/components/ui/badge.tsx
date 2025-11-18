import * as React from "react"
import { cn } from "@/lib/utils"

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "primary" | "success" | "error" | "warning" | "default"
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = "default", ...props }, ref) => {
    const variants = {
      primary: "bg-primary-100 text-primary-800",
      success: "bg-success-light text-success-dark",
      error: "bg-error-light text-error-dark",
      warning: "bg-warning-light text-warning-dark",
      default: "bg-surface-tertiary text-text-secondary",
    }

    return (
      <div
        ref={ref}
        className={cn(
          "px-3 py-1 rounded-full text-xs font-medium inline-flex items-center gap-1",
          variants[variant],
          className
        )}
        {...props}
      />
    )
  }
)
Badge.displayName = "Badge"

export { Badge }
