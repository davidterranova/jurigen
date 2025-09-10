import React from 'react';
import { useDAGTraversalStore } from '../stores/dagTraversalStore';
import { Node, Answer } from '../types/dag';

interface DAGViewerProps {
  className?: string;
}

export const DAGViewer: React.FC<DAGViewerProps> = ({ className = '' }) => {
  const {
    currentDAG,
    getCurrentNode,
    selectAnswer,
    goBack,
    visitedNodes,
    isComplete,
    context,
    isLoading,
    error,
  } = useDAGTraversalStore();

  const currentNode = getCurrentNode();

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center p-8 ${className}`}>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`p-4 bg-red-50 border border-red-200 rounded-md ${className}`}>
        <div className="flex">
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">Error</h3>
            <p className="mt-2 text-sm text-red-700">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  if (!currentDAG || !currentNode) {
    return (
      <div className={`p-8 text-center text-gray-500 ${className}`}>
        <p>No DAG loaded or no current node selected.</p>
      </div>
    );
  }

  if (isComplete) {
    return (
      <CompletionView 
        context={context} 
        onRestart={() => useDAGTraversalStore.getState().reset()}
        className={className}
      />
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Progress indicator */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <div className="flex items-center justify-between text-sm text-gray-600 mb-2">
          <span>Progress</span>
          <span>{visitedNodes.length} nodes visited</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div 
            className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
            style={{ width: `${Math.min(100, (visitedNodes.length / Object.keys(currentDAG.nodes).length) * 100)}%` }}
          />
        </div>
      </div>

      {/* Navigation */}
      {visitedNodes.length > 1 && (
        <button
          onClick={goBack}
          className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          ← Back
        </button>
      )}

      {/* Current question */}
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">
          {currentNode.question}
        </h2>

        {currentNode.answers.length > 0 ? (
          <div className="space-y-3">
            {currentNode.answers.map((answer) => (
              <AnswerButton
                key={answer.id}
                answer={answer}
                onClick={() => selectAnswer(currentNode.id, answer)}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <p className="text-gray-500 mb-4">This is a leaf node - no answers available.</p>
            <button
              onClick={() => useDAGTraversalStore.getState().reset()}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Complete Traversal
            </button>
          </div>
        )}
      </div>

      {/* Context so far */}
      {context.length > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h3 className="text-lg font-medium text-blue-900 mb-3">Context Built So Far:</h3>
          <ul className="space-y-1">
            {context.map((item, index) => (
              <li key={index} className="text-sm text-blue-800">
                {index + 1}. {item}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

interface AnswerButtonProps {
  answer: Answer;
  onClick: () => void;
}

const AnswerButton: React.FC<AnswerButtonProps> = ({ answer, onClick }) => {
  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors"
    >
      <div className="flex items-center justify-between">
        <span className="text-gray-900">{answer.answer}</span>
        <span className="text-gray-400">→</span>
      </div>
      {answer.user_context && (
        <p className="text-sm text-gray-600 mt-1">{answer.user_context}</p>
      )}
    </button>
  );
};

interface CompletionViewProps {
  context: string[];
  onRestart: () => void;
  className?: string;
}

const CompletionView: React.FC<CompletionViewProps> = ({ context, onRestart, className = '' }) => {
  return (
    <div className={`text-center space-y-6 ${className}`}>
      <div className="bg-green-50 border border-green-200 rounded-lg p-8">
        <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
          <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 className="text-2xl font-bold text-green-900 mb-2">Context Complete!</h2>
        <p className="text-green-700">You have successfully built the legal case context.</p>
      </div>

      <div className="bg-white border border-gray-200 rounded-lg p-6 text-left">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Final Context:</h3>
        <ol className="space-y-2">
          {context.map((item, index) => (
            <li key={index} className="text-gray-700">
              <span className="font-medium">{index + 1}.</span> {item}
            </li>
          ))}
        </ol>
      </div>

      <button
        onClick={onRestart}
        className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Start New Traversal
      </button>
    </div>
  );
};
