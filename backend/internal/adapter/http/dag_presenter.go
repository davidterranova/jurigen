package http

import (
	"davidterranova/jurigen/backend/internal/model"

	"github.com/google/uuid"
)

// DAGPresenter represents a complete Legal Case DAG structure for API responses
//
// @Description Legal Case DAG with questions, answers, and context
// @Example {"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Employment Discrimination Case", "nodes": [{"id": "8b007ce4-b676-5fb3-9f93-f5f6c41cb655", "question": "Were you discriminated against?", "answers": [{"id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8", "answer": "Yes, age discrimination occurred", "user_context": "Manager explicitly mentioned my age", "metadata": {"confidence": 0.9, "tags": ["age_discrimination"]}}]}]}
type DAGPresenter struct {
	Id    uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Unique identifier for the Legal Case DAG"`
	Title string          `json:"title" example:"Employment Discrimination Case" description:"Human-readable title describing the legal case context"`
	Nodes []NodePresenter `json:"nodes" description:"Array of question nodes that make up the legal case decision tree"`
}

func NewDAGPresenter(dag *model.DAG) DAGPresenter {
	nodes := make([]NodePresenter, 0, len(dag.Nodes))
	for _, node := range dag.Nodes {
		nodes = append(nodes, NewNodePresenter(node))
	}

	return DAGPresenter{
		Id:    dag.Id,
		Title: dag.Title,
		Nodes: nodes,
	}
}

// NodePresenter represents a question node in the Legal Case DAG
//
// @Description A question node with potential answers for legal case context building
// @Example {"id": "8b007ce4-b676-5fb3-9f93-f5f6c41cb655", "question": "Were you discriminated against?", "answers": [{"id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8", "answer": "Yes", "user_context": "Manager made age-related comments"}]}
type NodePresenter struct {
	Id       uuid.UUID         `json:"id" example:"8b007ce4-b676-5fb3-9f93-f5f6c41cb655" description:"Unique identifier for the question node"`
	Question string            `json:"question" example:"Were you discriminated against in the workplace?" description:"The legal question being asked"`
	Answers  []AnswerPresenter `json:"answers" description:"Available answer options for this question"`
}

func NewNodePresenter(node model.Node) NodePresenter {
	answers := make([]AnswerPresenter, 0, len(node.Answers))
	for _, answer := range node.Answers {
		answers = append(answers, NewAnswerPresenter(answer))
	}

	np := NodePresenter{
		Id:       node.Id,
		Question: node.Question,
		Answers:  answers,
	}

	return np
}

// AnswerPresenter represents an answer option with optional legal context
//
// @Description An answer to a legal question with optional user context and structured metadata for evidence tracking
// @Example {"id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8", "answer": "Yes, age discrimination occurred", "user_context": "Manager explicitly mentioned my age during termination", "metadata": {"confidence": 0.9, "severity": "high", "tags": ["age_discrimination", "wrongful_termination"], "sources": ["HR_Email.pdf", "Witness_Statement.pdf"], "damages_estimate": 75000}}
type AnswerPresenter struct {
	Id          uuid.UUID              `json:"id" example:"fc28c4b6-d185-cf56-a7e4-dead499ff1e8" description:"Unique identifier for the answer"`
	Statement   string                 `json:"answer" example:"Yes, age discrimination occurred" description:"The answer statement or response"`
	NextNode    *uuid.UUID             `json:"next_node,omitempty" example:"8b007ce4-b676-5fb3-9f93-f5f6c41cb655" description:"ID of the next node to navigate to (null for leaf nodes)"`
	UserContext string                 `json:"user_context,omitempty" example:"Manager explicitly mentioned my age during termination" description:"Free-form user notes and context for this answer"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Structured metadata for legal assessment: confidence scores, evidence tracking, damages estimates, action items, etc."`
}

func NewAnswerPresenter(answer model.Answer) AnswerPresenter {
	return AnswerPresenter{
		Id:          answer.Id,
		Statement:   answer.Statement,
		NextNode:    answer.NextNode,
		UserContext: answer.UserContext,
		Metadata:    answer.Metadata,
	}
}

// DAGListPresenter represents a list of Legal Case DAG identifiers for API responses
//
// @Description List of Legal Case DAG identifiers available in the system
// @Example {"dags": ["550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"], "count": 2}
type DAGListPresenter struct {
	DAGs  []uuid.UUID `json:"dags" description:"Array of DAG unique identifiers"`
	Count int         `json:"count" description:"Total number of DAGs available"`
}

func NewDAGListPresenter(dagIds []uuid.UUID) DAGListPresenter {
	return DAGListPresenter{
		DAGs:  dagIds,
		Count: len(dagIds),
	}
}

// DAGSummaryPresenter represents a DAG summary with essential information for list endpoints
//
// @Description Summary information for a DAG including ID, title, and validation status
// @Example {"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Employment Law Case", "is_valid": true}
type DAGSummaryPresenter struct {
	Id      uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Unique identifier for the Legal Case DAG"`
	Title   string    `json:"title" example:"Employment Discrimination Case" description:"Human-readable title describing the legal case context"`
	IsValid bool      `json:"is_valid" example:"true" description:"Whether the DAG has passed validation successfully"`
}

// DAGSummaryListPresenter represents a list of DAG summaries for API responses
//
// @Description List of DAG summaries with essential information for efficient overview
// @Example {"dags": [{"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Employment Law", "is_valid": true}], "count": 1}
type DAGSummaryListPresenter struct {
	DAGs  []DAGSummaryPresenter `json:"dags" description:"Array of DAG summaries with essential information"`
	Count int                   `json:"count" description:"Total number of DAGs available"`
}

func NewDAGSummaryPresenter(dag *model.DAG) DAGSummaryPresenter {
	isValid := false
	if dag.Metadata != nil {
		isValid = dag.Metadata.IsValid
	}

	return DAGSummaryPresenter{
		Id:      dag.Id,
		Title:   dag.Title,
		IsValid: isValid,
	}
}

func NewDAGSummaryListPresenter(dags []*model.DAG) DAGSummaryListPresenter {
	summaries := make([]DAGSummaryPresenter, len(dags))
	for i, dag := range dags {
		summaries[i] = NewDAGSummaryPresenter(dag)
	}

	return DAGSummaryListPresenter{
		DAGs:  summaries,
		Count: len(summaries),
	}
}

// DAGMetadataPresenter represents DAG metadata information without content
//
// @Description DAG metadata including ID, title, validation status, and statistics
// @Example {"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Employment Law Case", "is_valid": true, "statistics": {"total_nodes": 5, "root_nodes": 1}}
type DAGMetadataPresenter struct {
	Id         uuid.UUID                     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Unique identifier for the Legal Case DAG"`
	Title      string                        `json:"title" example:"Employment Discrimination Case" description:"Human-readable title describing the legal case context"`
	IsValid    bool                          `json:"is_valid" example:"true" description:"Whether the DAG has passed validation successfully"`
	Statistics ValidationStatisticsPresenter `json:"statistics" description:"DAG validation statistics"`
}

// DAGContentPresenter represents a DAG with only content (no metadata)
//
// @Description DAG content including ID, title, and all nodes with answers
// @Example {"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Employment Law", "nodes": [...]}
type DAGContentPresenter struct {
	Id    uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Unique identifier for the Legal Case DAG"`
	Title string          `json:"title" example:"Employment Discrimination Case" description:"Human-readable title describing the legal case context"`
	Nodes []NodePresenter `json:"nodes" description:"Array of question nodes that make up the legal case decision tree"`
}

func NewDAGMetadataPresenter(dag *model.DAG) DAGMetadataPresenter {
	isValid := false
	stats := ValidationStatisticsPresenter{}

	if dag.Metadata != nil {
		isValid = dag.Metadata.IsValid
		stats = convertValidationStatsToPresenter(dag.Metadata.Statistics)
	}

	return DAGMetadataPresenter{
		Id:         dag.Id,
		Title:      dag.Title,
		IsValid:    isValid,
		Statistics: stats,
	}
}

func NewDAGContentPresenter(dag *model.DAG) DAGContentPresenter {
	nodes := make([]NodePresenter, 0, len(dag.Nodes))
	for _, node := range dag.Nodes {
		nodes = append(nodes, NewNodePresenter(node))
	}

	return DAGContentPresenter{
		Id:    dag.Id,
		Title: dag.Title,
		Nodes: nodes,
	}
}

// Helper function to convert model.ValidationStatistics to ValidationStatisticsPresenter
func convertValidationStatsToPresenter(stats model.ValidationStatistics) ValidationStatisticsPresenter {
	return ValidationStatisticsPresenter{
		TotalNodes:   stats.TotalNodes,
		RootNodes:    stats.RootNodes,
		LeafNodes:    stats.LeafNodes,
		TotalAnswers: stats.TotalAnswers,
		MaxDepth:     stats.MaxDepth,
		HasCycles:    stats.HasCycles,
		RootNodeIDs:  stats.RootNodeIDs,
		LeafNodeIDs:  stats.LeafNodeIDs,
		CyclePaths:   stats.CyclePaths,
	}
}

// presenterToDAG converts a DAGPresenter to a DAG struct
func (h *dagHandler) presenterToDAG(presenter DAGPresenter) *model.DAG {
	nodes := make(map[uuid.UUID]model.Node)

	for _, nodePresenter := range presenter.Nodes {
		answers := make([]model.Answer, len(nodePresenter.Answers))

		for i, answerPresenter := range nodePresenter.Answers {
			answers[i] = model.Answer{
				Id:          answerPresenter.Id,
				Statement:   answerPresenter.Statement,
				NextNode:    answerPresenter.NextNode,
				UserContext: answerPresenter.UserContext,
				Metadata:    answerPresenter.Metadata,
			}
		}

		node := model.Node{
			Id:       nodePresenter.Id,
			Question: nodePresenter.Question,
			Answers:  answers,
		}

		// Set parent pointers for answers
		for i := range node.Answers {
			node.Answers[i].ParentNode = &node
		}

		nodes[node.Id] = node
	}

	return &model.DAG{
		Id:    presenter.Id,
		Title: presenter.Title,
		Nodes: nodes,
	}
}
