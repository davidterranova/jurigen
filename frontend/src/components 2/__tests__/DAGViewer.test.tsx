import { render, screen } from '@testing-library/react';
import { DAGViewer } from '../DAGViewer';
import { useDAGTraversalStore } from '../../stores/dagTraversalStore';

// Mock the store
jest.mock('../../stores/dagTraversalStore');

describe('DAGViewer', () => {
  const mockStore = {
    currentDAG: null,
    getCurrentNode: jest.fn(() => null),
    selectAnswer: jest.fn(),
    goBack: jest.fn(),
    visitedNodes: [],
    isComplete: false,
    context: [],
    isLoading: false,
    error: null,
  };

  beforeEach(() => {
    (useDAGTraversalStore as unknown as jest.Mock).mockReturnValue(mockStore);
  });

  it('renders loading state', () => {
    (useDAGTraversalStore as unknown as jest.Mock).mockReturnValue({
      ...mockStore,
      isLoading: true,
    });

    render(<DAGViewer />);
    
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('renders error state', () => {
    const errorMessage = 'Failed to load DAG';
    (useDAGTraversalStore as unknown as jest.Mock).mockReturnValue({
      ...mockStore,
      error: errorMessage,
    });

    render(<DAGViewer />);
    
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('renders empty state when no DAG is loaded', () => {
    render(<DAGViewer />);
    
    expect(screen.getByText('No DAG loaded or no current node selected.')).toBeInTheDocument();
  });
});
