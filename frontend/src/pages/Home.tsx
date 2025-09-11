import { useRef, useEffect } from 'react';
import { ReactFlowProvider } from '@xyflow/react';
import { useDAGStore } from '../stores/dagStore';
import DagList from '../components/DagList';
import GraphView from '../components/GraphView';
import type { GraphViewRef } from '../components/GraphView';
import DetailsPanel from '../components/DetailsPanel';

const Home = () => {
  const graphRef = useRef<GraphViewRef>(null);
  const { 
    selectedDAGId, 
    isDetailsOpen, 
    layoutMode, 
    hasManualChanges, 
    clearSelection, 
    setLayoutMode,
    resetToAutoLayout 
  } = useDAGStore();

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // F key - fit view
      if (event.key === 'f' || event.key === 'F') {
        event.preventDefault();
        graphRef.current?.fitView();
      }
      
      // Escape key - clear selection
      if (event.key === 'Escape') {
        event.preventDefault();
        clearSelection();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [clearSelection]);

  const handleFitView = () => {
    graphRef.current?.fitView();
  };

  const handleResetZoom = () => {
    graphRef.current?.resetView();
  };

  const handleLayoutModeChange = (mode: 'auto' | 'manual') => {
    setLayoutMode(mode);
    graphRef.current?.setLayoutMode(mode);
  };

  const handleResetToAutoLayout = () => {
    resetToAutoLayout();
    graphRef.current?.resetToAutoLayout();
    graphRef.current?.applyAutoLayout();
  };

  const handleApplyAutoLayout = () => {
    graphRef.current?.applyAutoLayout();
  };

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Left Sidebar - DAG List */}
      <div className="w-80 flex-shrink-0">
        <div className="h-full flex flex-col">
          <DagList className="flex-1" />
          
          {/* Graph Controls */}
          {selectedDAGId && (
            <div className="p-4 border-t border-gray-200 bg-white">
              {/* Layout Mode Controls */}
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-900 mb-3">Layout Mode</h4>
                <div className="space-y-2">
                  <div className="flex items-center">
                    <input
                      id="auto-layout"
                      name="layout-mode"
                      type="radio"
                      checked={layoutMode === 'auto'}
                      onChange={() => handleLayoutModeChange('auto')}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                    />
                    <label htmlFor="auto-layout" className="ml-2 text-sm text-gray-700">
                      Auto Layout
                    </label>
                  </div>
                  <div className="flex items-center">
                    <input
                      id="manual-layout"
                      name="layout-mode"
                      type="radio"
                      checked={layoutMode === 'manual'}
                      onChange={() => handleLayoutModeChange('manual')}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300"
                    />
                    <label htmlFor="manual-layout" className="ml-2 text-sm text-gray-700">
                      Manual Layout
                      {hasManualChanges && (
                        <span className="ml-1 inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          Modified
                        </span>
                      )}
                    </label>
                  </div>
                </div>
              </div>

              {/* Layout Action Buttons */}
              <div className="space-y-2 mb-4">
                {layoutMode === 'manual' && (
                  <button
                    onClick={handleApplyAutoLayout}
                    className="w-full px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    Apply Auto Layout
                  </button>
                )}
                {hasManualChanges && (
                  <button
                    onClick={handleResetToAutoLayout}
                    className="w-full px-3 py-2 text-sm font-medium text-orange-700 bg-orange-50 border border-orange-300 rounded-md hover:bg-orange-100 focus:outline-none focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                  >
                    Reset to Auto Layout
                  </button>
                )}
              </div>

              {/* View Controls */}
              <div className="space-y-2">
                <button
                  onClick={handleFitView}
                  className="w-full px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  Fit to View (F)
                </button>
                <button
                  onClick={handleResetZoom}
                  className="w-full px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  Reset Zoom
                </button>
              </div>

              {/* Help Text */}
              <div className="mt-3 text-xs text-gray-500">
                <div>Press <kbd className="px-1 bg-gray-100 rounded">F</kbd> to fit view</div>
                <div>Press <kbd className="px-1 bg-gray-100 rounded">Esc</kbd> to clear selection</div>
                {layoutMode === 'manual' && (
                  <div className="mt-1 font-medium text-blue-600">
                    💡 Drag nodes to reposition them
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col">
        {selectedDAGId ? (
          <ReactFlowProvider>
            <div className="flex-1 flex">
              {/* Graph Canvas */}
              <div className={`flex-1 ${isDetailsOpen ? 'mr-0' : ''}`}>
                <GraphView ref={graphRef} className="w-full h-full" />
              </div>
              
              {/* Right Panel - Details */}
              {isDetailsOpen && (
                <div className="w-96 flex-shrink-0">
                  <DetailsPanel className="h-full" />
                </div>
              )}
            </div>
          </ReactFlowProvider>
        ) : (
          // Empty State
          <div className="flex-1 flex items-center justify-center bg-gray-50">
            <div className="text-center">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h2a2 2 0 01-2-2z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No DAG selected</h3>
              <p className="mt-1 text-sm text-gray-500">
                Select a DAG from the sidebar to start exploring.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;