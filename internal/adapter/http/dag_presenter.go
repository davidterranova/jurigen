package http

import (
	"davidterranova/jurigen/internal/dag"

	"github.com/google/uuid"
)

// DAGPresenter represents a complete Legal Case DAG structure for API responses
//
// @Description Legal Case DAG with questions, answers, and context
// @Example {"id": "550e8400-e29b-41d4-a716-446655440000", "nodes": [{"id": "8b007ce4-b676-5fb3-9f93-f5f6c41cb655", "question": "Were you discriminated against?", "answers": [{"id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8", "answer": "Yes, age discrimination occurred", "user_context": "Manager explicitly mentioned my age", "metadata": {"confidence": 0.9, "tags": ["age_discrimination"]}}]}]}
type DAGPresenter struct {
	Id    uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Unique identifier for the Legal Case DAG"`
	Nodes []NodePresenter `json:"nodes" description:"Array of question nodes that make up the legal case decision tree"`
}

func NewDAGPresenter(dag *dag.DAG) DAGPresenter {
	nodes := make([]NodePresenter, 0, len(dag.Nodes))
	for _, node := range dag.Nodes {
		nodes = append(nodes, NewNodePresenter(node))
	}

	return DAGPresenter{
		Id:    dag.Id,
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

func NewNodePresenter(node dag.Node) NodePresenter {
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
	UserContext string                 `json:"user_context,omitempty" example:"Manager explicitly mentioned my age during termination" description:"Free-form user notes and context for this answer"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Structured metadata for legal assessment: confidence scores, evidence tracking, damages estimates, action items, etc."`
}

func NewAnswerPresenter(answer dag.Answer) AnswerPresenter {
	return AnswerPresenter{
		Id:          answer.Id,
		Statement:   answer.Statement,
		UserContext: answer.UserContext,
		Metadata:    answer.Metadata,
	}
}
