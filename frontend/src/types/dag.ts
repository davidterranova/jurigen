import { z } from 'zod'
import type { Node as FlowNode, Edge as FlowEdge } from '@xyflow/react'

// Basic DAG schemas matching the backend
export const AnswerSchema = z.object({
  id: z.string(),
  answer: z.string(),
  next_node: z.string().optional(),
  is_terminal: z.boolean().optional(),
  meta: z.record(z.string(), z.unknown()).optional(),
})

export const NodeSchema = z.object({
  id: z.string(),
  question: z.string(),
  description: z.string().optional(),
  answers: z.array(AnswerSchema),
  meta: z.record(z.string(), z.unknown()).optional(),
})

export const DAGSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  nodes: z.record(z.string(), NodeSchema),
  root_node: z.string(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
  meta: z.record(z.string(), z.unknown()).optional(),
})

export const DAGSummarySchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  node_count: z.number().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
})

// Derived types
export type Answer = z.infer<typeof AnswerSchema>
export type Node = z.infer<typeof NodeSchema>
export type DAG = z.infer<typeof DAGSchema>
export type DAGSummary = z.infer<typeof DAGSummarySchema>

// React Flow specific types
export interface DAGNode extends FlowNode {
  data: {
    question: string;
    description?: string;
    answers: Answer[];
    meta?: Record<string, unknown>;
    isRoot?: boolean;
  }
}

export interface TerminalNode extends FlowNode {
  data: {
    answer: string;
    sourceNodeId: string;
    meta?: Record<string, unknown>;
  }
}

export interface DAGEdge extends FlowEdge {
  data: {
    answer: string;
    is_terminal?: boolean;
    meta?: Record<string, unknown>;
  }
}

// API request/response types
export interface CreateDAGRequest {
  name: string;
  description?: string;
  nodes: Record<string, Node>;
  root_node: string;
}

export interface UpdateDAGRequest extends CreateDAGRequest {
  id: string;
}

export interface DAGResponse {
  dag: DAG;
}

export interface DAGListResponse {
  dags: DAGSummary[];
}

// Traversal state types
export interface DAGTraversalState {
  currentNodeId: string | null;
  visitedNodes: string[];
  selectedAnswers: Record<string, Answer>;
  isComplete: boolean;
  context: string[];
}

// UI Selection types
export type SelectionType = 'node' | 'edge' | null;

export interface NodeSelection {
  type: 'node';
  nodeId: string;
  node: Node;
}

export interface EdgeSelection {
  type: 'edge';
  edgeId: string;
  sourceNodeId: string;
  targetNodeId: string | null;
  answer: Answer;
}

export type TerminalSelection = {
  type: 'terminal';
  terminalId: string;
  sourceNodeId: string;
  answer: Answer;
};

export type Selection = NodeSelection | EdgeSelection | TerminalSelection | null;
