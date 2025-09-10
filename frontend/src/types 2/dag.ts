// TypeScript interfaces matching the Go backend DAG structures

export interface DAG {
  id: string;
  title: string;
  nodes: Record<string, Node>;
}

export interface Node {
  id: string;
  question: string;
  answers: Answer[];
}

export interface Answer {
  id: string;
  answer: string;
  next_node?: string;
  user_context?: string;
  metadata?: Record<string, unknown>;
}

// Frontend-specific types for state management

export interface DAGTraversalState {
  currentNodeId: string | null;
  visitedNodes: string[];
  selectedAnswers: Record<string, Answer>;
  isComplete: boolean;
  context: string[];
}

export interface DAGResponse {
  dag: DAG;
  success: boolean;
  message?: string;
}

export interface DAGListResponse {
  dags: DAGSummary[];
  total: number;
}

export interface DAGSummary {
  id: string;
  title: string;
  nodeCount: number;
  createdAt?: string;
  updatedAt?: string;
}

// API request/response types
export interface CreateDAGRequest {
  title: string;
}

export interface UpdateDAGRequest {
  dag: DAG;
}

export interface TraversalRequest {
  dagId: string;
  answerId: string;
}
