import { useEffect, useState } from 'react';
import { useDAGStore } from '../stores/dagStore';
import { DAGService } from '../services/dagService';

interface DagListProps {
  className?: string;
}

export const DagList = ({ className = '' }: DagListProps) => {
  const {
    availableDAGs,
    selectedDAGId,
    setAvailableDAGs,
    selectDAG,
    isLoading,
    setLoading,
    setError,
  } = useDAGStore();

  const [searchTerm, setSearchTerm] = useState('');

  // Load available DAGs on mount
  useEffect(() => {
    const loadDAGs = async () => {
      try {
        setLoading(true);
        const dags = await DAGService.listDAGs();
        setAvailableDAGs(dags);
      } catch (error) {
        setError(error instanceof Error ? error.message : 'Failed to load DAGs');
      } finally {
        setLoading(false);
      }
    };

    loadDAGs();
  }, [setAvailableDAGs, setLoading, setError]);

  // Filter DAGs based on search term
  const filteredDAGs = Array.isArray(availableDAGs) 
    ? availableDAGs.filter(dag =>
        dag.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        dag.description?.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : [];

  const handleDagSelect = (dagId: string) => {
    selectDAG(dagId);
  };

  return (
    <div className={`bg-white border-r border-gray-200 ${className}`}>
      <div className="p-4">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">DAG Library</h2>
        
        {/* Search input */}
        <div className="mb-4">
          <input
            type="text"
            placeholder="Search DAGs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm"
          />
        </div>

        {/* DAG list */}
        <div className="space-y-2">
          {isLoading ? (
            // Loading skeletons
            Array.from({ length: 3 }).map((_, index) => (
              <div key={index} className="animate-pulse">
                <div className="h-16 bg-gray-200 rounded-lg"></div>
              </div>
            ))
          ) : filteredDAGs.length === 0 ? (
            // Empty state
            <div className="text-center py-8 text-gray-500">
              {searchTerm ? 'No DAGs match your search.' : 'No DAGs available.'}
            </div>
          ) : (
            // DAG items
            filteredDAGs.map((dag) => (
              <button
                key={dag.id}
                onClick={() => handleDagSelect(dag.id)}
                className={`w-full text-left p-3 rounded-lg border transition-colors ${
                  selectedDAGId === dag.id
                    ? 'bg-blue-50 border-blue-200 ring-2 ring-blue-500'
                    : 'bg-white border-gray-200 hover:bg-gray-50 hover:border-gray-300'
                }`}
              >
                <div className="flex flex-col">
                  <h3 className="font-medium text-gray-900 truncate">
                    {dag.name}
                  </h3>
                  {dag.description && (
                    <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                      {dag.description}
                    </p>
                  )}
                  <div className="flex items-center justify-between mt-2 text-xs text-gray-400">
                    <span>
                      {dag.node_count ? `${dag.node_count} nodes` : 'Unknown nodes'}
                    </span>
                    {dag.updated_at && (
                      <span>
                        {new Date(dag.updated_at).toLocaleDateString()}
                      </span>
                    )}
                  </div>
                </div>
              </button>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default DagList;
