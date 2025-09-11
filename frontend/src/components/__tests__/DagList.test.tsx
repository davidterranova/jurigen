import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DagList } from '../DagList';
import { useDAGStore } from '../../stores/dagStore';
import { DAGService } from '../../services/dagService';

// Mock the store
vi.mock('../../stores/dagStore');
const mockUseDAGStore = vi.mocked(useDAGStore);

// Mock the service
vi.mock('../../services/dagService');
const mockDAGService = vi.mocked(DAGService);

const mockDAGs = [
  {
    id: '1',
    name: 'Legal Case Builder',
    description: 'Build legal cases step by step',
    node_count: 5,
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Contract Review',
    description: 'Review contracts for compliance',
    node_count: 8,
    updated_at: '2024-01-02T00:00:00Z',
  },
];

describe('DagList', () => {
  const mockStore = {
    availableDAGs: [],
    selectedDAGId: null,
    setAvailableDAGs: vi.fn(),
    selectDAG: vi.fn(),
    isLoading: false,
    setLoading: vi.fn(),
    setError: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseDAGStore.mockReturnValue(mockStore);
    mockDAGService.listDAGs.mockResolvedValue(mockDAGs);
  });

  it('should render DAG list title', () => {
    render(<DagList />);
    expect(screen.getByText('DAG Library')).toBeInTheDocument();
  });

  it('should render search input', () => {
    render(<DagList />);
    expect(screen.getByPlaceholderText('Search DAGs...')).toBeInTheDocument();
  });

  it('should display loading skeletons when loading', () => {
    mockUseDAGStore.mockReturnValue({ ...mockStore, isLoading: true });
    render(<DagList />);
    
    // Check for loading skeletons (they have animate-pulse class)
    const loadingElements = document.querySelectorAll('.animate-pulse');
    expect(loadingElements.length).toBeGreaterThan(0);
  });

  it('should display empty state when no DAGs available', () => {
    mockUseDAGStore.mockReturnValue({ 
      ...mockStore, 
      availableDAGs: [],
      isLoading: false 
    });
    
    render(<DagList />);
    expect(screen.getByText('No DAGs available.')).toBeInTheDocument();
  });

  it('should render DAG items when available', () => {
    mockUseDAGStore.mockReturnValue({ 
      ...mockStore, 
      availableDAGs: mockDAGs,
      isLoading: false 
    });
    
    render(<DagList />);
    
    expect(screen.getByText('Legal Case Builder')).toBeInTheDocument();
    expect(screen.getByText('Contract Review')).toBeInTheDocument();
  });

  it('should filter DAGs based on search term', () => {
    mockUseDAGStore.mockReturnValue({ 
      ...mockStore, 
      availableDAGs: mockDAGs,
      isLoading: false 
    });
    
    render(<DagList />);
    
    const searchInput = screen.getByPlaceholderText('Search DAGs...');
    fireEvent.change(searchInput, { target: { value: 'Legal' } });
    
    expect(screen.getByText('Legal Case Builder')).toBeInTheDocument();
    expect(screen.queryByText('Contract Review')).not.toBeInTheDocument();
  });

  it('should select DAG when clicked', () => {
    mockUseDAGStore.mockReturnValue({ 
      ...mockStore, 
      availableDAGs: mockDAGs,
      isLoading: false 
    });
    
    render(<DagList />);
    
    const dagButton = screen.getByText('Legal Case Builder').closest('button');
    fireEvent.click(dagButton!);
    
    expect(mockStore.selectDAG).toHaveBeenCalledWith('1');
  });

  it('should highlight selected DAG', () => {
    mockUseDAGStore.mockReturnValue({ 
      ...mockStore, 
      availableDAGs: mockDAGs,
      selectedDAGId: '1',
      isLoading: false 
    });
    
    render(<DagList />);
    
    const selectedButton = screen.getByText('Legal Case Builder').closest('button');
    expect(selectedButton).toHaveClass('bg-blue-50');
  });

  it('should load DAGs on mount', async () => {
    render(<DagList />);
    
    await waitFor(() => {
      expect(mockDAGService.listDAGs).toHaveBeenCalled();
    });
  });
});
