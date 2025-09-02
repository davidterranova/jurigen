package port

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryDAGRepository(t *testing.T) {
	repo := NewInMemoryDAGRepository()

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.dags)
	assert.Len(t, repo.dags, 0)
}

func TestInMemoryDAGRepository_CreateAndGet(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*InMemoryDAGRepository, *dag.DAG)
		wantErr bool
	}{
		{
			name: "successfully create and retrieve DAG",
			setup: func() (*InMemoryDAGRepository, *dag.DAG) {
				repo := NewInMemoryDAGRepository()
				testDAG := dag.NewDAG()
				return repo, testDAG
			},
			wantErr: false,
		},
		{
			name: "create nil DAG should fail",
			setup: func() (*InMemoryDAGRepository, *dag.DAG) {
				repo := NewInMemoryDAGRepository()
				return repo, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, testDAG := tt.setup()
			ctx := context.Background()

			if testDAG == nil {
				// Test creating nil DAG
				err := repo.Create(ctx, testDAG)
				assert.Error(t, err)
				assert.ErrorIs(t, err, usecase.ErrInvalidCommand)
				return
			}

			// Create the DAG
			err := repo.Create(ctx, testDAG)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Retrieve the DAG
			retrievedDAG, err := repo.Get(ctx, testDAG.Id)
			require.NoError(t, err)
			assert.Equal(t, testDAG.Id, retrievedDAG.Id)
			assert.Equal(t, testDAG.Nodes, retrievedDAG.Nodes)
		})
	}
}

func TestInMemoryDAGRepository_Get_NotFound(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	nonExistentId := uuid.New()

	dag, err := repo.Get(ctx, nonExistentId)
	assert.Nil(t, dag)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestInMemoryDAGRepository_Delete(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	testDAG := dag.NewDAG()

	// Create a DAG first
	err := repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)

	// Delete the DAG
	err = repo.Delete(ctx, testDAG.Id)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = repo.Get(ctx, testDAG.Id)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestInMemoryDAGRepository_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	nonExistentId := uuid.New()

	err := repo.Delete(ctx, nonExistentId)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestInMemoryDAGRepository_List(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()

	// Initially empty
	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 0)

	// Add some DAGs
	dag1 := dag.NewDAG()
	dag2 := dag.NewDAG()

	err = repo.Create(ctx, dag1)
	require.NoError(t, err)
	err = repo.Create(ctx, dag2)
	require.NoError(t, err)

	// List should return both IDs
	ids, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, dag1.Id)
	assert.Contains(t, ids, dag2.Id)
}

func TestInMemoryDAGRepository_Update(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	testDAG := dag.NewDAG()

	// Create a DAG first
	err := repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that modifies the DAG
	updateFn := func(d dag.DAG) (dag.DAG, error) {
		// Add a test node to the DAG
		testNode := dag.Node{
			Id:       uuid.New(),
			Question: "Updated question",
			Answers:  []dag.Answer{},
		}
		d.Nodes[testNode.Id] = testNode
		return d, nil
	}

	// Apply the update with the specific DAG ID
	err = repo.Update(ctx, testDAG.Id, updateFn)
	assert.NoError(t, err)

	// Verify the DAG was updated
	updatedDAG, err := repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	assert.Len(t, updatedDAG.Nodes, 1)

	// Find the test node
	var foundTestNode bool
	for _, node := range updatedDAG.Nodes {
		if node.Question == "Updated question" {
			foundTestNode = true
			break
		}
	}
	assert.True(t, foundTestNode, "Expected to find the updated test node")
}

func TestInMemoryDAGRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	nonExistentId := uuid.New()

	updateFn := func(d dag.DAG) (dag.DAG, error) {
		return d, nil
	}

	err := repo.Update(ctx, nonExistentId, updateFn)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestInMemoryDAGRepository_Update_FunctionFails(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	testDAG := dag.NewDAG()

	// Create a DAG
	err := repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that returns an error
	updateFn := func(d dag.DAG) (dag.DAG, error) {
		return dag.DAG{}, fmt.Errorf("update function error")
	}

	err = repo.Update(ctx, testDAG.Id, updateFn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update function failed")
}

func TestInMemoryDAGRepository_Update_IDChange(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()
	testDAG := dag.NewDAG()

	// Create a DAG
	err := repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that tries to change the ID
	updateFn := func(d dag.DAG) (dag.DAG, error) {
		d.Id = uuid.New() // This should cause an error
		return d, nil
	}

	err = repo.Update(ctx, testDAG.Id, updateFn)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrInvalidCommand)
	assert.Contains(t, err.Error(), "update function cannot change DAG ID")
}

func TestInMemoryDAGRepository_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryDAGRepository()
	ctx := context.Background()

	const numGoroutines = 10
	const numDAGsPerGoroutine = 5

	// Test concurrent creates and reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numDAGsPerGoroutine; j++ {
				testDAG := dag.NewDAG()
				err := repo.Create(ctx, testDAG)
				assert.NoError(t, err)

				retrievedDAG, err := repo.Get(ctx, testDAG.Id)
				assert.NoError(t, err)
				assert.Equal(t, testDAG.Id, retrievedDAG.Id)
			}
		}()
	}
}
