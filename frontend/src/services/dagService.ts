import { api } from './api';
import type { 
  DAG, 
  CreateDAGRequest,
  DAGSummary,
  Node
} from '../types/dag';

export class DAGService {
  private static readonly BASE_PATH = '/v1/dags';

  /**
   * Get list of all DAGs
   */
  static async listDAGs(): Promise<DAGSummary[]> {
    const response = await api.get<{dags: string[]; count: number}>(this.BASE_PATH);
    
    // The backend returns {dags: [uuid...], count: number}
    // For now, we'll create mock DAG summaries from the UUIDs
    // TODO: Update when backend provides full DAG summary endpoint
    if (response.data && Array.isArray(response.data.dags)) {
      return response.data.dags.map((dagId: string) => ({
        id: dagId,
        name: `DAG ${dagId.slice(0, 8)}`, // Use first 8 chars of UUID as name
        description: 'Legal case decision tree',
        node_count: 0, // Unknown for now
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }));
    }
    
    return [];
  }

  /**
   * Get a specific DAG by ID
   */
  static async getDAG(dagId: string): Promise<DAG> {
    interface BackendDAG {
      id: string;
      title: string;
      nodes: Array<{
        id: string;
        question: string;
        answers: Array<{
          id: string;
          answer: string;
          next_node?: string;
          is_terminal?: boolean;
          meta?: Record<string, unknown>;
        }>;
      }>;
    }

    const response = await api.get<BackendDAG>(`${this.BASE_PATH}/${dagId}`);
    const backendDAG = response.data;

    // Transform backend array structure to frontend object structure
    const nodesMap: Record<string, Node> = {};
    let rootNode = '';

    backendDAG.nodes.forEach((node, index) => {
      // First node is typically the root node
      if (index === 0) {
        rootNode = node.id;
      }
      
      nodesMap[node.id] = {
        id: node.id,
        question: node.question,
        answers: node.answers.map(answer => ({
          id: answer.id,
          answer: answer.answer,
          next_node: answer.next_node || undefined,
          is_terminal: !answer.next_node,
          meta: answer.meta || {},
        })),
        description: undefined, // Backend doesn't provide this
        meta: {},
      };
    });

    // Transform to expected frontend structure
    const transformedDAG: DAG = {
      id: backendDAG.id,
      name: backendDAG.title,
      description: 'Legal case decision tree',
      nodes: nodesMap,
      root_node: rootNode,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      meta: {},
    };

    return transformedDAG;
  }

  /**
   * Update an existing DAG
   */
  static async updateDAG(dagId: string, dag: DAG): Promise<DAG> {
    const response = await api.put<DAG>(`${this.BASE_PATH}/${dagId}`, dag);
    return response.data;
  }

  /**
   * Validate a DAG structure
   */
  static async validateDAG(dag: DAG): Promise<{ valid: boolean; errors?: string[] }> {
    const response = await api.post<{ valid: boolean; errors?: string[] }>(
      `${this.BASE_PATH}/validate`, 
      dag
    );
    return response.data;
  }

  /**
   * Create a new DAG (if endpoint exists in future)
   */
  static async createDAG(request: CreateDAGRequest): Promise<DAG> {
    const response = await api.post<DAG>(this.BASE_PATH, request);
    return response.data;
  }

  /**
   * Delete a DAG (if endpoint exists in future)
   */
  static async deleteDAG(dagId: string): Promise<void> {
    await api.delete(`${this.BASE_PATH}/${dagId}`);
  }
}

// React Query hooks for better caching and state management
export const dagQueryKeys = {
  all: ['dags'] as const,
  lists: () => [...dagQueryKeys.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...dagQueryKeys.lists(), { filters }] as const,
  details: () => [...dagQueryKeys.all, 'detail'] as const,
  detail: (id: string) => [...dagQueryKeys.details(), id] as const,
} as const;
