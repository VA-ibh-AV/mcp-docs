# UI Design Specification - MCP-Docs Platform

**Version:** 1.0  
**Date:** 2024  
**Target Audience:** UI/UX Developer

---

## Table of Contents

1. [Overview](#overview)
2. [Design System](#design-system)
3. [Landing Page](#landing-page)
4. [User Dashboard](#user-dashboard)
5. [Component Specifications](#component-specifications)
6. [Responsive Design](#responsive-design)
7. [User Flows](#user-flows)
8. [API Integration Points](#api-integration-points)

---

## Overview

This document provides detailed UI/UX specifications for the MCP-Docs platform - a SaaS application for indexing documentation and managing MCP (Model Context Protocol) endpoints. The platform uses a subscription-based model with tiered access.

### Technology Stack

- **Frontend Framework:** Next.js 14+ (App Router) - **Primary Framework**
- **UI Library:** shadcn/ui (built on Radix UI and Tailwind CSS)
- **Icons:** lucide-react (recommended icon library for shadcn/ui)
- **Styling:** Tailwind CSS
- **TypeScript:** Required for all components
- **State Management:** React Query / TanStack Query (for server state)
- **Forms:** React Hook Form + Zod validation

### Next.js Implementation Details

**Project Structure:**
```
app/
├── (auth)/
│   ├── login/
│   └── register/
├── (dashboard)/
│   ├── dashboard/
│   ├── projects/
│   ├── subscription/
│   └── settings/
├── layout.tsx
├── page.tsx (landing page)
└── globals.css
```

**Key Next.js Features to Use:**
- **App Router:** Use Next.js 14+ App Router (not Pages Router)
- **Server Components:** Use Server Components by default for static content
- **Client Components:** Mark interactive components with `'use client'` directive
- **Route Groups:** Use `(auth)` and `(dashboard)` route groups for layout organization
- **Server Actions:** Use Server Actions for form submissions (optional, or use API routes)
- **API Routes:** Create API routes in `app/api/` if needed for client-side data fetching
- **Metadata:** Use Next.js Metadata API for SEO

**Routing Structure:**
- `/` - Landing page (public)
- `/auth/login` - Login page
- `/auth/register` - Registration page
- `/dashboard` - Main dashboard (protected)
- `/dashboard/projects` - Projects list
- `/dashboard/projects/[id]` - Project details
- `/dashboard/subscription` - Subscription management
- `/dashboard/settings` - User settings

**Authentication:**
- Use middleware for route protection
- Implement authentication context/provider
- Store auth tokens in httpOnly cookies or secure storage

### shadcn/ui Setup

1. Initialize shadcn/ui in your Next.js project:
   ```bash
   npx shadcn-ui@latest init
   ```

2. Configure `components.json` with your design system colors

3. Install required components as needed:
   ```bash
   npx shadcn-ui@latest add button
   npx shadcn-ui@latest add card
   npx shadcn-ui@latest add dialog
   # ... etc
   ```

4. Install lucide-react for icons:
   ```bash
   npm install lucide-react
   ```

### Key Pages to Design

1. **Landing Page** (`/`) - Public-facing homepage
2. **User Dashboard** (`/dashboard`) - Main user interface after login
3. **Subscription Management** (`/subscription`) - Manage subscription and billing
4. **MCP Endpoint Management** (`/dashboard/endpoints`) - Create and manage MCP endpoints

---

## Design System

### Color Palette

**Primary Colors:**
- Primary: `#3B82F6` (Blue-500) - Main CTA buttons, links
- Primary Dark: `#2563EB` (Blue-600) - Hover states
- Primary Light: `#60A5FA` (Blue-400) - Secondary actions

**Secondary Colors:**
- Success: `#10B981` (Green-500)
- Warning: `#F59E0B` (Amber-500)
- Error: `#EF4444` (Red-500)
- Info: `#06B6D4` (Cyan-500)

**Neutral Colors:**
- Background: `#FFFFFF` (White)
- Surface: `#F9FAFB` (Gray-50)
- Border: `#E5E7EB` (Gray-200)
- Text Primary: `#111827` (Gray-900)
- Text Secondary: `#6B7280` (Gray-500)
- Text Muted: `#9CA3AF` (Gray-400)

### Typography

**Font Family:**
- Primary: `Inter` or `System UI` (sans-serif)
- Code: `Fira Code` or `Monaco` (monospace)

**Font Sizes:**
- H1: `3rem` (48px) - Landing page hero
- H2: `2.25rem` (36px) - Section headers
- H3: `1.875rem` (30px) - Subsection headers
- H4: `1.5rem` (24px) - Card titles
- Body: `1rem` (16px) - Default text
- Small: `0.875rem` (14px) - Secondary text
- Code: `0.875rem` (14px) - Code blocks

**Font Weights:**
- Light: 300
- Regular: 400
- Medium: 500
- Semibold: 600
- Bold: 700

### Spacing

Use 4px base unit:
- XS: `0.25rem` (4px)
- SM: `0.5rem` (8px)
- MD: `1rem` (16px)
- LG: `1.5rem` (24px)
- XL: `2rem` (32px)
- 2XL: `3rem` (48px)
- 3XL: `4rem` (64px)

### Border Radius

- Small: `0.25rem` (4px) - Buttons, inputs
- Medium: `0.5rem` (8px) - Cards
- Large: `1rem` (16px) - Modals, containers

### Shadows

- Small: `0 1px 2px 0 rgba(0, 0, 0, 0.05)`
- Medium: `0 4px 6px -1px rgba(0, 0, 0, 0.1)`
- Large: `0 10px 15px -3px rgba(0, 0, 0, 0.1)`

---

## Landing Page

### Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│                    Navigation Bar                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│                    Hero Section                         │
│              (Try Now + Install Box)                    │
│                                                         │
├─────────────────────────────────────────────────────────┤
│                    Features Section                     │
├─────────────────────────────────────────────────────────┤
│                  Pricing Section                        │
├─────────────────────────────────────────────────────────┤
│                    Footer                               │
└─────────────────────────────────────────────────────────┘
```

### Navigation Bar

**Position:** Fixed at top, sticky on scroll

**Content:**
- **Left Side:**
  - Logo: "MCP-Docs" (text or image)
  - Navigation Links:
    - Features
    - Pricing
    - Documentation
    - About

- **Right Side:**
  - "Sign In" button (text link, secondary style)
  - "Try Free" button (primary button style)

**Styling:**
- Background: White with subtle shadow
- Height: `64px`
- Padding: `0 2rem` (horizontal)
- Border bottom: `1px solid #E5E7EB`

### Hero Section

**Layout:** Centered, full-width with max-width container (1200px)

**Content Structure:**

```
┌─────────────────────────────────────────────────────┐
│                                                      │
│         Main Headline (H1)                          │
│    "Index Your Documentation with AI"              │
│                                                      │
│    Subheadline (Body, large, gray)                  │
│    "Transform your docs into intelligent MCP        │
│     endpoints. No API keys required."              │
│                                                      │
│    ┌──────────────────────────────────────┐        │
│    │  [Try Free]  [Learn More]            │        │
│    └──────────────────────────────────────┘        │
│                                                      │
│    ┌──────────────────────────────────────┐        │
│    │  pip install mcp-docs                │        │
│    │  [Copy Icon]                         │        │
│    └──────────────────────────────────────┘        │
│                                                      │
│    Optional: Hero Image/Illustration                │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Specifications:**

1. **Main Headline:**
   - Text: "Index Your Documentation with AI"
   - Size: H1 (3rem / 48px)
   - Weight: Bold (700)
   - Color: Text Primary (#111827)
   - Max-width: 800px
   - Center aligned
   - Margin bottom: `1.5rem`

2. **Subheadline:**
   - Text: "Transform your docs into intelligent MCP endpoints. No API keys required."
   - Size: Body Large (1.125rem / 18px)
   - Weight: Regular (400)
   - Color: Text Secondary (#6B7280)
   - Max-width: 600px
   - Center aligned
   - Margin bottom: `2.5rem`

3. **CTA Buttons:**
   - **"Try Free" Button:**
     - Style: Primary button
     - Background: Primary (#3B82F6)
     - Text: White
     - Padding: `0.75rem 2rem`
     - Border radius: Small (4px)
     - Font size: Body (16px)
     - Font weight: Semibold (600)
     - Hover: Darker blue (#2563EB)
     - Margin right: `1rem`
   
   - **"Learn More" Button:**
     - Style: Secondary button (outline)
     - Border: 1px solid Primary
     - Text: Primary color
     - Background: Transparent
     - Same padding and sizing as primary
     - Hover: Light blue background (#EFF6FF)

4. **Installation Code Box:**
   - Container: Card style with code block appearance
   - Background: `#1F2937` (Dark gray, code-like)
   - Border: `1px solid #374151`
   - Border radius: Medium (8px)
   - Padding: `1.25rem 1.5rem`
   - Max-width: `500px`
   - Margin: `3rem auto 0`
   
   - **Content:**
     ```
     pip install mcp-docs
     ```
   - Font: Monospace (Fira Code or Monaco)
   - Font size: `1rem` (16px)
   - Color: `#F9FAFB` (Light gray)
   - Display: Flex, space-between, align-center
   
   - **Copy Button:**
     - Icon: Use `Copy` icon from lucide-react on the right
     - Size: `20px x 20px`
     - Color: `#9CA3AF` (Gray-400)
     - Hover: `#FFFFFF`
     - Cursor: Pointer
     - Click action: Copy to clipboard, show toast notification

**Hero Section Spacing:**
- Padding top: `8rem` (128px)
- Padding bottom: `6rem` (96px)
- Background: Gradient from white to light gray (`#F9FAFB`)

### Features Section

**Layout:** Grid of 3 columns (desktop), 1 column (mobile)

**Content:** 3-4 feature cards highlighting key capabilities

**Card Structure:**
- Icon (top) - Use lucide-react icons
- Title (H3)
- Description (Body text)
- Border radius: Medium (8px)
- Padding: `2rem`
- Background: White
- Shadow: Medium

### Pricing Section

**Layout:** Grid of 4 pricing tiers (Free, Basic, Pro, Advanced)

**Card Structure:**
- Tier name
- Price
- Feature list
- CTA button
- Highlight current/popular tier

### Footer

**Layout:** Multi-column layout

**Content:**
- **Column 1:** Product
  - Features
  - Pricing
  - Documentation
  
- **Column 2:** Company
  - About
  - Blog
  - Careers
  
- **Column 3:** Support
  - Help Center
  - Contact
  - Status
  
- **Column 4:** Legal
  - Privacy Policy
  - Terms of Service
  - Cookie Policy

**Bottom Bar:**
- Copyright: "© 2024 MCP-Docs. All rights reserved."
- Social media links (optional, use lucide-react icons if needed: `Twitter`, `Github`, `Linkedin`)

**Styling:**
- Background: `#111827` (Dark gray)
- Text: Light gray (#9CA3AF)
- Padding: `4rem 2rem 2rem`
- Border top: None

---

## User Dashboard

### Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│              Dashboard Navigation Bar                   │
│  [Logo] [Dashboard] [Projects] [Endpoints] [Settings] │
│                    [User Menu ▼]                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Welcome    │  │ Subscription │  │   Usage      │ │
│  │   Card       │  │    Status    │  │   Stats      │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │         MCP Endpoints Section                    │ │
│  │  [+ Create New Endpoint]                         │ │
│  │                                                  │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │ │
│  │  │ Endpoint │  │ Endpoint │  │ Endpoint │     │ │
│  │  │  Card    │  │  Card    │  │  Card    │     │ │
│  │  └──────────┘  └──────────┘  └──────────┘     │ │
│  └──────────────────────────────────────────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │         Recent Activity                          │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Dashboard Navigation Bar

**Position:** Fixed at top (same as landing page nav)

**Content:**
- **Left Side:**
  - Logo: "MCP-Docs" (clickable, links to dashboard)
  - Navigation Links:
    - Dashboard (active indicator)
    - Projects
    - Endpoints
    - Settings

- **Right Side:**
  - Notifications icon (use `Bell` icon from lucide-react, optional badge)
  - User menu dropdown:
    - User name/email
    - Avatar (initials or image)
    - Dropdown:
      - Profile
      - Subscription
      - Usage
      - Settings
      - Sign Out

**Styling:**
- Same as landing page nav
- Active link: Primary color with underline

### Welcome Card

**Layout:** Horizontal card at top of dashboard

**Content:**
- Welcome message: "Welcome back, [User Name]!"
- Quick stats:
  - Total Projects: X
  - Active Endpoints: Y
  - Usage this month: Z / Limit
- Quick action: "Create New Project" button

**Styling:**
- Background: Gradient (Primary to Primary Light)
- Text: White
- Padding: `2rem`
- Border radius: Medium (8px)
- Shadow: Large

### Subscription Status Card

**Layout:** Card showing current subscription tier

**Content:**
- Current Tier: Badge (Free/Basic/Pro/Advanced)
- Status: Active/Expiring/Cancelled
- Expiry Date: "Expires on [Date]"
- Action Button: "Manage Subscription" (links to `/subscription`)

**Styling:**
- Background: White
- Border: 1px solid Border color
- Padding: `1.5rem`
- Border radius: Medium (8px)
- Shadow: Small

**Tier Badge Colors:**
- Free: Gray (#6B7280)
- Basic: Blue (#3B82F6)
- Pro: Purple (#8B5CF6)
- Advanced: Gold (#F59E0B)

### Usage Stats Card

**Layout:** Card showing current usage metrics

**Content:**
- **SSE Executions:**
  - Current: X
  - Limit: Y
  - Progress bar: X/Y
  - Percentage: Z%
- **Period:** "Current Period: [Start] - [End]"
- **Action:** "View Detailed Usage" link

**Styling:**
- Same as Subscription card
- Progress bar: Primary color
- Warning state: If usage > 80%, show amber color

### MCP Endpoints Section

**Layout:** Main content area

**Header:**
- Title: "MCP Endpoints" (H2)
- Action Button: "+ Create New Endpoint" (Primary button, right-aligned)

**Content:**
- **Empty State** (if no endpoints):
  - Icon: Use `Server` or `FileText` icon from lucide-react
  - Message: "No MCP endpoints yet"
  - Description: "Create your first endpoint to get started"
  - CTA: "Create New Endpoint" button

- **Endpoint Cards Grid** (if endpoints exist):
  - Layout: Grid, 3 columns (desktop), 1 column (mobile)
  - Card spacing: `1.5rem` gap

### Endpoint Card

**Layout:** Individual card for each MCP endpoint

**Content:**
```
┌─────────────────────────────────────┐
│  Endpoint Name (H4)          [•••]  │
│  Status: [Running/Stopped] Badge    │
│                                     │
│  Project: [Project Name]            │
│  URL: [Endpoint URL]                │
│  Port: [Port Number]                │
│                                     │
│  Last Active: [Timestamp]           │
│                                     │
│  [Start/Stop] [View Details]        │
└─────────────────────────────────────┘
```

**Specifications:**

1. **Card Header:**
   - Endpoint name: H4, bold
   - Actions menu: Use `MoreVertical` icon from lucide-react on right
     - Dropdown: Edit, Delete, View Logs

2. **Status Badge:**
   - Running: Green badge (#10B981)
   - Stopped: Gray badge (#6B7280)
   - Error: Red badge (#EF4444)
   - Starting: Amber badge (#F59E0B)

3. **Endpoint Details:**
   - Project name: Link to project
   - Endpoint URL: Monospace font, copyable
   - Port: Monospace font
   - Last Active: Relative time ("2 hours ago")

4. **Action Buttons:**
   - Start/Stop: Primary button (contextual)
   - View Details: Secondary button (outline)

**Styling:**
- Background: White
- Border: 1px solid Border color
- Padding: `1.5rem`
- Border radius: Medium (8px)
- Shadow: Small
- Hover: Slight elevation (shadow increase)

### Create New Endpoint Modal/Form

**Trigger:** Clicking "+ Create New Endpoint" button

**Layout:** Modal overlay with form

**Content:**
```
┌─────────────────────────────────────────┐
│  Create New MCP Endpoint          [×]   │
├─────────────────────────────────────────┤
│                                         │
│  Project: [Dropdown Select ▼]          │
│                                         │
│  Endpoint Name: [Text Input]           │
│  (Optional, defaults to project name)   │
│                                         │
│  Configuration:                         │
│  ☐ Enable AI Search (Advanced tier)    │
│  ☐ Enable Code Snippet Finder           │
│                                         │
│  [Cancel]  [Create Endpoint]            │
│                                         │
└─────────────────────────────────────────┘
```

**Form Fields:**

1. **Project Selection:**
   - Type: Dropdown select
   - Options: List of user's projects
   - Required: Yes
   - Placeholder: "Select a project..."

2. **Endpoint Name:**
   - Type: Text input
   - Required: No (defaults to project name)
   - Placeholder: "My MCP Endpoint"
   - Max length: 100 characters

3. **Configuration Options:**
   - Checkboxes (tier-dependent)
   - Only show if user has Advanced tier
   - Options:
     - Enable AI Search (CrewAI)
     - Enable Code Snippet Finder

**Validation:**
- Project must be selected
- Endpoint name must be unique (if provided)
- Show error messages inline

**Styling:**
- Modal overlay: Semi-transparent black (rgba(0,0,0,0.5))
- Modal container: White, centered, max-width 500px
- Border radius: Large (16px)
- Shadow: Large
- Padding: `2rem`

---

## Subscription Management Page

### Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│              Dashboard Navigation Bar                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │         Current Subscription                     │ │
│  │  [Tier Badge] [Status] [Expiry]                 │ │
│  │  [Manage Button]                                │ │
│  └──────────────────────────────────────────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │         Available Plans                          │ │
│  │  [Free] [Basic] [Pro] [Advanced]                │ │
│  │  (Highlight current tier)                        │ │
│  └──────────────────────────────────────────────────┘ │
│                                                         │
│  ┌──────────────────────────────────────────────────┐ │
│  │         Usage & Billing                          │ │
│  │  [Usage Stats] [Billing History]                 │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Current Subscription Card

**Content:**
- Tier name with badge
- Status: Active/Cancelled/Expiring
- Billing cycle: Monthly/Yearly
- Price: $X.XX/month or $Y.YY/year
- Expiry date: "Renews on [Date]" or "Expires on [Date]"
- Actions:
  - "Upgrade" button (if on lower tier)
  - "Cancel Subscription" link (if active)
  - "Reactivate" button (if cancelled)

### Available Plans Section

**Layout:** Grid of 4 pricing cards

**Card Structure:**
- Tier name
- Price (large, bold)
- Billing cycle toggle (Monthly/Yearly)
- Feature list (checkmarks)
- CTA button:
  - "Current Plan" (if selected)
  - "Upgrade" (if higher tier)
  - "Downgrade" (if lower tier)
  - "Select Plan" (if not subscribed)

**Highlighting:**
- Current tier: Border in primary color, "Current Plan" badge
- Recommended tier: "Popular" badge (optional)

### Usage & Billing Section

**Tabs:**
- Usage
- Billing History

**Usage Tab:**
- Current period usage stats
- SSE executions: Progress bar
- Projects: Count
- Indexing jobs: Count
- Visual charts (optional)

**Billing History Tab:**
- Table of invoices
- Columns: Date, Description, Amount, Status, Download
- Pagination

---

## Component Specifications

### Buttons

**Primary Button:**
- Background: Primary color (#3B82F6)
- Text: White
- Padding: `0.75rem 1.5rem`
- Border radius: Small (4px)
- Font weight: Semibold (600)
- Hover: Darker shade (#2563EB)
- Active: Even darker (#1D4ED8)
- Disabled: Gray (#9CA3AF), cursor not-allowed

**Secondary Button:**
- Background: Transparent
- Border: 1px solid Primary
- Text: Primary color
- Same padding and sizing
- Hover: Light background (#EFF6FF)

**Danger Button:**
- Background: Error color (#EF4444)
- Text: White
- Same styling as primary

### Input Fields

**Text Input:**
- Border: 1px solid Border color (#E5E7EB)
- Border radius: Small (4px)
- Padding: `0.75rem 1rem`
- Font size: Body (16px)
- Focus: Border color Primary, outline ring
- Error state: Red border (#EF4444)
- Placeholder: Text Muted color

**Dropdown Select:**
- Same styling as text input
- Dropdown arrow icon on right
- Custom styled dropdown menu

### Cards

**Standard Card:**
- Background: White
- Border: 1px solid Border color
- Border radius: Medium (8px)
- Padding: `1.5rem`
- Shadow: Small
- Hover: Shadow increase (optional)

### Badges

**Status Badge:**
- Display: Inline-block
- Padding: `0.25rem 0.75rem`
- Border radius: Full (9999px)
- Font size: Small (14px)
- Font weight: Medium (500)

**Tier Badge:**
- Same as status badge
- Background: Tier-specific color
- Text: White

### Modals

**Modal Overlay:**
- Background: Semi-transparent black (rgba(0,0,0,0.5))
- Position: Fixed, full screen
- Z-index: High (1000+)
- Backdrop blur: Optional

**Modal Container:**
- Background: White
- Border radius: Large (16px)
- Max-width: 500px (forms) or 800px (content)
- Centered (vertical and horizontal)
- Shadow: Large
- Padding: `2rem`

### Loading States

**Spinner:**
- Circular spinner animation
- Primary color
- Size: 24px, 32px, 48px (context-dependent)

**Skeleton Loader:**
- Gray background with shimmer animation
- Match content shape

### Toast Notifications

**Success Toast:**
- Background: Success color (#10B981)
- Text: White
- Icon: Use `CheckCircle` icon from lucide-react
- Position: Top-right
- Auto-dismiss: 3-5 seconds

**Error Toast:**
- Background: Error color (#EF4444)
- Text: White
- Icon: Use `XCircle` or `AlertCircle` icon from lucide-react
- Same positioning

---

## Responsive Design

### Breakpoints

- **Mobile:** < 640px
- **Tablet:** 640px - 1024px
- **Desktop:** > 1024px

### Mobile Adaptations

**Landing Page:**
- Navigation: Hamburger menu
- Hero: Full width, reduced padding
- Features: Single column
- Pricing: Single column, stacked

**Dashboard:**
- Navigation: Hamburger menu
- Cards: Single column
- Endpoint cards: Full width
- Modals: Full screen on mobile

**Grid Layouts:**
- Desktop: 3-4 columns
- Tablet: 2 columns
- Mobile: 1 column

---

## User Flows

### Landing Page → Sign Up

1. User visits landing page
2. Clicks "Try Free" button
3. Redirected to sign up page
4. After sign up, redirected to dashboard

### Dashboard → Create Endpoint

1. User on dashboard
2. Clicks "+ Create New Endpoint"
3. Modal opens
4. Selects project
5. (Optional) Enters endpoint name
6. Clicks "Create Endpoint"
7. Endpoint created, card appears in list
8. Status shows "Starting" then "Running"

### Dashboard → Manage Subscription

1. User clicks "Manage Subscription" in status card
2. Redirected to `/subscription` page
3. Views current plan
4. Clicks "Upgrade" on desired tier
5. Redirected to payment (Stripe Checkout)
6. After payment, redirected back with updated tier

---

## API Integration Points

### Landing Page

- No API calls required (static content)
- "Try Free" button: Link to `/auth/register`

### Dashboard

**On Load:**
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/projects` - Get user's projects
- `GET /api/v1/mcp/instances` - Get MCP endpoints
- `GET /api/v1/subscription` - Get subscription info
- `GET /api/v1/usage/current` - Get usage stats

**Create Endpoint:**
- `POST /api/v1/projects/:id/mcp/start` - Create and start endpoint

**Endpoint Actions:**
- `POST /api/v1/projects/:id/mcp/stop` - Stop endpoint
- `GET /api/v1/projects/:id/mcp/status` - Get status (polling or WebSocket)

### Subscription Page

**On Load:**
- `GET /api/v1/subscription` - Get current subscription
- `GET /api/v1/usage/current` - Get usage
- `GET /api/v1/usage/history` - Get billing history

**Actions:**
- `POST /api/v1/subscription/upgrade` - Upgrade tier
- `POST /api/v1/subscription/cancel` - Cancel subscription

### WebSocket Events

**Subscribe to:**
- `project:updated` - Project status changes
- `mcp:status` - MCP endpoint status updates
- `job:progress` - Indexing progress (if applicable)

---

## Additional Notes

### Accessibility

- All interactive elements must be keyboard accessible
- Proper ARIA labels for screen readers
- Color contrast ratios meet WCAG AA standards
- Focus indicators visible on all focusable elements

### Performance

- **Next.js Optimizations:**
  - Use Next.js Image component for optimized images
  - Implement code splitting with dynamic imports
  - Use Server Components to reduce client bundle size
  - Leverage Next.js built-in optimizations (automatic code splitting, etc.)
- **General:**
  - Lazy load heavy components
  - Optimize bundle size
  - Use React.memo for expensive components
  - Implement proper loading states

### Browser Support

- Chrome (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Edge (latest 2 versions)

---

## shadcn/ui Component Usage

### Recommended Components

Use shadcn/ui components where applicable:

- **Button** - For all button variants (primary, secondary, outline, ghost, etc.)
- **Card** - For all card containers
- **Input** - For text inputs
- **Select** - For dropdown selects
- **Dialog** - For modals
- **Dropdown Menu** - For action menus
- **Badge** - For status and tier badges
- **Toast** - For notifications (use sonner or react-hot-toast with shadcn/ui)
- **Progress** - For progress bars
- **Tabs** - For tabbed interfaces
- **Avatar** - For user avatars
- **Separator** - For dividers
- **Skeleton** - For loading states

### Icon Library

**Use lucide-react for all icons:**
- Install: `npm install lucide-react`
- Import: `import { IconName } from 'lucide-react'`
- Common icons needed:
  - `Copy` - Copy to clipboard
  - `Bell` - Notifications
  - `Server` - MCP endpoints
  - `FileText` - Documents
  - `MoreVertical` - Action menu
  - `CheckCircle` - Success
  - `XCircle` - Error
  - `AlertCircle` - Warning
  - `ChevronDown` - Dropdowns
  - `Plus` - Add/create actions
  - `Settings` - Settings
  - `User` - User profile

### Theme Configuration

- Use Tailwind CSS with shadcn/ui's default theme
- Customize colors in `tailwind.config.js` to match the design system
- No dark mode support needed (light mode only)

## Design Assets Needed

1. Logo (SVG preferred, or text-based logo)
2. Favicon
3. Social media images (Open Graph) - optional

**Note:** Icons will be provided by lucide-react library, no custom icon set needed.

---

**Document Version:** 1.0  
**Last Updated:** 2024  
**Status:** Ready for Implementation

