package port

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHybridDAGRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	config := HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	}

	repo := NewHybridDAGRepository(config)

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.fileRepo)
	assert.NotNil(t, repo.memoryRepo)
	assert.True(t, repo.writeThrough)
}

func TestHybridDAGRepository_Initialize_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Should have no DAGs in memory
	dagIds, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, dagIds)
}

func TestHybridDAGRepository_Initialize_WithExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create some DAG files manually
	testDAGs := createTestDAGs(t, 3)
	for _, testDAG := range testDAGs {
		data, err := testDAG.MarshalJSON()
		require.NoError(t, err)

		filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
		err = os.WriteFile(filename, data, 0644)
		require.NoError(t, err)
	}

	// Initialize repository
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Should have all DAGs loaded in memory
	dagIds, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(dagIds))

	// Verify each DAG can be retrieved
	for _, testDAG := range testDAGs {
		retrievedDAG, err := repo.Get(ctx, testDAG.Id)
		require.NoError(t, err)
		assert.Equal(t, testDAG.Id, retrievedDAG.Id)
	}
}

func TestHybridDAGRepository_CreateWithWriteThrough(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Create a new DAG
	testDAG := createTestDAG(t)
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Should be in memory
	retrievedFromMemory, err := repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	assert.Equal(t, testDAG.Id, retrievedFromMemory.Id)

	// Should also be in file
	filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
	assert.FileExists(t, filename)

	// Verify file contents by loading directly
	retrievedFromFile, err := repo.fileRepo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	assert.Equal(t, testDAG.Id, retrievedFromFile.Id)
}

func TestHybridDAGRepository_CreateWithoutWriteThrough(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: false, // Disabled
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Create a new DAG
	testDAG := createTestDAG(t)
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Should be in memory
	retrievedFromMemory, err := repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	assert.Equal(t, testDAG.Id, retrievedFromMemory.Id)

	// Should NOT be in file
	filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
	assert.NoFileExists(t, filename)
}

func TestHybridDAGRepository_UpdateWithWriteThrough(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Create a DAG
	testDAG := createTestDAG(t)
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update the DAG
	newNodeID := uuid.New()
	err = repo.Update(ctx, testDAG.Id, func(existing dag.DAG) (dag.DAG, error) {
		// Add a new node
		existing.Nodes[newNodeID] = dag.Node{
			Id:       newNodeID,
			Question: "Updated question",
			Answers:  []dag.Answer{},
		}
		return existing, nil
	})
	require.NoError(t, err)

	// Verify update in memory
	updatedDAG, err := repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	_, exists := updatedDAG.Nodes[newNodeID]
	assert.True(t, exists)

	// Verify update in file
	fileDAG, err := repo.fileRepo.Get(ctx, testDAG.Id)
	require.NoError(t, err)
	_, exists = fileDAG.Nodes[newNodeID]
	assert.True(t, exists)
}

func TestHybridDAGRepository_Delete(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Create a DAG
	testDAG := createTestDAG(t)
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)

	// Delete the DAG
	err = repo.Delete(ctx, testDAG.Id)
	require.NoError(t, err)

	// Should not be in memory
	_, err = repo.Get(ctx, testDAG.Id)
	assert.ErrorIs(t, err, usecase.ErrNotFound)

	// Should not be in file
	filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
	assert.NoFileExists(t, filename)
}

func TestHybridDAGRepository_Sync(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: false, // Disabled for this test
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Create several DAGs in memory only
	testDAGs := createTestDAGs(t, 3)
	for _, testDAG := range testDAGs {
		err = repo.Create(ctx, testDAG)
		require.NoError(t, err)
	}

	// Verify files don't exist yet
	for _, testDAG := range testDAGs {
		filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
		assert.NoFileExists(t, filename)
	}

	// Sync to files
	err = repo.Sync(ctx)
	require.NoError(t, err)

	// Verify files now exist
	for _, testDAG := range testDAGs {
		filename := filepath.Join(tempDir, testDAG.Id.String()+".json")
		assert.FileExists(t, filename)
	}
}

func TestHybridDAGRepository_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	logger := zerolog.Nop()
	repo := NewHybridDAGRepository(HybridDAGRepositoryConfig{
		FilePath:     tempDir,
		WriteThrough: true,
		Logger:       &logger,
	})

	ctx := context.Background()
	err := repo.Initialize(ctx)
	require.NoError(t, err)

	// Initial stats
	stats, err := repo.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.MemoryDAGCount)
	assert.Equal(t, 0, stats.FileDAGCount)
	assert.True(t, stats.WriteThrough)

	// Create some DAGs
	testDAGs := createTestDAGs(t, 2)
	for _, testDAG := range testDAGs {
		err = repo.Create(ctx, testDAG)
		require.NoError(t, err)
	}

	// Updated stats
	stats, err = repo.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.MemoryDAGCount)
	assert.Equal(t, 2, stats.FileDAGCount)
}

// Helper functions for tests

func createTestDAG(_ *testing.T) *dag.DAG {
	dagID := uuid.New()
	nodeID := uuid.New()
	answerID := uuid.New()

	return &dag.DAG{
		Id: dagID,
		Nodes: map[uuid.UUID]dag.Node{
			nodeID: {
				Id:       nodeID,
				Question: "Test question?",
				Answers: []dag.Answer{
					{
						Id:        answerID,
						Statement: "Test answer",
						NextNode:  nil,
					},
				},
			},
		},
	}
}

func createTestDAGs(t *testing.T, count int) []*dag.DAG {
	dags := make([]*dag.DAG, count)
	for i := 0; i < count; i++ {
		dags[i] = createTestDAG(t)
	}
	return dags
}
