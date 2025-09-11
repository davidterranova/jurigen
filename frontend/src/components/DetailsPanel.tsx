import { useState } from 'react';
import { useDAGStore } from '../stores/dagStore';
import type { NodeSelection, EdgeSelection, TerminalSelection } from '../types/dag';

interface DetailsPanelProps {
  className?: string;
}

const MetaDisplay = ({ meta }: { meta?: Record<string, unknown> }) => {
  if (!meta || Object.keys(meta).length === 0) {
    return (
      <p className="text-sm text-gray-500 italic">No additional metadata</p>
    );
  }

  return (
    <div className="space-y-2">
      {Object.entries(meta).map(([key, value]) => (
        <div key={key} className="flex flex-col">
          <dt className="text-xs font-medium text-gray-500 uppercase tracking-wide">
            {key}
          </dt>
          <dd className="text-sm text-gray-900 mt-1">
            {typeof value === 'string' 
              ? value 
              : JSON.stringify(value, null, 2)
            }
          </dd>
        </div>
      ))}
    </div>
  );
};

const NodeDetails = ({ selection }: { selection: NodeSelection }) => {
  const { node } = selection;
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Question Details</h3>
        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-500 font-mono bg-gray-100 px-2 py-1 rounded">
            {node.id}
          </span>
          <button
            onClick={() => copyToClipboard(node.id, 'node-id')}
            className="text-xs text-blue-600 hover:text-blue-800 font-medium"
          >
            {copiedId === 'node-id' ? 'Copied!' : 'Copy ID'}
          </button>
        </div>
      </div>

      {/* Question */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">Question</h4>
        <p className="text-sm text-gray-900 bg-gray-50 p-3 rounded-lg">
          {node.question}
        </p>
      </div>

      {/* Description */}
      {node.description && (
        <div>
          <h4 className="text-sm font-medium text-gray-700 mb-2">Description</h4>
          <p className="text-sm text-gray-600 bg-gray-50 p-3 rounded-lg">
            {node.description}
          </p>
        </div>
      )}

      {/* Answers */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">
          Available Answers ({node.answers.length})
        </h4>
        <div className="space-y-2">
          {node.answers.map((answer) => (
            <div
              key={answer.id}
              className="border border-gray-200 rounded-lg p-3 bg-white"
            >
              <div className="flex items-start justify-between mb-2">
                <span className="text-sm font-medium text-gray-900">
                  {answer.answer}
                </span>
                {answer.is_terminal && (
                  <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                    Terminal
                  </span>
                )}
              </div>
              <div className="text-xs text-gray-500">
                {answer.next_node ? (
                  <span>→ Leads to node: <code className="bg-gray-100 px-1 rounded">{answer.next_node}</code></span>
                ) : (
                  <span className="italic">No next node (terminal answer)</span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Metadata */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">Metadata</h4>
        <div className="bg-gray-50 p-3 rounded-lg">
          <MetaDisplay meta={node.meta} />
        </div>
      </div>
    </div>
  );
};

const TerminalDetails = ({ selection }: { selection: TerminalSelection }) => {
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Terminal Answer Details</h3>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Answer Text</label>
            <div className="bg-blue-50 p-3 rounded-lg border-l-4 border-blue-400">
              <p className="text-sm text-gray-900 leading-relaxed">
                {selection.answer.answer}
              </p>
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <div className="flex items-center space-x-3">
              <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                🏁 Terminal Answer
              </span>
              <span className="text-sm text-gray-600">No further questions</span>
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Source Node</label>
            <div className="flex items-center space-x-2">
              <code className="text-xs bg-gray-100 px-2 py-1 rounded font-mono">
                {selection.sourceNodeId}
              </code>
              <button
                onClick={() => copyToClipboard(selection.sourceNodeId, 'sourceNode')}
                className="text-xs text-gray-500 hover:text-gray-700"
              >
                {copiedId === 'sourceNode' ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>

          {/* Action buttons */}
          <div className="flex items-center space-x-2 pt-3">
            <button
              onClick={() => copyToClipboard(selection.terminalId, 'terminalId')}
              className="px-3 py-2 text-xs bg-blue-100 hover:bg-blue-200 rounded-lg border text-blue-700 font-medium transition-colors"
            >
              {copiedId === 'terminalId' ? '✓ Copied!' : 'Copy Terminal ID'}
            </button>
            <button
              onClick={() => copyToClipboard(selection.answer.answer, 'answer')}
              className="px-3 py-2 text-xs bg-gray-100 hover:bg-gray-200 rounded-lg border text-gray-700 transition-colors"
            >
              {copiedId === 'answer' ? '✓ Copied!' : 'Copy Answer'}
            </button>
          </div>
        </div>
      </div>

      {/* Metadata */}
      <div>
        <h4 className="text-md font-medium text-gray-900 mb-3">Metadata</h4>
        <MetaDisplay meta={selection.answer.meta} />
      </div>
    </div>
  );
};

const EdgeDetails = ({ selection }: { selection: EdgeSelection }) => {
  const { answer, sourceNodeId, targetNodeId } = selection;
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Answer Details</h3>
        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-500 font-mono bg-gray-100 px-2 py-1 rounded">
            {answer.id}
          </span>
          <button
            onClick={() => copyToClipboard(answer.id, 'answer-id')}
            className="text-xs text-blue-600 hover:text-blue-800 font-medium"
          >
            {copiedId === 'answer-id' ? 'Copied!' : 'Copy ID'}
          </button>
        </div>
      </div>

      {/* Answer Text */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">Answer</h4>
        <p className="text-sm text-gray-900 bg-gray-50 p-3 rounded-lg">
          {answer.answer}
        </p>
      </div>

      {/* Flow Information */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">Flow</h4>
        <div className="bg-gray-50 p-3 rounded-lg space-y-2">
          <div className="text-sm">
            <span className="font-medium text-gray-700">From:</span>
            <code className="ml-2 bg-gray-100 px-2 py-1 rounded text-xs">{sourceNodeId}</code>
          </div>
          <div className="text-sm">
            <span className="font-medium text-gray-700">To:</span>
            {targetNodeId ? (
              <code className="ml-2 bg-gray-100 px-2 py-1 rounded text-xs">{targetNodeId}</code>
            ) : (
              <span className="ml-2 text-gray-500 italic">Terminal (no target)</span>
            )}
          </div>
          {answer.is_terminal && (
            <div className="flex items-center mt-2">
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                Terminal Answer
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Metadata */}
      <div>
        <h4 className="text-sm font-medium text-gray-700 mb-2">Metadata</h4>
        <div className="bg-gray-50 p-3 rounded-lg">
          <MetaDisplay meta={answer.meta} />
        </div>
      </div>
    </div>
  );
};

export const DetailsPanel = ({ className = '' }: DetailsPanelProps) => {
  const { selection, isDetailsOpen, clearSelection } = useDAGStore();

  if (!isDetailsOpen || !selection) {
    return null;
  }

  return (
    <div className={`bg-white border-l border-gray-200 ${className}`}>
      <div className="flex items-center justify-between p-4 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900">
          {selection.type === 'node' ? 'Node Details' : 
           selection.type === 'terminal' ? 'Terminal Answer' : 'Answer Details'}
        </h2>
        <button
          onClick={clearSelection}
          className="text-gray-400 hover:text-gray-600 transition-colors p-1"
          aria-label="Close details panel"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div className="p-4 overflow-y-auto max-h-full">
        {selection.type === 'node' && <NodeDetails selection={selection as NodeSelection} />}
        {selection.type === 'edge' && <EdgeDetails selection={selection as EdgeSelection} />}
        {selection.type === 'terminal' && <TerminalDetails selection={selection as TerminalSelection} />}
      </div>
    </div>
  );
};

export default DetailsPanel;
