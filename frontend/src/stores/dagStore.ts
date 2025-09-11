import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { DAG, DAGSummary, Selection, DAGNode, DAGEdge, TerminalNode, Answer } from '../types/dag';
import type { Viewport } from '@xyflow/react';

interface DAGStore {
  // DAG data
  availableDAGs: DAGSummary[];
  selectedDAGId: string | null;
  selectedDAG: DAG | null;
  
  // Graph visualization
  nodes: (DAGNode | TerminalNode)[];
  edges: DAGEdge[];
  
  // Layout management
  layoutMode: 'auto' | 'manual';
  customPositions: Record<string, { x: number; y: number }>;
  hasManualChanges: boolean;
  
  // UI state
  selection: Selection;
  isDetailsOpen: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Viewport control
  viewport: Viewport;
  
  // Actions - DAG management
  setAvailableDAGs: (dags: DAGSummary[]) => void;
  selectDAG: (dagId: string) => void;
  setSelectedDAG: (dag: DAG) => void;
  clearDAG: () => void;
  
  // Actions - Graph visualization  
  setNodes: (nodes: (DAGNode | TerminalNode)[]) => void;
  setEdges: (edges: DAGEdge[]) => void;
  updateNode: (nodeId: string, updates: Partial<DAGNode>) => void;
  updateEdge: (edgeId: string, updates: Partial<DAGEdge>) => void;
  
  // Actions - Selection and UI
  selectNode: (nodeId: string) => void;
  selectEdge: (edgeId: string, sourceNodeId: string, targetNodeId: string | null) => void;
  selectTerminal: (terminalId: string, sourceNodeId: string, answer: Answer) => void;
  clearSelection: () => void;
  setDetailsOpen: (open: boolean) => void;
  
  // Actions - State management
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Actions - Layout management
  setLayoutMode: (mode: 'auto' | 'manual') => void;
  updateCustomPosition: (nodeId: string, position: { x: number; y: number }) => void;
  resetToAutoLayout: () => void;
  clearCustomPositions: () => void;

  // Actions - Viewport
  setViewport: (viewport: Viewport) => void;
  fitView: () => void;
  resetView: () => void;
  
  // Getters
  getNode: (nodeId: string) => DAGNode | undefined;
  getEdge: (edgeId: string) => DAGEdge | undefined;
  getSelectedNode: () => DAGNode | null;
  getSelectedEdge: () => DAGEdge | null;
}

const initialViewport: Viewport = {
  x: 0,
  y: 0,
  zoom: 1,
};

export const useDAGStore = create<DAGStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      availableDAGs: [],
      selectedDAGId: null,
      selectedDAG: null,
      nodes: [],
      edges: [],
      layoutMode: 'auto',
      customPositions: {},
      hasManualChanges: false,
      selection: null,
      isDetailsOpen: false,
      isLoading: false,
      error: null,
      viewport: initialViewport,

      // DAG management actions
      setAvailableDAGs: (dags) =>
        set({ availableDAGs: dags }, false, 'setAvailableDAGs'),

      selectDAG: (dagId) =>
        set({ selectedDAGId: dagId, selection: null, isDetailsOpen: false }, false, 'selectDAG'),

      setSelectedDAG: (dag) => {
        set(
          {
            selectedDAG: dag,
            selectedDAGId: dag.id,
            error: null,
          },
          false,
          'setSelectedDAG'
        );
      },

      clearDAG: () =>
        set(
          {
            selectedDAGId: null,
            selectedDAG: null,
            nodes: [],
            edges: [],
            selection: null,
            isDetailsOpen: false,
          },
          false,
          'clearDAG'
        ),

      // Graph visualization actions
      setNodes: (nodes) => set({ nodes }, false, 'setNodes'),
      
      setEdges: (edges) => set({ edges }, false, 'setEdges'),

      updateNode: (nodeId, updates) =>
        set(
          (state) => ({
            nodes: state.nodes.map(node =>
              node.id === nodeId && node.type === 'dagNode' 
                ? { ...node, ...updates } as DAGNode 
                : node
            ),
          }),
          false,
          'updateNode'
        ),

      updateEdge: (edgeId, updates) =>
        set(
          (state) => ({
            edges: state.edges.map(edge =>
              edge.id === edgeId ? { ...edge, ...updates } : edge
            ),
          }),
          false,
          'updateEdge'
        ),

      // Selection and UI actions
      selectNode: (nodeId) => {
        const { selectedDAG } = get();
        if (!selectedDAG || !selectedDAG.nodes[nodeId]) return;

        const node = selectedDAG.nodes[nodeId];
        set(
          {
            selection: {
              type: 'node',
              nodeId,
              node,
            },
            isDetailsOpen: true,
          },
          false,
          'selectNode'
        );
      },

      selectEdge: (edgeId, sourceNodeId, targetNodeId) => {
        const { selectedDAG } = get();
        if (!selectedDAG) return;

        const sourceNode = selectedDAG.nodes[sourceNodeId];
        if (!sourceNode) return;

        // Find the answer that corresponds to this edge
        const answerId = edgeId.replace(`${sourceNodeId}-`, '');
        const answer = sourceNode.answers.find((a) => a.id === answerId);
        if (!answer) return;

        set(
          {
            selection: {
              type: 'edge',
              edgeId,
              sourceNodeId,
              targetNodeId,
              answer,
            },
            isDetailsOpen: true,
          },
          false,
          'selectEdge'
        );
      },

      selectTerminal: (terminalId, sourceNodeId, answer) => {
        set(
          {
            selection: {
              type: 'terminal',
              terminalId,
              sourceNodeId,
              answer,
            },
            isDetailsOpen: true,
          },
          false,
          'selectTerminal'
        );
      },

      clearSelection: () =>
        set(
          { selection: null, isDetailsOpen: false },
          false,
          'clearSelection'
        ),

      setDetailsOpen: (open) =>
        set({ isDetailsOpen: open }, false, 'setDetailsOpen'),

      // State management actions
      setLoading: (loading) => set({ isLoading: loading }, false, 'setLoading'),
      
      setError: (error) => set({ error }, false, 'setError'),

      // Layout management actions
      setLayoutMode: (mode) => 
        set({ layoutMode: mode }, false, 'setLayoutMode'),
      
      updateCustomPosition: (nodeId, position) =>
        set(
          (state) => ({
            customPositions: {
              ...state.customPositions,
              [nodeId]: position,
            },
            hasManualChanges: true,
          }),
          false,
          'updateCustomPosition'
        ),

      resetToAutoLayout: () =>
        set(
          {
            layoutMode: 'auto',
            customPositions: {},
            hasManualChanges: false,
          },
          false,
          'resetToAutoLayout'
        ),

      clearCustomPositions: () =>
        set(
          {
            customPositions: {},
            hasManualChanges: false,
          },
          false,
          'clearCustomPositions'
        ),

      // Viewport actions
      setViewport: (viewport) => set({ viewport }, false, 'setViewport'),
      
      fitView: () => {
        // This will be handled by the React Flow instance
        // We just track the intent here
        set({ viewport: { ...get().viewport } }, false, 'fitView');
      },
      
      resetView: () =>
        set({ viewport: initialViewport }, false, 'resetView'),

      // Getters
      getNode: (nodeId) => get().nodes.find(node => node.id === nodeId),
      
      getEdge: (edgeId) => get().edges.find(edge => edge.id === edgeId),
      
      getSelectedNode: () => {
        const { selection, nodes } = get();
        if (selection?.type !== 'node') return null;
        return nodes.find(node => node.id === selection.nodeId) || null;
      },
      
      getSelectedEdge: () => {
        const { selection, edges } = get();
        if (selection?.type !== 'edge') return null;
        return edges.find(edge => edge.id === selection.edgeId) || null;
      },
    }),
    {
      name: 'dag-store',
    }
  )
);
