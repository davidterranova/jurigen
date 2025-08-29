package dag

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDAG(t *testing.T) {
	t.Parallel()

	t.Run("creates empty DAG with initialized fields", func(t *testing.T) {
		t.Parallel()

		d := NewDAG()

		assert.NotEqual(t, uuid.Nil, d.Id)
		assert.NotNil(t, d.Nodes)
		assert.Equal(t, 0, len(d.Nodes))
	})
}

func TestDAG_GetNode(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	testId1 := uuid.New()
	testId2 := uuid.New()

	node1 := Node{
		Id:       testId1,
		Question: "Test question 1?",
		Answers:  []Answer{},
	}

	node2 := Node{
		Id:       testId2,
		Question: "Test question 2?",
		Answers:  []Answer{},
	}

	d.Nodes[testId1] = node1
	d.Nodes[testId2] = node2

	tests := []struct {
		name     string
		nodeId   uuid.UUID
		expected Node
		wantErr  bool
	}{
		{
			name:     "returns existing node",
			nodeId:   testId1,
			expected: node1,
			wantErr:  false,
		},
		{
			name:     "returns another existing node",
			nodeId:   testId2,
			expected: node2,
			wantErr:  false,
		},
		{
			name:    "returns error for non-existent node",
			nodeId:  uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := d.GetNode(tt.nodeId)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDAG_GetRootNode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() DAG
		wantErr bool
		errMsg  string
	}{
		{
			name: "finds single root node",
			setup: func() DAG {
				d := NewDAG()
				rootId := uuid.New()
				childId := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Go to child",
							NextNode:  &childId,
						},
					},
				}

				child := Node{
					Id:       childId,
					Question: "Child question?",
					Answers:  []Answer{},
				}

				d.Nodes[rootId] = root
				d.Nodes[childId] = child
				return d
			},
		},
		{
			name: "returns error for empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			wantErr: true,
			errMsg:  "no root node found",
		},
		{
			name: "returns error when no root node found",
			setup: func() DAG {
				d := NewDAG()
				node1Id := uuid.New()
				node2Id := uuid.New()

				// Create a circular reference
				node1 := Node{
					Id:       node1Id,
					Question: "Node 1?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Go to node2",
							NextNode:  &node2Id,
						},
					},
				}

				node2 := Node{
					Id:       node2Id,
					Question: "Node 2?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Go to node1",
							NextNode:  &node1Id,
						},
					},
				}

				d.Nodes[node1Id] = node1
				d.Nodes[node2Id] = node2
				return d
			},
			wantErr: true,
			errMsg:  "no root node found",
		},
		{
			name: "returns error when multiple root nodes found",
			setup: func() DAG {
				d := NewDAG()
				root1Id := uuid.New()
				root2Id := uuid.New()

				root1 := Node{
					Id:       root1Id,
					Question: "Root 1?",
					Answers:  []Answer{},
				}

				root2 := Node{
					Id:       root2Id,
					Question: "Root 2?",
					Answers:  []Answer{},
				}

				d.Nodes[root1Id] = root1
				d.Nodes[root2Id] = root2
				return d
			},
			wantErr: true,
			errMsg:  "multiple root nodes found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := tt.setup()
			got, err := d.GetRootNode()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, got.Id)
		})
	}
}

func TestDAG_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func() DAG
		verify func(t *testing.T, data []byte)
	}{
		{
			name: "marshals empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			verify: func(t *testing.T, data []byte) {
				assert.Contains(t, string(data), `"nodes":[]`)
				assert.Contains(t, string(data), `"id"`)
			},
		},
		{
			name: "marshals DAG with single node",
			setup: func() DAG {
				d := NewDAG()
				nodeId := uuid.New()
				answerId := uuid.New()

				node := Node{
					Id:       nodeId,
					Question: "Test question?",
					Answers: []Answer{
						{
							Id:        answerId,
							Statement: "Test answer",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[nodeId] = node
				return d
			},
			verify: func(t *testing.T, data []byte) {
				assert.Contains(t, string(data), "Test question")
				assert.Contains(t, string(data), "Test answer")
				assert.Contains(t, string(data), `"id"`)
				assert.Contains(t, string(data), `"nodes"`)
			},
		},
		{
			name: "marshals DAG with multiple nodes",
			setup: func() DAG {
				d := NewDAG()
				node1Id := uuid.New()
				node2Id := uuid.New()

				node1 := Node{
					Id:       node1Id,
					Question: "First question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "First answer",
							NextNode:  &node2Id,
						},
					},
				}

				node2 := Node{
					Id:       node2Id,
					Question: "Second question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Second answer",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[node1Id] = node1
				d.Nodes[node2Id] = node2
				return d
			},
			verify: func(t *testing.T, data []byte) {
				jsonStr := string(data)
				assert.Contains(t, jsonStr, "First question")
				assert.Contains(t, jsonStr, "Second question")
				assert.Contains(t, jsonStr, "First answer")
				assert.Contains(t, jsonStr, "Second answer")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := tt.setup()
			data, err := d.MarshalJSON()

			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.verify != nil {
				tt.verify(t, data)
			}
		})
	}
}

func TestDAG_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr bool
		errMsg  string
		verify  func(t *testing.T, d DAG)
	}{
		{
			name: "unmarshals valid DAG JSON",
			data: `{
				"id": "550e8400-e29b-41d4-a716-446655440000",
				"nodes": [
					{
						"id": "8b007ce4-b676-5fb3-9f93-f5f6c41cb655",
						"question": "Test question?",
						"answers": [
							{
								"id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8",
								"answer": "Yes"
							}
						]
					}
				]
			}`,
			verify: func(t *testing.T, d DAG) {
				assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), d.Id)
				assert.Equal(t, 1, len(d.Nodes))

				nodeId := uuid.MustParse("8b007ce4-b676-5fb3-9f93-f5f6c41cb655")
				node, exists := d.Nodes[nodeId]
				assert.True(t, exists)
				assert.Equal(t, "Test question?", node.Question)
				assert.Equal(t, 1, len(node.Answers))
				assert.Equal(t, "Yes", node.Answers[0].Statement)

				// Test ParentNode pointer is set
				assert.NotNil(t, node.Answers[0].ParentNode)
				assert.Equal(t, nodeId, node.Answers[0].ParentNode.Id)
				assert.Equal(t, "Test question?", node.Answers[0].ParentNode.Question)
			},
		},
		{
			name: "unmarshals empty nodes",
			data: `{
				"id": "550e8400-e29b-41d4-a716-446655440001",
				"nodes": []
			}`,
			verify: func(t *testing.T, d DAG) {
				assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), d.Id)
				assert.Equal(t, 0, len(d.Nodes))
			},
		},
		{
			name:    "returns error for invalid JSON",
			data:    "invalid json",
			wantErr: true,
			errMsg:  "error unmarshalling DAG data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := NewDAG()
			err := d.UnmarshalJSON([]byte(tt.data))

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, d)
			}
		})
	}
}

func TestDAG_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func() DAG
		verify func(t *testing.T, result string)
	}{
		{
			name: "formats empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			verify: func(t *testing.T, result string) {
				assert.Equal(t, "", result)
			},
		},
		{
			name: "formats DAG with nodes and answers",
			setup: func() DAG {
				d := NewDAG()
				nodeId := uuid.New()

				node := Node{
					Id:       nodeId,
					Question: "Test question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Answer 1",
							NextNode:  nil,
						},
						{
							Id:        uuid.New(),
							Statement: "Answer 2",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[nodeId] = node
				return d
			},
			verify: func(t *testing.T, result string) {
				assert.Contains(t, result, "Question: Test question?")
				assert.Contains(t, result, "Answer: Answer 1")
				assert.Contains(t, result, "Answer: Answer 2")
				assert.Contains(t, result, "[LEAF]")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := tt.setup()
			result := d.String()

			if tt.verify != nil {
				tt.verify(t, result)
			}
		})
	}
}

func TestDAG_Walk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() (DAG, uuid.UUID)
		fnAnswer func(Node) (Answer, error)
		wantPath int // Expected path length
		wantErr  bool
		errMsg   string
	}{
		{
			name: "walks through simple path",
			setup: func() (DAG, uuid.UUID) {
				d := NewDAG()
				rootId := uuid.New()
				childId := uuid.New()
				answerId1 := uuid.New()
				answerId2 := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question?",
					Answers: []Answer{
						{
							Id:        answerId1,
							Statement: "Go to child",
							NextNode:  &childId,
						},
					},
				}

				child := Node{
					Id:       childId,
					Question: "Child question?",
					Answers: []Answer{
						{
							Id:        answerId2,
							Statement: "End here",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[rootId] = root
				d.Nodes[childId] = child

				// Set parent pointers manually for test
				rootCopy := d.Nodes[rootId]
				rootCopy.Answers[0].ParentNode = &rootCopy
				d.Nodes[rootId] = rootCopy

				childCopy := d.Nodes[childId]
				childCopy.Answers[0].ParentNode = &childCopy
				d.Nodes[childId] = childCopy

				return d, rootId
			},
			fnAnswer: func(node Node) (Answer, error) {
				// Always return the first answer
				return node.Answers[0], nil
			},
			wantPath: 2,
		},
		{
			name: "stops at leaf node",
			setup: func() (DAG, uuid.UUID) {
				d := NewDAG()
				rootId := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question?",
					Answers:  []Answer{}, // Leaf node
				}

				d.Nodes[rootId] = root
				return d, rootId
			},
			fnAnswer: func(node Node) (Answer, error) {
				return Answer{}, nil
			},
			wantPath: 0,
		},
		{
			name: "returns error for invalid node",
			setup: func() (DAG, uuid.UUID) {
				return NewDAG(), uuid.New()
			},
			fnAnswer: func(node Node) (Answer, error) {
				return Answer{}, nil
			},
			wantErr: true,
			errMsg:  "error getting node",
		},
		{
			name: "returns error when fnAnswer fails",
			setup: func() (DAG, uuid.UUID) {
				d := NewDAG()
				rootId := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Test answer",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[rootId] = root
				return d, rootId
			},
			fnAnswer: func(node Node) (Answer, error) {
				return Answer{}, assert.AnError
			},
			wantErr: true,
			errMsg:  "error getting answer",
		},
		{
			name: "returns error for invalid answer",
			setup: func() (DAG, uuid.UUID) {
				d := NewDAG()
				rootId := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question?",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Valid answer",
							NextNode:  nil,
						},
					},
				}

				d.Nodes[rootId] = root
				return d, rootId
			},
			fnAnswer: func(node Node) (Answer, error) {
				// Return an answer with different ID
				return Answer{Id: uuid.New(), Statement: "Invalid"}, nil
			},
			wantErr: true,
			errMsg:  "is not valid for node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d, startId := tt.setup()
			got, err := d.Walk(startId, tt.fnAnswer)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, len(got))
		})
	}
}

// TestDAG_WalkParentNodePointers tests that Walk returns answers with correct parent pointers
func TestDAG_WalkParentNodePointers(t *testing.T) {
	t.Parallel()

	// Create a DAG
	d := NewDAG()
	node1Id := uuid.New()
	node2Id := uuid.New()
	answer1Id := uuid.New()
	answer2Id := uuid.New()

	// Build the DAG structure
	d.Nodes[node1Id] = Node{
		Id:       node1Id,
		Question: "Do you like programming?",
		Answers: []Answer{
			{
				Id:        answer1Id,
				Statement: "Yes",
				NextNode:  &node2Id,
			},
		},
	}

	d.Nodes[node2Id] = Node{
		Id:       node2Id,
		Question: "What language do you prefer?",
		Answers: []Answer{
			{
				Id:        answer2Id,
				Statement: "Go",
				NextNode:  nil,
			},
		},
	}

	// Marshal and unmarshal to set parent pointers
	jsonData, err := d.MarshalJSON()
	require.NoError(t, err)

	testDAG := NewDAG()
	err = testDAG.UnmarshalJSON(jsonData)
	require.NoError(t, err)

	// Test Walk function with parent pointers
	answerFn := func(node Node) (Answer, error) {
		return node.Answers[0], nil
	}

	path, err := testDAG.Walk(node1Id, answerFn)
	require.NoError(t, err)
	require.Equal(t, 2, len(path))

	// Verify first answer has correct parent pointer
	firstAnswer := path[0]
	assert.Equal(t, "Yes", firstAnswer.Statement)
	assert.NotNil(t, firstAnswer.ParentNode)
	assert.Equal(t, "Do you like programming?", firstAnswer.ParentNode.Question)
	assert.Equal(t, node1Id, firstAnswer.ParentNode.Id)

	// Verify second answer has correct parent pointer
	secondAnswer := path[1]
	assert.Equal(t, "Go", secondAnswer.Statement)
	assert.NotNil(t, secondAnswer.ParentNode)
	assert.Equal(t, "What language do you prefer?", secondAnswer.ParentNode.Question)
	assert.Equal(t, node2Id, secondAnswer.ParentNode.Id)
}

// TestDAG_JSONRoundTrip tests the complete marshal/unmarshal cycle
func TestDAG_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("preserves DAG structure through JSON round trip", func(t *testing.T) {
		t.Parallel()

		// Setup original DAG
		original := NewDAG()
		nodeId := uuid.New()
		answerId := uuid.New()

		node := Node{
			Id:       nodeId,
			Question: "Test roundtrip?",
			Answers: []Answer{
				{
					Id:        answerId,
					Statement: "Yes",
					NextNode:  nil,
				},
			},
		}

		original.Nodes[nodeId] = node

		// Marshal
		data, err := original.MarshalJSON()
		require.NoError(t, err)

		// Unmarshal
		roundtrip := NewDAG()
		err = roundtrip.UnmarshalJSON(data)
		require.NoError(t, err)

		// Verify structure is preserved
		assert.Equal(t, original.Id, roundtrip.Id)
		assert.Equal(t, len(original.Nodes), len(roundtrip.Nodes))

		roundtripNode, exists := roundtrip.Nodes[nodeId]
		assert.True(t, exists)
		assert.Equal(t, node.Question, roundtripNode.Question)
		assert.Equal(t, len(node.Answers), len(roundtripNode.Answers))
		assert.Equal(t, node.Answers[0].Statement, roundtripNode.Answers[0].Statement)

		// Verify parent pointers are set correctly after unmarshal
		assert.NotNil(t, roundtripNode.Answers[0].ParentNode)
		assert.Equal(t, nodeId, roundtripNode.Answers[0].ParentNode.Id)
		assert.Equal(t, "Test roundtrip?", roundtripNode.Answers[0].ParentNode.Question)
	})
}
