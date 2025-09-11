import { useCallback, useEffect, useImperativeHandle, forwardRef, useMemo } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useReactFlow,
  useNodesState,
  useEdgesState,
  Handle,
  Position,
  MarkerType,
} from '@xyflow/react';
import type { Node, Edge, NodeChange, EdgeChange } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useDAGStore } from '../stores/dagStore';
import type { DAG, DAGNode, DAGEdge, TerminalNode, Answer } from '../types/dag';
import { DAGService } from '../services/dagService';

// Custom node component for DAG nodes
const DAGNodeComponent = ({ data, selected }: { data: { question: string; answers: unknown[] }; selected: boolean }) => {
  const { question, answers } = data;
  
  return (
    <>
      {/* Input handle - for incoming edges */}
      <Handle
        type="target"
        position={Position.Top}
        style={{ 
          background: '#6b7280', 
          width: 8, 
          height: 8,
          border: '2px solid #fff'
        }}
      />
      
      <div
        className={`px-4 py-3 rounded-lg border-2 bg-white shadow-sm min-w-48 max-w-64 ${
          selected
            ? 'border-blue-500 ring-2 ring-blue-200'
            : 'border-gray-300 hover:border-gray-400'
        }`}
      >
        <div className="font-medium text-gray-900 text-sm mb-2 line-clamp-3">
          {question}
        </div>
        <div className="text-xs text-gray-500">
          {answers.length} answer{answers.length !== 1 ? 's' : ''}
        </div>
      </div>

      {/* Output handle - for outgoing edges */}
      <Handle
        type="source"
        position={Position.Bottom}
        style={{ 
          background: '#6b7280', 
          width: 8, 
          height: 8,
          border: '2px solid #fff'
        }}
      />
    </>
  );
};

// Custom terminal node component for terminal answers
const TerminalNodeComponent = ({ 
  data, 
  selected 
}: { 
  data: { 
    answer: string;
    sourceNodeId: string;
    meta?: Record<string, unknown>;
  }; 
  selected: boolean 
}) => {
  const { answer } = data;
  
  const getBorderClass = () => {
    if (selected) {
      return 'border-blue-500 ring-2 ring-blue-200 border-3';
    }
    return 'border-blue-400 border-2 hover:border-blue-500';
  };

  const nodeAriaLabel = `Terminal answer: ${answer}`;
  
  return (
    <>
      {/* Input handle - for incoming edges */}
      <Handle
        type="target"
        position={Position.Top}
        style={{ 
          background: '#3b82f6', 
          width: 6, 
          height: 6,
          border: '2px solid #fff'
        }}
      />
      
      <div
        className={`px-3 py-2 rounded-full bg-blue-50 shadow-sm min-w-24 max-w-48 text-center relative ${getBorderClass()}`}
        aria-label={nodeAriaLabel}
        role="button"
        tabIndex={0}
      >
        {/* END badge */}
        <div className="absolute -top-2 -right-2 bg-blue-500 text-white text-xs font-bold px-1.5 py-0.5 rounded-full shadow-sm">
          END
        </div>
        
        <div className="font-medium text-blue-900 text-xs line-clamp-2">
          {answer.length > 40 ? answer.substring(0, 40) + '...' : answer}
        </div>
      </div>
    </>
  );
};

const nodeTypes = {
  dagNode: DAGNodeComponent,
  terminalNode: TerminalNodeComponent,
};

export interface GraphViewRef {
  fitView: () => void;
  zoomIn: () => void;
  zoomOut: () => void;
  resetView: () => void;
  setLayoutMode: (mode: 'auto' | 'manual') => void;
  resetToAutoLayout: () => void;
  applyAutoLayout: () => void;
}

interface GraphViewProps {
  className?: string;
}

export const GraphView = forwardRef<GraphViewRef, GraphViewProps>(
  ({ className = '' }, ref) => {
    const {
      selectedDAGId,
      selectedDAG,
      selection,
      layoutMode,
      customPositions,
      setSelectedDAG,
      setNodes,
      setEdges,
      selectNode,
      selectEdge,
      selectTerminal,
      clearSelection,
      setLoading,
      setError,
      setLayoutMode,
      updateCustomPosition,
      resetToAutoLayout,
    } = useDAGStore();

    const { fitView, zoomIn, zoomOut, setCenter } = useReactFlow();
    const [reactFlowNodes, setReactFlowNodes, onNodesChange] = useNodesState<DAGNode | TerminalNode>([]);
    const [reactFlowEdges, setReactFlowEdges, onEdgesChange] = useEdgesState<DAGEdge>([]);

    // Expose methods via ref
    useImperativeHandle(ref, () => ({
      fitView: () => fitView({ padding: 0.1 }),
      zoomIn,
      zoomOut,
      resetView: () => {
        setCenter(0, 0);
        fitView({ padding: 0.1 });
      },
      setLayoutMode,
      resetToAutoLayout,
      applyAutoLayout: () => {
        // Force re-calculation of layout even in manual mode
        if (selectedDAG) {
          // Force recalculation of layout including terminals
          const { nodes: flowNodes, edges: flowEdges } = convertDAGToFlow(selectedDAG);
          
          setReactFlowNodes(flowNodes);
          setReactFlowEdges(flowEdges);
          fitView({ padding: 0.1 });
        }
      },
    }));

    // Layout algorithm constants (moved to useMemo to fix dependency warnings)
    const LAYOUT_CONFIG = useMemo(() => ({
      NODE_HORIZONTAL_SPACING: 180, // Reduced from 250 to minimize edge length
      LAYER_VERTICAL_SPACING: 120,  // Reduced from 180 for more compact layout
      NODE_WIDTH: 200,              // Slightly reduced from 220
      LAYER_CENTERING: true,
      START_X: 100,
      START_Y: 50,
      MIN_NODE_SPACING: 20,         // Minimum space between nodes
      EDGE_LENGTH_WEIGHT: 0.3,      // Weight for edge length in positioning
    }), []);

    // Calculate hierarchical DAG layout
    const calculateDAGLayout = useCallback((dag: DAG): Record<string, { x: number; y: number }> => {
      const nodeIds = Object.keys(dag.nodes);
      
      console.log('🎯 Starting DAG layout calculation for nodes:', nodeIds);
      
      // Find root node (from dag.root_node or node with no incoming edges)
      let rootId = dag.root_node;
      if (!rootId || !dag.nodes[rootId]) {
        // Fallback: find node with no incoming edges
        const hasIncomingEdge = new Set<string>();
        nodeIds.forEach(nodeId => {
          model.Nodes[nodeId].answers.forEach(answer => {
            if (answer.next_node) {
              hasIncomingEdge.add(answer.next_node);
            }
          });
        });
        rootId = nodeIds.find(id => !hasIncomingEdge.has(id)) || nodeIds[0];
      }

      console.log('🌳 Root node identified:', rootId);

      // Calculate node levels using BFS
      const nodeLevels = getNodeLevels(dag, rootId);
      console.log('📊 Node levels:', nodeLevels);

      // Group nodes by level
      const levelGroups: Record<number, string[]> = {};
      Object.entries(nodeLevels).forEach(([nodeId, level]) => {
        if (!levelGroups[level]) levelGroups[level] = [];
        levelGroups[level].push(nodeId);
      });

      const maxLevel = Math.max(...Object.keys(levelGroups).map(Number));
      console.log('🏗️ Level groups:', levelGroups, 'Max level:', maxLevel);

      // Position nodes within each layer
      const positions: Record<string, { x: number; y: number }> = {};
      
      for (let level = 0; level <= maxLevel; level++) {
        const nodesInLevel = levelGroups[level] || [];
        // Remove unused startX variable - using centeredStartX instead

        // Minimize crossings by sorting nodes based on their connections
        const sortedNodes = minimizeCrossings(nodesInLevel, dag, level > 0 ? levelGroups[level - 1] : []);

        // Calculate dynamic spacing based on number of nodes and available space
        const minSpacing = LAYOUT_CONFIG.NODE_WIDTH + LAYOUT_CONFIG.MIN_NODE_SPACING;
        const actualSpacing = Math.max(minSpacing, LAYOUT_CONFIG.NODE_HORIZONTAL_SPACING);
        
        // Calculate layer width and centering for better positioning
        const layerWidth = (sortedNodes.length - 1) * actualSpacing;
        const centeredStartX = LAYOUT_CONFIG.LAYER_CENTERING ? 
          LAYOUT_CONFIG.START_X - (layerWidth / 2) : 
          LAYOUT_CONFIG.START_X;

        sortedNodes.forEach((nodeId, index) => {
          positions[nodeId] = {
            x: centeredStartX + (index * actualSpacing),
            y: LAYOUT_CONFIG.START_Y + (level * LAYOUT_CONFIG.LAYER_VERTICAL_SPACING),
          };
        });

        console.log(`📊 Level ${level}: ${sortedNodes.length} nodes, spacing: ${actualSpacing}px, width: ${layerWidth}px`);
      }

      console.log('📍 Final positions:', positions);
      return positions;
    }, [LAYOUT_CONFIG]);

    // Calculate optimal node levels considering edge lengths and layer distribution
    const getNodeLevels = useCallback((dag: DAG, rootId: string): Record<string, number> => {
      const levels: Record<string, number> = {};
      const nodeIds = Object.keys(dag.nodes);
      
      // Step 1: Initial BFS level assignment
      const visited = new Set<string>();
      const queue: Array<{ nodeId: string; level: number }> = [{ nodeId: rootId, level: 0 }];

      while (queue.length > 0) {
        const { nodeId, level } = queue.shift()!;

        if (visited.has(nodeId)) continue;
        visited.add(nodeId);
        levels[nodeId] = level;

        // Add children to queue
        const node = model.Nodes[nodeId];
        if (node && node.answers) {
          node.answers.forEach(answer => {
            if (answer.next_node && !visited.has(answer.next_node)) {
              queue.push({ nodeId: answer.next_node, level: level + 1 });
            }
          });
        }
      }

      // Handle any unvisited nodes
      Object.keys(dag.nodes).forEach(nodeId => {
        if (!Object.prototype.hasOwnProperty.call(levels, nodeId)) {
          levels[nodeId] = 0;
        }
      });

      // Step 2: Optimize levels to reduce long edges and balance layers
      const maxIterations = 5;
      let improved = true;
      let iteration = 0;

      while (improved && iteration < maxIterations) {
        improved = false;
        iteration++;

        // Try to move nodes to better levels
        nodeIds.forEach(nodeId => {
          if (nodeId === rootId) return; // Don't move root

          const currentLevel = levels[nodeId];
          const node = model.Nodes[nodeId];
          
          // Calculate constraints from predecessors and successors
          let minLevel = 0;
          let maxLevel = Object.keys(levels).length;
          
          // Minimum level based on predecessors
          nodeIds.forEach(predId => {
            const predNode = model.Nodes[predId];
            predNode.answers.forEach(answer => {
              if (answer.next_node === nodeId) {
                minLevel = Math.max(minLevel, levels[predId] + 1);
              }
            });
          });
          
          // Maximum level based on successors
          if (node && node.answers) {
            node.answers.forEach(answer => {
              if (answer.next_node && levels[answer.next_node] !== undefined) {
                maxLevel = Math.min(maxLevel, levels[answer.next_node] - 1);
              }
            });
          }

          // Find the best level within constraints
          let bestLevel = currentLevel;
          let bestScore = calculateNodeScore(nodeId, currentLevel, dag, levels);
          
          for (let testLevel = minLevel; testLevel <= Math.min(maxLevel, currentLevel + 2); testLevel++) {
            if (testLevel !== currentLevel) {
              const testScore = calculateNodeScore(nodeId, testLevel, dag, levels);
              if (testScore < bestScore) {
                bestScore = testScore;
                bestLevel = testLevel;
              }
            }
          }

          if (bestLevel !== currentLevel) {
            levels[nodeId] = bestLevel;
            improved = true;
            console.log(`📈 Moved node ${nodeId} from level ${currentLevel} to ${bestLevel}`);
          }
        });
      }

      console.log(`🎯 Level optimization completed in ${iteration} iterations`);
      return levels;
    }, []);

    // Calculate a score for a node at a given level (lower is better)
    const calculateNodeScore = useCallback((nodeId: string, level: number, dag: DAG, levels: Record<string, number>): number => {
      let score = 0;
      const node = model.Nodes[nodeId];
      
      // Penalize long edges (edge length contributes to score)
      Object.keys(dag.nodes).forEach(predId => {
        const predNode = model.Nodes[predId];
        predNode.answers.forEach(answer => {
          if (answer.next_node === nodeId) {
            const edgeLength = Math.abs(level - levels[predId]);
            score += edgeLength * edgeLength; // Quadratic penalty for long edges
          }
        });
      });

      if (node && node.answers) {
        node.answers.forEach(answer => {
          if (answer.next_node && levels[answer.next_node] !== undefined) {
            const edgeLength = Math.abs(levels[answer.next_node] - level);
            score += edgeLength * edgeLength;
          }
        });
      }

      return score;
    }, []);

    // Enhanced crossing minimization using multiple criteria
    const minimizeCrossings = useCallback((layerNodes: string[], dag: DAG, previousLayer: string[]): string[] => {
      if (layerNodes.length <= 1) return layerNodes;

      // Multi-criteria sorting to minimize crossings and edge lengths
      const nodeScores = layerNodes.map(nodeId => {
        let predecessorScore = 0;
        let connectionCount = 0;

        // Calculate position based on predecessors
        previousLayer.forEach((prevNodeId, prevIndex) => {
          const prevNode = model.Nodes[prevNodeId];
          if (prevNode && prevNode.answers) {
            prevNode.answers.forEach(answer => {
              if (answer.next_node === nodeId) {
                predecessorScore += prevIndex;
                connectionCount++;
              }
            });
          }
        });

        // Consider successors for better positioning
        const node = model.Nodes[nodeId];
        let successorInfluence = 0;
        let successorCount = 0;
        
        if (node && node.answers) {
          node.answers.forEach(answer => {
            if (answer.next_node) {
              // This is a simple heuristic - in a real implementation,
              // we'd know the positions of successors from future iterations
              successorInfluence += answer.next_node.length; // Simple hash-based influence
              successorCount++;
            }
          });
        }

        // Calculate final score with multiple criteria
        const avgPredecessorPos = connectionCount > 0 ? predecessorScore / connectionCount : layerNodes.length / 2;
        const avgSuccessorInfluence = successorCount > 0 ? successorInfluence / successorCount : 0;
        
        // Combine scores with weights
        const finalScore = avgPredecessorPos * 0.7 + (avgSuccessorInfluence % 100) * 0.3;

        return {
          nodeId,
          score: finalScore,
          connectionCount,
          debug: { avgPredecessorPos, avgSuccessorInfluence, finalScore }
        };
      });

      // Sort by score and return node IDs
      const result = nodeScores.sort((a, b) => a.score - b.score).map(item => item.nodeId);
      
      console.log('🔄 Crossing minimization scores:', nodeScores.map(n => ({ 
        node: n.nodeId.substring(0, 8), 
        score: n.score.toFixed(2),
        connections: n.connectionCount 
      })));
      
      return result;
    }, []);

    // Calculate optimal positions for terminal nodes to prevent overlaps
    const calculateTerminalPositions = useCallback((
      dag: DAG, 
      nodePositions: Record<string, { x: number; y: number }>
    ): Record<string, { x: number; y: number }> => {
      const LAYOUT_CONFIG = {
        TERMINAL_VERTICAL_OFFSET: 140, // Distance below source node
        TERMINAL_HORIZONTAL_SPACING: 180, // Space between terminals
        TERMINAL_NODE_WIDTH: 120, // Estimated terminal node width
        MIN_TERMINAL_SPACING: 20, // Minimum gap between terminals
      };

      const terminalPositions: Record<string, { x: number; y: number }> = {};
      const terminalsByLevel: Record<number, Array<{ id: string; sourceX: number; sourceY: number }>> = {};

      // Group terminals by their Y-level (based on source node Y position)
      Object.keys(dag.nodes).forEach((nodeId) => {
        const dagNode = model.Nodes[nodeId];
        const sourcePosition = nodePositions[nodeId] || { x: 0, y: 0 };
        const terminalAnswers = dagNode.answers.filter(a => !a.next_node);

        if (terminalAnswers.length > 0) {
          const terminalLevel = sourcePosition.y + LAYOUT_CONFIG.TERMINAL_VERTICAL_OFFSET;
          
          if (!terminalsByLevel[terminalLevel]) {
            terminalsByLevel[terminalLevel] = [];
          }

          terminalAnswers.forEach((answer) => {
            const terminalId = `terminal-${nodeId}-${answer.id}`;
            terminalsByLevel[terminalLevel].push({
              id: terminalId,
              sourceX: sourcePosition.x,
              sourceY: sourcePosition.y,
            });
          });
        }
      });

      // Position terminals within each level to prevent overlaps
      Object.keys(terminalsByLevel).forEach((levelY) => {
        const level = parseInt(levelY);
        const terminalsAtLevel = terminalsByLevel[level];
        
        // Sort terminals by source X position for consistent ordering
        terminalsAtLevel.sort((a, b) => a.sourceX - b.sourceX);
        
        // Calculate positions with proper spacing
        terminalsAtLevel.forEach((terminal) => {
          // Start with source position
          let baseX = terminal.sourceX;
          
          // For multiple terminals from the same source, spread them horizontally
          const sameSourceTerminals = terminalsAtLevel.filter(t => 
            Math.abs(t.sourceX - terminal.sourceX) < 50 // Within 50px considered same source
          );
          
          if (sameSourceTerminals.length > 1) {
            const terminalIndex = sameSourceTerminals.findIndex(t => t.id === terminal.id);
            const totalWidth = (sameSourceTerminals.length - 1) * LAYOUT_CONFIG.TERMINAL_HORIZONTAL_SPACING;
            baseX = terminal.sourceX - (totalWidth / 2) + (terminalIndex * LAYOUT_CONFIG.TERMINAL_HORIZONTAL_SPACING);
          }
          
          // Check for conflicts with other terminals at this level and adjust
          let finalX = baseX;
          let attempts = 0;
          const maxAttempts = 20;
          
          while (attempts < maxAttempts) {
            let hasConflict = false;
            
            // Check against already positioned terminals at this level
            for (const otherTerminalId of Object.keys(terminalPositions)) {
              const otherPos = terminalPositions[otherTerminalId];
              if (Math.abs(otherPos.y - level) < 10) { // Same level
                const distance = Math.abs(otherPos.x - finalX);
                const minDistance = LAYOUT_CONFIG.TERMINAL_NODE_WIDTH + LAYOUT_CONFIG.MIN_TERMINAL_SPACING;
                
                if (distance < minDistance) {
                  hasConflict = true;
                  // Move to the right to avoid conflict
                  finalX = otherPos.x + minDistance;
                  break;
                }
              }
            }
            
            if (!hasConflict) break;
            attempts++;
          }

          terminalPositions[terminal.id] = {
            x: finalX,
            y: level,
          };
        });
      });

      console.log('📍 Calculated terminal positions:', terminalPositions);
      return terminalPositions;
    }, []);

    // Convert DAG to React Flow format
    const convertDAGToFlow = useCallback((dag: DAG): { nodes: (DAGNode | TerminalNode)[]; edges: DAGEdge[] } => {
      const flowNodes: (DAGNode | TerminalNode)[] = [];
      const flowEdges: DAGEdge[] = [];
      
      // Calculate hierarchical layout positions
      const nodePositions = calculateDAGLayout(dag);
      
      // Calculate optimal terminal positions
      const terminalPositions = calculateTerminalPositions(dag, nodePositions);

      // Convert nodes
      Object.keys(dag.nodes).forEach((nodeId) => {
        const dagNode = model.Nodes[nodeId];
        // Use custom position if in manual mode and position exists, otherwise use calculated position
        const position = (layoutMode === 'manual' && customPositions[nodeId]) 
          ? customPositions[nodeId] 
          : nodePositions[nodeId] || { x: 0, y: 0 };
        
        // Check if this is the root node
        const isRoot = nodeId === dag.root_node;

        const flowNode: DAGNode = {
          id: nodeId,
          type: 'dagNode',
          position,
          data: {
            question: dagNode.question,
            description: dagNode.description,
            answers: dagNode.answers,
            meta: dagNode.meta,
            isRoot,
          },
          targetPosition: Position.Top,
          sourcePosition: Position.Bottom,
        };

        flowNodes.push(flowNode);
      });

      // Convert edges (answers become edges) and create terminal nodes
      Object.keys(dag.nodes).forEach((nodeId) => {
        const dagNode = model.Nodes[nodeId];
        console.log(`🔍 Processing node ${nodeId} with ${dagNode.answers.length} answers`);
        
        dagNode.answers.forEach((answer: Answer) => {
          console.log(`  📝 Answer: "${answer.answer}" -> ${answer.next_node || 'TERMINAL'}`);
          if (answer.next_node) {
            // Create regular edge to next node
            const edgeId = `${nodeId}-${answer.id}`;
            const flowEdge: DAGEdge = {
              id: edgeId,
              type: 'bezier',
              source: nodeId,
              target: answer.next_node,
              label: answer.answer.length > 30 ? answer.answer.substring(0, 30) + '...' : answer.answer,
              data: {
                answer: answer.answer,
                is_terminal: answer.is_terminal,
                meta: answer.meta,
              },
              markerEnd: {
                type: MarkerType.ArrowClosed,
                width: 20,
                height: 20,
                color: answer.is_terminal ? '#ef4444' : '#6b7280',
              },
              style: {
                stroke: answer.is_terminal ? '#ef4444' : '#6b7280',
                strokeWidth: 2,
                strokeDasharray: answer.is_terminal ? '5,5' : undefined,
              },
              labelStyle: {
                fontSize: 12,
                fontWeight: 500,
                backgroundColor: 'white',
                padding: '2px 6px',
                borderRadius: '4px',
                border: '1px solid #e5e7eb',
              },
            };

            flowEdges.push(flowEdge);
          } else {
            // Create terminal node for answers without next_node
            const terminalId = `terminal-${nodeId}-${answer.id}`;
            
            // Use calculated terminal position or fall back to default
            const terminalPosition = terminalPositions[terminalId] || {
              x: nodePositions[nodeId]?.x || 0,
              y: (nodePositions[nodeId]?.y || 0) + 140,
            };

            // In manual mode, use custom position if available
            const finalPosition = (layoutMode === 'manual' && customPositions[terminalId]) 
              ? customPositions[terminalId] 
              : terminalPosition;

            console.log(`  🎯 Creating terminal node for answer: "${answer.answer}" at position:`, finalPosition);
            
            const terminalNode: TerminalNode = {
              id: terminalId,
              type: 'terminalNode',
              position: finalPosition,
              data: {
                answer: answer.answer,
                sourceNodeId: nodeId,
                meta: answer.meta,
              },
              targetPosition: Position.Top,
              sourcePosition: Position.Bottom,
            };

            flowNodes.push(terminalNode);

            // Create edge from source node to terminal node
            const terminalEdgeId = `edge-${terminalId}`;
            const terminalEdge: DAGEdge = {
              id: terminalEdgeId,
              type: 'bezier',
              source: nodeId,
              target: terminalId,
              label: answer.answer.length > 30 ? answer.answer.substring(0, 30) + '...' : answer.answer,
              data: {
                answer: answer.answer,
                is_terminal: true,
                meta: answer.meta,
              },
              markerEnd: {
                type: MarkerType.ArrowClosed,
                width: 20,
                height: 20,
                color: '#3b82f6',
              },
              style: {
                stroke: '#3b82f6',
                strokeWidth: 2,
                strokeDasharray: '5,5',
              },
              labelStyle: {
                fontSize: 12,
                fontWeight: 500,
                backgroundColor: 'white',
                padding: '2px 6px',
                borderRadius: '4px',
                border: '1px solid #e5e7eb',
              },
            };

            flowEdges.push(terminalEdge);
          }
        });
      });

      return { nodes: flowNodes, edges: flowEdges };
    }, [calculateDAGLayout, calculateTerminalPositions, layoutMode, customPositions]);

    // Load DAG data when selectedDAGId changes
    useEffect(() => {
      if (!selectedDAGId) {
        setNodes([]);
        setEdges([]);
        setReactFlowNodes([]);
        setReactFlowEdges([]);
        return;
      }

      if (selectedDAG?.id === selectedDAGId) {
        // Already have the right DAG
        return;
      }

      const loadDAG = async () => {
        try {
          setLoading(true);
          const dag = await DAGService.getDAG(selectedDAGId);
          setSelectedDAG(dag);
        } catch (error) {
          setError(error instanceof Error ? error.message : 'Failed to load DAG');
        } finally {
          setLoading(false);
        }
      };

      loadDAG();
    }, [selectedDAGId, selectedDAG, setSelectedDAG, setLoading, setError, setNodes, setEdges, setReactFlowNodes, setReactFlowEdges]);

    // Convert DAG to flow format when DAG changes
    useEffect(() => {
      if (!selectedDAG) return;

      console.log('🔍 Converting DAG to flow format:', selectedDAG);
      const { nodes: flowNodes, edges: flowEdges } = convertDAGToFlow(selectedDAG);
      console.log('📊 Created nodes:', flowNodes.length);
      console.log('🔗 Created edges:', flowEdges.length, flowEdges);
      
      setNodes(flowNodes);
      setEdges(flowEdges);
      setReactFlowNodes(flowNodes);
      setReactFlowEdges(flowEdges);
    }, [selectedDAG, convertDAGToFlow, setNodes, setEdges, setReactFlowNodes, setReactFlowEdges, calculateDAGLayout]);

    // Handle node selection
    const onNodeClick = useCallback(
      (event: React.MouseEvent, node: Node) => {
        event.stopPropagation();
        
        // Check if this is a terminal node
        if (node.type === 'terminalNode') {
          const terminalData = node.data as TerminalNode['data'];
          console.log('🎯 Terminal node clicked:', terminalData);
          
          // Find the answer from the source node
          const { selectedDAG } = useDAGStore.getState();
          if (selectedDAG) {
            const sourceNode = selectedDAG.nodes[terminalData.sourceNodeId];
            const answer = sourceNode?.answers.find(a => a.answer === terminalData.answer);
            if (answer) {
              selectTerminal(node.id, terminalData.sourceNodeId, answer);
            }
          }
        } else {
          selectNode(node.id);
        }
      },
      [selectNode, selectTerminal]
    );

    // Handle edge selection
    const onEdgeClick = useCallback(
      (event: React.MouseEvent, edge: Edge) => {
        event.stopPropagation();
        selectEdge(edge.id, edge.source, edge.target);
      },
      [selectEdge]
    );

    // Handle pane click (clear selection)
    const onPaneClick = useCallback(() => {
      clearSelection();
    }, [clearSelection]);

    // Handle node changes (allow position changes in manual mode)
    const handleNodesChange = useCallback(
      (changes: NodeChange<DAGNode | TerminalNode>[]) => {
        if (layoutMode === 'manual') {
          // In manual mode, allow all changes including position changes
          onNodesChange(changes);
        } else {
          // In auto mode, filter out position changes to keep layout controlled
          const filteredChanges = changes.filter(change => change.type !== 'position');
          onNodesChange(filteredChanges);
        }
      },
      [onNodesChange, layoutMode]
    );

    const handleEdgesChange = useCallback(
      (changes: EdgeChange<DAGEdge>[]) => {
        onEdgesChange(changes);
      },
      [onEdgesChange]
    );

    // Handle node drag events
    const onNodeDrag = useCallback(
      () => {
        // Real-time drag feedback (optional - can be used for visual feedback)
      },
      []
    );

    const onNodeDragStop = useCallback(
      (_event: React.MouseEvent, node: Node) => {
        // Update custom position when drag ends
        if (layoutMode === 'manual') {
          updateCustomPosition(node.id, node.position);
          console.log(`🖱️ ${node.type === 'terminalNode' ? 'Terminal' : 'Node'} dragged to:`, node.position);
        }
      },
      [layoutMode, updateCustomPosition]
    );

    // Prevent connections
    const onConnect = useCallback(
      () => {
        // Do nothing - this is read-only
      },
      []
    );

    // Update node selection state
    useEffect(() => {
      if (selection?.type === 'node') {
        setReactFlowNodes(nodes => 
          nodes.map(node => ({
            ...node,
            selected: node.id === selection.nodeId,
          }))
        );
      } else {
        setReactFlowNodes(nodes => 
          nodes.map(node => ({
            ...node,
            selected: false,
          }))
        );
      }
    }, [selection, setReactFlowNodes]);

    // Update edge selection state
    useEffect(() => {
      if (selection?.type === 'edge') {
        setReactFlowEdges(edges => 
          edges.map(edge => ({
            ...edge,
            selected: edge.id === selection.edgeId,
          }))
        );
      } else {
        setReactFlowEdges(edges => 
          edges.map(edge => ({
            ...edge,
            selected: false,
          }))
        );
      }
    }, [selection, setReactFlowEdges]);

    const proOptions = useMemo(() => ({ hideAttribution: true }), []);

    return (
      <div className={`relative bg-gray-50 ${className}`}>
        <ReactFlow
          nodes={reactFlowNodes}
          edges={reactFlowEdges}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
          onPaneClick={onPaneClick}
          onNodeDrag={onNodeDrag}
          onNodeDragStop={onNodeDragStop}
          nodeTypes={nodeTypes}
          fitView
          fitViewOptions={{
            padding: 0.1,
          }}
          defaultEdgeOptions={{
            type: 'bezier',
            style: {
              strokeWidth: 2,
            },
          }}
          proOptions={proOptions}
          nodesDraggable={layoutMode === 'manual'}
          nodesConnectable={false}
          elementsSelectable={true}
        >
          <Background color="#e5e7eb" gap={20} />
          <Controls
            position="top-left"
            className="bg-white border border-gray-200 shadow-sm"
          />
          <MiniMap
            position="bottom-right"
            className="bg-white border border-gray-200 shadow-sm"
            nodeColor="#f3f4f6"
            nodeStrokeColor="#d1d5db"
            nodeBorderRadius={4}
            maskColor="rgba(0, 0, 0, 0.1)"
            style={{
              backgroundColor: 'white',
            }}
          />
        </ReactFlow>
      </div>
    );
  }
);

GraphView.displayName = 'GraphView';

export default GraphView;
