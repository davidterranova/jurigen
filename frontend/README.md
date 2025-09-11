# Jurigen Frontend

A modern React + TypeScript frontend for the Jurigen legal helper application, featuring interactive DAG (Directed Acyclic Graph) visualization using React Flow.

## 🚀 Features

- **Interactive DAG Visualization**: Explore legal decision DAGs with pan, zoom, and click interactions
- **Modern UI**: Clean, professional interface built with Tailwind CSS
- **Type Safety**: Full TypeScript implementation with strict type checking
- **State Management**: Zustand for lightweight, scalable state management
- **Responsive Design**: Works beautifully on all device sizes
- **Accessibility**: ARIA labels and keyboard navigation support

## 📋 Requirements

- Node.js 16+ 
- npm 7+
- Backend API server running on port 8080

## 🛠 Installation & Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. Visit [http://localhost:3000](http://localhost:3000)

## 📜 Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production  
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run test` - Run tests in watch mode
- `npm run test:run` - Run tests once
- `npm run coverage` - Generate test coverage report

## 🏗 Architecture

### Component Structure
```
src/
├── components/           # React components
│   ├── DagList.tsx      # Sidebar DAG selector
│   ├── GraphView.tsx    # React Flow graph canvas
│   ├── DetailsPanel.tsx # Node/edge details display
│   └── Layout.tsx       # App shell layout
├── stores/              # Zustand state stores
│   └── dagStore.ts      # DAG visualization state
├── services/            # API integration
│   ├── api.ts          # HTTP client
│   └── dagService.ts   # DAG-specific endpoints
├── types/              # TypeScript type definitions
│   ├── dag.ts         # DAG domain types
│   └── api.ts         # API response types
└── pages/             # Route components
    └── Home.tsx       # Main application page
```

### Key Technologies

- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **React Flow** - Graph visualization
- **Tailwind CSS** - Styling
- **Zustand** - State management
- **Vitest** - Testing framework

## 🎯 Usage

### Basic Navigation
1. **Select a DAG**: Choose from the sidebar list
2. **Explore the Graph**: Pan, zoom, and click nodes/edges
3. **View Details**: Click any node or edge to see detailed information
4. **Keyboard Shortcuts**: 
   - `F` - Fit graph to view
   - `Esc` - Clear selection

### Graph Interactions
- **Pan**: Click and drag empty canvas
- **Zoom**: Mouse wheel or zoom controls
- **Select**: Click nodes or edges to view details
- **Mini-map**: Navigate large graphs using the mini-map

## 🔧 Configuration

### Environment Variables
- `VITE_API_BASE_URL` - Backend API URL (default: http://localhost:8080)

### API Integration
The frontend expects these backend endpoints:
- `GET /v1/dags` - List all DAGs
- `GET /v1/dags/{id}` - Get specific DAG
- `PUT /v1/dags/{id}` - Update DAG
- `POST /v1/dags/validate` - Validate DAG

## 🧪 Testing

Tests are written using Vitest and React Testing Library:

```bash
# Run all tests
npm run test

# Run tests with coverage
npm run coverage

# Run specific test file
npm run test DagList.test.tsx
```

## 🚨 Troubleshooting

### Common Issues

1. **API Connection Issues**
   - Ensure backend is running on port 8080
   - Check VITE_API_BASE_URL configuration

2. **Build Errors**
   - Clear node_modules and reinstall: `rm -rf node_modules && npm install`
   - Check TypeScript version compatibility

3. **Graph Not Rendering**
   - Verify DAG data structure matches expected format
   - Check browser console for React Flow errors

## 🤝 Contributing

1. Follow existing code style and patterns
2. Add tests for new features
3. Update type definitions for API changes
4. Run linting before committing: `npm run lint`

## 📝 License

This project is part of the Jurigen legal helper application.