package dag

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDAG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want DAG
	}{
		{
			name: "creates empty DAG with initialized fields",
			want: DAG{
				Nodes: make(map[uuid.UUID]Node),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewDAG()

			assert.NotNil(t, got.Nodes)
			assert.Equal(t, 0, len(got.Nodes))
		})
	}
}

func TestDAG_GetNode(t *testing.T) {
	t.Parallel()

	// Setup test data
	nodeId1 := uuid.New()
	nodeId2 := uuid.New()
	nonExistentId := uuid.New()

	node1 := Node{
		Id:       nodeId1,
		Question: "Test question 1",
		Answers:  []Answer{},
	}

	node2 := Node{
		Id:       nodeId2,
		Question: "Test question 2",
		Answers:  []Answer{},
	}

	d := NewDAG()
	d.Nodes[nodeId1] = node1
	d.Nodes[nodeId2] = node2

	tests := []struct {
		name    string
		dag     DAG
		id      uuid.UUID
		want    Node
		wantErr bool
		errMsg  string
	}{
		{
			name: "returns existing node",
			dag:  d,
			id:   nodeId1,
			want: node1,
		},
		{
			name: "returns another existing node",
			dag:  d,
			id:   nodeId2,
			want: node2,
		},
		{
			name:    "returns error for non-existent node",
			dag:     d,
			id:      nonExistentId,
			wantErr: true,
			errMsg:  "node not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.dag.GetNode(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDAG_GetRootNode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() DAG
		want    Node
		wantErr bool
		errMsg  string
	}{
		{
			name: "finds single root node",
			setup: func() DAG {
				d := NewDAG()
				rootId := uuid.New()
				childId := uuid.New()
				answerId := uuid.New()

				root := Node{
					Id:       rootId,
					Question: "Root question",
					Answers: []Answer{
						{
							Id:        answerId,
							Statement: "Go to child",
							NextNode:  &childId,
						},
					},
				}

				child := Node{
					Id:       childId,
					Question: "Child question",
					Answers:  []Answer{},
				}

				d.Nodes[rootId] = root
				d.Nodes[childId] = child
				return d
			},
			want: Node{
				Id:       uuid.UUID{},
				Question: "Root question",
			},
		},
		{
			name: "returns error when no root node found",
			setup: func() DAG {
				d := NewDAG()
				id1 := uuid.New()
				id2 := uuid.New()
				answerId := uuid.New()

				// Create circular reference
				node1 := Node{
					Id:       id1,
					Question: "Node 1",
					Answers: []Answer{
						{
							Id:        answerId,
							Statement: "Go to node 2",
							NextNode:  &id2,
						},
					},
				}

				node2 := Node{
					Id:       id2,
					Question: "Node 2",
					Answers: []Answer{
						{
							Id:        uuid.New(),
							Statement: "Go to node 1",
							NextNode:  &id1,
						},
					},
				}

				d.Nodes[id1] = node1
				d.Nodes[id2] = node2
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
					Question: "Root 1",
					Answers:  []Answer{},
				}

				root2 := Node{
					Id:       root2Id,
					Question: "Root 2",
					Answers:  []Answer{},
				}

				d.Nodes[root1Id] = root1
				d.Nodes[root2Id] = root2
				return d
			},
			wantErr: true,
			errMsg:  "multiple root nodes found",
		},
		{
			name: "returns error for empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			wantErr: true,
			errMsg:  "no root node found",
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
			assert.Equal(t, tt.want.Question, got.Question)
		})
	}
}

func TestDAG_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() DAG
		want    string // We'll check that it contains expected elements
		wantErr bool
	}{
		{
			name: "marshals empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			want: "[]",
		},
		{
			name: "marshals DAG with single node",
			setup: func() DAG {
				d := NewDAG()
				nodeId := uuid.New()
				answerId := uuid.New()

				node := Node{
					Id:       nodeId,
					Question: "Test question",
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
			want: "Test question",
		},
		{
			name: "marshals DAG with multiple nodes",
			setup: func() DAG {
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
				return d
			},
			want: "Root question",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := tt.setup()
			got, err := d.MarshalJSON()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, string(got), tt.want)

			// Verify it's valid JSON
			var result interface{}
			err = json.Unmarshal(got, &result)
			assert.NoError(t, err)
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
		name  string
		setup func() DAG
		want  []string // Strings that should be contained in output
	}{
		{
			name: "formats empty DAG",
			setup: func() DAG {
				return NewDAG()
			},
			want: []string{},
		},
		{
			name: "formats DAG with nodes and answers",
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
			want: []string{
				"Question: Test question?",
				"Answer: Test answer",
				"[LEAF]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := tt.setup()
			got := d.String()

			for _, want := range tt.want {
				assert.Contains(t, got, want)
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
				return Answer{
					Id:        uuid.New(),
					Statement: "Invalid answer",
					NextNode:  nil,
				}, nil
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

// TestDAG_JSONRoundTrip tests the complete marshal/unmarshal cycle
func TestDAG_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("preserves DAG structure through JSON round trip", func(t *testing.T) {
		t.Parallel()

		// Setup original DAG
		original := NewDAG()
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

		original.Nodes[rootId] = root
		original.Nodes[childId] = child

		// Marshal to JSON
		data, err := original.MarshalJSON()
		require.NoError(t, err)

		// Unmarshal to new DAG
		restored := NewDAG()
		err = restored.UnmarshalJSON(data)
		require.NoError(t, err)

		// Verify structure is preserved
		assert.Equal(t, len(original.Nodes), len(restored.Nodes))

		for id, originalNode := range original.Nodes {
			restoredNode, exists := restored.Nodes[id]
			assert.True(t, exists)
			assert.Equal(t, originalNode.Question, restoredNode.Question)
			assert.Equal(t, len(originalNode.Answers), len(restoredNode.Answers))
		}
	})
}
