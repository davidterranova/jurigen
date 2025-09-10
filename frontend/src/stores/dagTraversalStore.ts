import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { DAG, Node, Answer, DAGTraversalState } from '../types/dag';

interface DAGTraversalStore extends DAGTraversalState {
  // State
  currentDAG: DAG | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  setDAG: (dag: DAG) => void;
  startTraversal: (rootNodeId: string) => void;
  selectAnswer: (nodeId: string, answer: Answer) => void;
  goToNode: (nodeId: string) => void;
  goBack: () => void;
  reset: () => void;
  setError: (error: string | null) => void;
  setLoading: (loading: boolean) => void;

  // Getters
  getCurrentNode: () => Node | null;
  getNextNode: (answerId: string) => Node | null;
  isAtLeafNode: () => boolean;
  getCompletedContext: () => string[];
}

const initialState: DAGTraversalState = {
  currentNodeId: null,
  visitedNodes: [],
  selectedAnswers: {},
  isComplete: false,
  context: [],
};

export const useDAGTraversalStore = create<DAGTraversalStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      ...initialState,
      currentDAG: null,
      isLoading: false,
      error: null,

      // Actions
      setDAG: (dag) =>
        set(
          (state) => ({
            ...state,
            currentDAG: dag,
            error: null,
          }),
          false,
          'setDAG'
        ),

      startTraversal: (rootNodeId) =>
        set(
          (state) => ({
            ...state,
            currentNodeId: rootNodeId,
            visitedNodes: [rootNodeId],
            selectedAnswers: {},
            isComplete: false,
            context: [],
            error: null,
          }),
          false,
          'startTraversal'
        ),

      selectAnswer: (nodeId, answer) => {
        const state = get();
        const nextNodeId = answer.next_node;

        set(
          {
            ...state,
            selectedAnswers: {
              ...state.selectedAnswers,
              [nodeId]: answer,
            },
            currentNodeId: nextNodeId || null,
            visitedNodes: nextNodeId 
              ? [...state.visitedNodes, nextNodeId]
              : state.visitedNodes,
            context: [...state.context, answer.answer],
            isComplete: !nextNodeId, // If no next node, we're at a leaf
          },
          false,
          'selectAnswer'
        );
      },

      goToNode: (nodeId) =>
        set(
          (state) => ({
            ...state,
            currentNodeId: nodeId,
            visitedNodes: [...state.visitedNodes, nodeId],
          }),
          false,
          'goToNode'
        ),

      goBack: () => {
        const state = get();
        const visitedNodes = [...state.visitedNodes];
        visitedNodes.pop(); // Remove current node
        const previousNodeId = visitedNodes[visitedNodes.length - 1] || null;

        set(
          {
            ...state,
            currentNodeId: previousNodeId,
            visitedNodes,
          },
          false,
          'goBack'
        );
      },

      reset: () =>
        set(
          {
            ...initialState,
            currentDAG: get().currentDAG, // Keep the current DAG
            isLoading: false,
            error: null,
          },
          false,
          'reset'
        ),

      setError: (error) =>
        set(
          (state) => ({
            ...state,
            error,
          }),
          false,
          'setError'
        ),

      setLoading: (loading) =>
        set(
          (state) => ({
            ...state,
            isLoading: loading,
          }),
          false,
          'setLoading'
        ),

      // Getters
      getCurrentNode: () => {
        const { currentDAG, currentNodeId } = get();
        if (!currentDAG || !currentNodeId) return null;
        return currentDAG.nodes[currentNodeId] || null;
      },

      getNextNode: (answerId) => {
        const { currentDAG, currentNodeId } = get();
        if (!currentDAG || !currentNodeId) return null;
        
        const currentNode = currentDAG.nodes[currentNodeId];
        const answer = currentNode?.answers.find(a => a.id === answerId);
        
        if (!answer?.next_node) return null;
        return currentDAG.nodes[answer.next_node] || null;
      },

      isAtLeafNode: () => {
        const currentNode = get().getCurrentNode();
        return currentNode ? currentNode.answers.length === 0 : false;
      },

      getCompletedContext: () => {
        const { context, isComplete } = get();
        return isComplete ? context : [];
      },
    }),
    {
      name: 'dag-traversal-store',
    }
  )
);
