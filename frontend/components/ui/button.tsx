import * as React from "react"
import { cn } from "@/lib/utils"
import { motion } from "framer-motion"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "pink" | "yellow" | "outline"
  size?: "sm" | "md" | "lg"
  animated?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", animated = true, children, ...props }, ref) => {
    const baseStyles = "inline-flex items-center justify-center gap-2 font-medium transition-all duration-200 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"

    const variants = {
      primary: "bg-primary-500 text-white hover:bg-primary-600 shadow-soft hover:shadow-medium focus-visible:ring-primary-500",
      secondary: "bg-surface-secondary text-text-primary hover:bg-surface-tertiary shadow-soft hover:shadow-medium",
      pink: "bg-accent-pink-500 text-white hover:bg-accent-pink-600 shadow-colored-pink hover:shadow-large focus-visible:ring-accent-pink-500",
      yellow: "bg-accent-yellow-500 text-white hover:bg-accent-yellow-600 shadow-colored-yellow hover:shadow-large focus-visible:ring-accent-yellow-500",
      outline: "border-2 border-text-tertiary text-text-primary hover:border-text-secondary hover:bg-surface-secondary",
    }

    const sizes = {
      sm: "px-4 py-2 text-sm",
      md: "px-6 py-3 text-base",
      lg: "px-8 py-4 text-lg",
    }

    const ButtonComponent = animated ? motion.button : "button"

    return (
      <ButtonComponent
        className={cn(baseStyles, variants[variant], sizes[size], className)}
        ref={ref}
        whileHover={animated ? { scale: 1.02 } : undefined}
        whileTap={animated ? { scale: 0.98 } : undefined}
        {...(props as any)}
      >
        {children}
      </ButtonComponent>
    )
  }
)

Button.displayName = "Button"

export { Button }
