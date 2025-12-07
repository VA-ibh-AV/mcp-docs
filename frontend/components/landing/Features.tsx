'use client'

import { FileText, Zap, Shield, Code, Database, Search } from "lucide-react"
import { cn } from "@/lib/utils"

interface Feature {
  icon: React.ComponentType<{ className?: string }>
  title: string
  description: string
}

const features: Feature[] = [
  {
    icon: Zap,
    title: "Lightning Fast Indexing",
    description: "Automatically index your documentation with AI-powered processing. Get your docs searchable in minutes, not hours."
  },
  {
    icon: Search,
    title: "Intelligent Search",
    description: "Find exactly what you need with semantic search powered by advanced AI. Understand context, not just keywords."
  },
  {
    icon: Code,
    title: "MCP Endpoint Ready",
    description: "Transform your documentation into MCP endpoints instantly. No complex setup, no API keys required."
  },
  {
    icon: Shield,
    title: "Secure & Private",
    description: "Your documentation stays private. All processing happens locally or in your secure environment."
  },
  {
    icon: Database,
    title: "Multiple Formats",
    description: "Support for Markdown, HTML, PDF, and more. Index documentation from any source or format."
  },
  {
    icon: FileText,
    title: "Easy Integration",
    description: "Simple CLI tools and APIs make it easy to integrate MCP-Docs into your existing workflow."
  }
]

export default function Features() {
  return (
    <section id="features" className="py-24 px-4 bg-white font-sans">
      <div className="max-w-7xl mx-auto">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-4">
            Powerful Features
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Everything you need to transform your documentation into intelligent MCP endpoints
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => {
            const Icon = feature.icon
            return (
              <div
                key={index}
                className={cn(
                  "bg-white border border-gray-200 rounded-lg p-8",
                  "shadow-md hover:shadow-lg transition-shadow duration-300",
                  "flex flex-col items-start"
                )}
              >
                <div className="mb-4 p-3 bg-primary/10 rounded-lg">
                  <Icon className="w-6 h-6 text-primary" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-3">
                  {feature.title}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

