import { api } from './api';
import { 
  DAG, 
  DAGResponse, 
  DAGListResponse, 
  CreateDAGRequest, 
  UpdateDAGRequest,
  DAGSummary
} from '../types/dag';
import { ApiResponse } from '../types/api';

export class DAGService {
  private static readonly BASE_PATH = '/v1/dags';

  /**
   * Get list of all DAGs
   */
  static async listDAGs(): Promise<DAGSummary[]> {
    const response = await api.get<DAGSummary[]>(this.BASE_PATH);
    return response.data;
  }

  /**
   * Get a specific DAG by ID
   */
  static async getDAG(dagId: string): Promise<DAG> {
    const response = await api.get<DAG>(`${this.BASE_PATH}/${dagId}`);
    return response.data;
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
