import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { DAGService, dagQueryKeys } from '../services/dagService';
import { DAG, DAGSummary } from '../types/dag';

/**
 * Hook to fetch all DAGs
 */
export const useDAGs = () => {
  return useQuery({
    queryKey: dagQueryKeys.lists(),
    queryFn: DAGService.listDAGs,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

/**
 * Hook to fetch a specific DAG by ID
 */
export const useDAG = (dagId: string) => {
  return useQuery({
    queryKey: dagQueryKeys.detail(dagId),
    queryFn: () => DAGService.getDAG(dagId),
    enabled: !!dagId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
};

/**
 * Hook to update a DAG
 */
export const useUpdateDAG = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ dagId, dag }: { dagId: string; dag: DAG }) =>
      DAGService.updateDAG(dagId, dag),
    onSuccess: (data, variables) => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: dagQueryKeys.lists() });
      queryClient.setQueryData(dagQueryKeys.detail(variables.dagId), data);
    },
  });
};

/**
 * Hook to validate a DAG
 */
export const useValidateDAG = () => {
  return useMutation({
    mutationFn: (dag: DAG) => DAGService.validateDAG(dag),
  });
};

/**
 * Hook to create a new DAG
 */
export const useCreateDAG = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: DAGService.createDAG,
    onSuccess: () => {
      // Invalidate the DAGs list
      queryClient.invalidateQueries({ queryKey: dagQueryKeys.lists() });
    },
  });
};

/**
 * Hook to delete a DAG
 */
export const useDeleteDAG = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (dagId: string) => DAGService.deleteDAG(dagId),
    onSuccess: () => {
      // Invalidate the DAGs list
      queryClient.invalidateQueries({ queryKey: dagQueryKeys.lists() });
    },
  });
};
