# Jurigen Frontend

A minimal React TypeScript frontend application built with Vite, featuring schema validation with Zod and styling with Tailwind CSS.

## Tech Stack

- **React 19** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Zod** - Schema validation
- **React Router** - Client-side routing
- **Tailwind CSS** - Utility-first styling

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the development server:
   ```bash
   npm run dev
   ```

3. Open your browser to `http://localhost:5173`

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## Project Structure

```
src/
├── components/     # Reusable UI components
├── pages/          # Page components for routing
├── types/          # TypeScript types and Zod schemas
├── App.tsx         # Main application component
└── main.tsx        # Application entry point
```

## Features Demonstrated

- **Zod Schema Validation**: Form validation with real-time error handling
- **TypeScript Integration**: Full type safety throughout the application
- **React Router**: Single-page application routing
- **Tailwind CSS**: Modern utility-first styling
- **Component Architecture**: Clean separation of concerns

## Expanding the Application

This is a minimal setup designed to be expanded. Consider adding:

- State management (Zustand, Redux Toolkit)
- Data fetching (TanStack Query, SWR)
- Testing (Vitest, Testing Library)
- UI components library (Radix UI, Headless UI)
- Form handling (React Hook Form)