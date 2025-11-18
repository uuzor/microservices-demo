"use client"

import { Sidebar } from "@/components/dashboard/sidebar"
import { DashboardHeader } from "@/components/dashboard/header"
import { useState, useEffect } from "react"

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const [sidebarWidth, setSidebarWidth] = useState(280)

  useEffect(() => {
    const handleResize = () => {
      // Responsive sidebar width
      if (window.innerWidth < 1024) {
        setSidebarWidth(80)
      } else {
        setSidebarWidth(280)
      }
    }

    handleResize()
    window.addEventListener("resize", handleResize)
    return () => window.removeEventListener("resize", handleResize)
  }, [])

  return (
    <div className="min-h-screen bg-surface">
      <Sidebar />
      <div style={{ marginLeft: sidebarWidth }} className="transition-all duration-300">
        <DashboardHeader />
        <main className="p-8">
          {children}
        </main>
      </div>
    </div>
  )
}
