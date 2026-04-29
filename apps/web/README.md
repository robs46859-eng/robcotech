# Papabase Web Frontend

Next.js-based frontend for the Papabase Family Studio Suite.

## Features

- **Onboarding Questionnaire** - Interactive signup flow to understand user needs
- **Dashboard** - Overview with stats and quick actions
- **Leads Management** - CRM with AI lead scoring and insights
- **Tasks** - Task management with AI generation from notes
- **Dad AI Website Builder** - Generate websites from prompts
- **Content Studio** - AI content generation (blog, SEO, social media)
- **Proposal Builder** - Create professional proposals and quotes
- **Pricing & Billing** - Subscription management

## Tech Stack

- **Framework**: Next.js 14
- **Styling**: Tailwind CSS
- **State**: Zustand
- **Animations**: Framer Motion
- **Icons**: Lucide React
- **Notifications**: React Hot Toast

## Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Pink | `#EC89A3` | Primary actions, accents |
| Green | `#316844` | Success, tasks |
| Teal | `#3DC2B9` | Dad AI, websites |
| Purple | `#643277` | Proposals, navigation |

## Getting Started

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## Pages

- `/` - Auth/Onboarding/Dashboard (based on user state)
- `/dashboard` - Main dashboard
- `/onboarding` - User questionnaire

## API Integration

The frontend connects to the Papabase backend API at `/api/v1`:

- CRM: `/leads`, `/tasks`
- AI: `/ai/generate`, `/ai/leads/score`, `/ai/tasks/generate`, `/ai/content/*`, `/ai/proposals/*`
- Billing: `/pricing/plans`, `/billing/usage`

## Environment Variables

```bash
NEXT_PUBLIC_API_URL=/api/v1
NEXT_PUBLIC_PAPABASE_URL=http://localhost:8087
```
