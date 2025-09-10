import React, { useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useDAG } from '../hooks/useDAG';
import { useDAGTraversalStore } from '../stores/dagTraversalStore';
import { DAGViewer } from '../components/DAGViewer';

export const DAGDetailPage: React.FC = () => {
  const { dagId } = useParams<{ dagId: string }>();
  const { data: dag, isLoading, error } = useDAG(dagId!);
  const { setDAG, startTraversal, reset, currentDAG } = useDAGTraversalStore();

  useEffect(() => {
    if (dag) {
      setDAG(dag);
    }
  }, [dag, setDAG]);

  const handleStartTraversal = () => {
    if (!dag) return;
    
    // Find root node - in a proper implementation, this would be determined by the backend
    // For now, we'll take the first node as root
    const rootNodeId = Object.keys(dag.nodes)[0];
    if (rootNodeId) {
      startTraversal(rootNodeId);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="p-4 bg-red-50 border border-red-200 rounded-md">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading DAG</h3>
              <p className="mt-2 text-sm text-red-700">{error.message}</p>
            </div>
          </div>
        </div>
        <Link
          to="/dags"
          className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          ← Back to DAGs
        </Link>
      </div>
    );
  }

  if (!dag) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">DAG not found</p>
        <Link
          to="/dags"
          className="mt-4 inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          ← Back to DAGs
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <Link
            to="/dags"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700 mb-2"
          >
            ← Back to DAGs
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">{dag.title}</h1>
          <p className="text-gray-600 mt-1">
            {Object.keys(dag.nodes).length} nodes in this DAG
          </p>
        </div>
        
        <div className="space-x-3">
          {currentDAG && (
            <button
              onClick={reset}
              className="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Reset
            </button>
          )}
          <button
            onClick={handleStartTraversal}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            Start Traversal
          </button>
        </div>
      </div>

      {/* DAG Viewer */}
      <div className="bg-gray-50 rounded-lg p-6">
        <DAGViewer />
      </div>

      {/* DAG Structure Preview */}
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">DAG Structure</h2>
        <div className="space-y-4">
          {Object.values(dag.nodes).map((node) => (
            <div key={node.id} className="border border-gray-200 rounded p-4">
              <h3 className="font-medium text-gray-900 mb-2">{node.question}</h3>
              {node.answers.length > 0 ? (
                <div className="space-y-2">
                  <p className="text-sm text-gray-600">Answers:</p>
                  <ul className="list-disc list-inside space-y-1 text-sm text-gray-700">
                    {node.answers.map((answer) => (
                      <li key={answer.id}>
                        {answer.answer}
                        {answer.next_node && (
                          <span className="text-gray-500 ml-2">
                            → {dag.nodes[answer.next_node]?.question?.substring(0, 50)}...
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </div>
              ) : (
                <p className="text-sm text-gray-500 italic">Leaf node - no answers</p>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
