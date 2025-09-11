package port

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"
	"davidterranova/jurigen/backend/internal/usecase"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileDAGRepository(t *testing.T) {
	testPath := "/tmp/test-repo"
	repo := NewFileDAGRepository(testPath)

	assert.NotNil(t, repo)
	assert.Equal(t, testPath, repo.filePath)
}

func TestFileDAGRepository_CreateAndGet(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	tests := []struct {
		name    string
		setup   func() (*FileDAGRepository, *model.DAG)
		wantErr bool
	}{
		{
			name: "successfully create and retrieve DAG",
			setup: func() (*FileDAGRepository, *model.DAG) {
				repo := NewFileDAGRepository(tempDir)
				testDAG := model.NewDAG("Test DAG")
				return repo, testDAG
			},
			wantErr: false,
		},
		{
			name: "create nil DAG should fail",
			setup: func() (*FileDAGRepository, *model.DAG) {
				repo := NewFileDAGRepository(tempDir)
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

			// Verify file was created
			expectedFile := filepath.Join(tempDir, testDAG.Id.String()+".json")
			assert.FileExists(t, expectedFile)

			// Retrieve the DAG
			retrievedDAG, err := repo.Get(ctx, testDAG.Id)
			require.NoError(t, err)
			assert.Equal(t, testDAG.Id, retrievedDAG.Id)
			assert.Equal(t, testDAG.Nodes, retrievedDAG.Nodes)
		})
	}
}

func TestFileDAGRepository_Create_AlreadyExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create the DAG first time
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Try to create the same DAG again - should fail
	err = repo.Create(ctx, testDAG)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrInvalidCommand)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFileDAGRepository_Get_NotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	nonExistentId := uuid.New()

	dagResult, err := repo.Get(ctx, nonExistentId)
	assert.Nil(t, dagResult)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestFileDAGRepository_Get_InvalidJSON(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testId := uuid.New()

	// Create a file with invalid JSON
	invalidFile := filepath.Join(tempDir, testId.String()+".json")
	err = os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	require.NoError(t, err)

	dagResult, err := repo.Get(ctx, testId)
	assert.Nil(t, dagResult)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrInternal)
}

func TestFileDAGRepository_Delete(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create a DAG first
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)

	// Delete the DAG
	err = repo.Delete(ctx, testDAG.Id)
	assert.NoError(t, err)

	// Verify file is gone
	expectedFile := filepath.Join(tempDir, testDAG.Id.String()+".json")
	assert.NoFileExists(t, expectedFile)

	// Verify it's gone from repository
	_, err = repo.Get(ctx, testDAG.Id)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestFileDAGRepository_Delete_NotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	nonExistentId := uuid.New()

	err = repo.Delete(ctx, nonExistentId)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestFileDAGRepository_List(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()

	// Initially empty
	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 0)

	// Add some DAGs
	dag1 := model.NewDAG("Test DAG")
	dag2 := model.NewDAG("Test DAG")

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

func TestFileDAGRepository_List_WithInvalidFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()

	// Create some valid DAG files
	validDAG := model.NewDAG("Test DAG")
	err = repo.Create(ctx, validDAG)
	require.NoError(t, err)

	// Create some invalid files that should be ignored
	invalidFiles := []string{
		"not-uuid.json",
		"invalid-uuid-format.json",
		"valid-uuid-but-wrong-ext.txt",
	}

	for _, filename := range invalidFiles {
		// Create invalid file
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Create a subdirectory (should be ignored)
	err = os.MkdirAll(filepath.Join(tempDir, "subdirectory"), 0755)
	require.NoError(t, err)

	// List should only return the valid DAG
	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Contains(t, ids, validDAG.Id)
}

func TestFileDAGRepository_List_DirectoryNotFound(t *testing.T) {
	// Use non-existent directory
	nonExistentDir := "/tmp/non-existent-dag-repo-test"
	repo := NewFileDAGRepository(nonExistentDir)
	ctx := context.Background()

	ids, err := repo.List(ctx)
	assert.Nil(t, ids)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading directory")
}

func TestFileDAGRepository_Update(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create a DAG first
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that modifies the DAG
	updateFn := func(d model.DAG) (model.DAG, error) {
		// Add a test node to the DAG
		testNode := model.Node{
			Id:       uuid.New(),
			Question: "Updated question",
			Answers:  []model.Answer{},
		}
		d.Nodes[testNode.Id] = testNode
		return d, nil
	}

	// Apply the update
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

func TestFileDAGRepository_Update_NotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	nonExistentId := uuid.New()

	updateFn := func(d model.DAG) (model.DAG, error) {
		return d, nil
	}

	err = repo.Update(ctx, nonExistentId, updateFn)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrNotFound)
}

func TestFileDAGRepository_Update_FunctionFails(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create a DAG
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that returns an error
	updateFn := func(d model.DAG) (model.DAG, error) {
		return model.DAG{}, assert.AnError
	}

	err = repo.Update(ctx, testDAG.Id, updateFn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update function failed")
}

func TestFileDAGRepository_Update_IDChange(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create a DAG
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Update function that tries to change the ID
	updateFn := func(d model.DAG) (model.DAG, error) {
		d.Id = uuid.New() // This should cause an error
		return d, nil
	}

	err = repo.Update(ctx, testDAG.Id, updateFn)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrInvalidCommand)
	assert.Contains(t, err.Error(), "update function cannot change DAG ID")
}

func TestFileDAGRepository_CreateDirectoryIfNotExists(t *testing.T) {
	// Use a nested directory path that doesn't exist
	tempBase, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempBase)
		require.NoError(t, err)
	}()

	nestedDir := filepath.Join(tempBase, "nested", "directory", "structure")
	repo := NewFileDAGRepository(nestedDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create should create the directory structure
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Verify directory was created
	assert.DirExists(t, nestedDir)

	// Verify file was created
	expectedFile := filepath.Join(nestedDir, testDAG.Id.String()+".json")
	assert.FileExists(t, expectedFile)
}

func TestFileDAGRepository_DAGWithComplexStructure(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()

	// Create a complex DAG with nodes and answers
	testDAG := model.NewDAG("Test DAG")

	// Root node
	rootNode := model.Node{
		Id:       uuid.New(),
		Question: "What is your primary concern?",
		Answers: []model.Answer{
			{
				Id:        uuid.New(),
				Statement: "Financial issues",
				NextNode:  nil, // This will be set to the next node
			},
			{
				Id:        uuid.New(),
				Statement: "Legal issues",
				NextNode:  nil,
			},
		},
	}

	// Secondary node
	secondaryNode := model.Node{
		Id:       uuid.New(),
		Question: "How severe is the financial impact?",
		Answers: []model.Answer{
			{
				Id:          uuid.New(),
				Statement:   "Minor impact",
				NextNode:    nil, // Leaf node
				UserContext: "Additional context",
				Metadata:    map[string]interface{}{"severity": "low"},
			},
			{
				Id:        uuid.New(),
				Statement: "Major impact",
				NextNode:  nil, // Leaf node
				Metadata:  map[string]interface{}{"severity": "high"},
			},
		},
	}

	// Link the nodes
	rootNode.Answers[0].NextNode = &secondaryNode.Id

	// Add nodes to DAG
	testDAG.Nodes[rootNode.Id] = rootNode
	testDAG.Nodes[secondaryNode.Id] = secondaryNode

	// Create the DAG
	err = repo.Create(ctx, testDAG)
	require.NoError(t, err)

	// Retrieve and verify the complex structure
	retrievedDAG, err := repo.Get(ctx, testDAG.Id)
	require.NoError(t, err)

	assert.Equal(t, testDAG.Id, retrievedDAG.Id)
	assert.Len(t, retrievedDAG.Nodes, 2)

	// Verify root node
	retrievedRoot, err := retrievedDAG.GetNode(rootNode.Id)
	require.NoError(t, err)
	assert.Equal(t, "What is your primary concern?", retrievedRoot.Question)
	assert.Len(t, retrievedRoot.Answers, 2)

	// Verify the link between nodes
	assert.NotNil(t, retrievedRoot.Answers[0].NextNode)
	assert.Equal(t, secondaryNode.Id, *retrievedRoot.Answers[0].NextNode)

	// Verify secondary node with metadata
	retrievedSecondary, err := retrievedDAG.GetNode(secondaryNode.Id)
	require.NoError(t, err)
	assert.Equal(t, "How severe is the financial impact?", retrievedSecondary.Question)
	assert.Len(t, retrievedSecondary.Answers, 2)

	// Check metadata and user context preservation
	minorImpactAnswer := retrievedSecondary.Answers[0]
	assert.Equal(t, "Additional context", minorImpactAnswer.UserContext)
	assert.Equal(t, "low", minorImpactAnswer.Metadata["severity"])
}

func TestFileDAGRepository_RoundTripConsistency(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	originalDAG := model.NewDAG("Test DAG")

	// Add a complex node structure
	node := model.Node{
		Id:       uuid.New(),
		Question: "Test question with unicode: αβγ 中文 🚀",
		Answers: []model.Answer{
			{
				Id:          uuid.New(),
				Statement:   "Answer with special chars: !@#$%^&*()",
				UserContext: "Context with newlines\nand\ttabs",
				Metadata: map[string]interface{}{
					"number":  42.5,
					"boolean": true,
					"array":   []string{"a", "b", "c"},
					"nested":  map[string]string{"key": "value"},
				},
			},
		},
	}
	originalDAG.Nodes[node.Id] = node

	// Create -> Get -> Update -> Get cycle
	err = repo.Create(ctx, originalDAG)
	require.NoError(t, err)

	firstRetrieved, err := repo.Get(ctx, originalDAG.Id)
	require.NoError(t, err)

	// Update with a function that adds another node
	updateFn := func(d model.DAG) (model.DAG, error) {
		newNode := model.Node{
			Id:       uuid.New(),
			Question: "Updated question",
			Answers:  []model.Answer{},
		}
		d.Nodes[newNode.Id] = newNode
		return d, nil
	}

	err = repo.Update(ctx, originalDAG.Id, updateFn)
	require.NoError(t, err)

	finalDAG, err := repo.Get(ctx, originalDAG.Id)
	require.NoError(t, err)

	// Verify the original node is preserved
	originalNode := finalDAG.Nodes[node.Id]
	assert.Equal(t, node.Question, originalNode.Question)
	assert.Equal(t, node.Answers[0].Statement, originalNode.Answers[0].Statement)
	assert.Equal(t, node.Answers[0].UserContext, originalNode.Answers[0].UserContext)

	// Check metadata - JSON unmarshaling changes types, so check values individually
	assert.Equal(t, 42.5, originalNode.Answers[0].Metadata["number"])
	assert.Equal(t, true, originalNode.Answers[0].Metadata["boolean"])

	// Array becomes []interface{} after JSON round trip
	arrayValue, ok := originalNode.Answers[0].Metadata["array"].([]interface{})
	assert.True(t, ok, "Expected array to be []interface{}")
	assert.Len(t, arrayValue, 3)
	assert.Equal(t, "a", arrayValue[0])
	assert.Equal(t, "b", arrayValue[1])
	assert.Equal(t, "c", arrayValue[2])

	// Nested map becomes map[string]interface{} after JSON round trip
	nestedValue, ok := originalNode.Answers[0].Metadata["nested"].(map[string]interface{})
	assert.True(t, ok, "Expected nested to be map[string]interface{}")
	assert.Equal(t, "value", nestedValue["key"])

	// Verify the new node was added
	assert.Len(t, finalDAG.Nodes, 2)

	// Verify the first retrieved DAG is consistent
	assert.Equal(t, firstRetrieved.Id, finalDAG.Id)
}

func TestFileDAGRepository_ErrorHandling_FilePermissions(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	// Make the directory read-only
	err = os.Chmod(tempDir, 0444)
	require.NoError(t, err)
	defer func() {
		err := os.Chmod(tempDir, 0755) // Restore permissions for cleanup
		require.NoError(t, err)
	}()

	repo := NewFileDAGRepository(tempDir)
	ctx := context.Background()
	testDAG := model.NewDAG("Test DAG")

	// Create should fail due to permission issues
	err = repo.Create(ctx, testDAG)
	assert.Error(t, err)
	assert.ErrorIs(t, err, usecase.ErrInternal)
}

func TestFileDAGRepository_InterfaceCompliance(t *testing.T) {
	// Compile-time check that FileDAGRepository implements DAGRepository
	tempDir, err := os.MkdirTemp("", "dag-repo-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	var _ usecase.DAGRepository = NewFileDAGRepository(tempDir)
}
