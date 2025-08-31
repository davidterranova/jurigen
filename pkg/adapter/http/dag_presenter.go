package http

import (
	"davidterranova/jurigen/internal/dag"

	"github.com/google/uuid"
)

type DAGPresenter struct {
	Id    uuid.UUID       `json:"id"`
	Nodes []NodePresenter `json:"nodes"`
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

type NodePresenter struct {
	Id       uuid.UUID         `json:"id"`
	Question string            `json:"question"`
	Answers  []AnswerPresenter `json:"answers"`
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

type AnswerPresenter struct {
	Id          uuid.UUID              `json:"id"`
	Statement   string                 `json:"answer"`
	UserContext string                 `json:"user_context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func NewAnswerPresenter(answer dag.Answer) AnswerPresenter {
	return AnswerPresenter{
		Id:          answer.Id,
		Statement:   answer.Statement,
		UserContext: answer.UserContext,
		Metadata:    answer.Metadata,
	}
}
