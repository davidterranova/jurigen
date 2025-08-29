package dag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type DAG struct {
	Nodes map[uuid.UUID]Node
}

func NewDAG() DAG {
	return DAG{
		Nodes: make(map[uuid.UUID]Node),
	}
}

func (d DAG) GetNode(id uuid.UUID) (Node, error) {
	node, ok := d.Nodes[id]
	if !ok {
		return Node{}, fmt.Errorf("node not found")
	}
	return node, nil
}

// GetRootNode finds and returns the root node of the DAG
// The root node is one that is not referenced as a next_node by any answer
func (d DAG) GetRootNode() (Node, error) {
	// Collect all node IDs that are referenced as next_node
	referencedNodes := make(map[uuid.UUID]bool)

	for _, node := range d.Nodes {
		for _, answer := range node.Answers {
			if answer.NextNode != nil {
				referencedNodes[*answer.NextNode] = true
			}
		}
	}

	// Find nodes that are not referenced (potential root nodes)
	var rootNodes []Node
	for _, node := range d.Nodes {
		if !referencedNodes[node.Id] {
			rootNodes = append(rootNodes, node)
		}
	}

	if len(rootNodes) == 0 {
		return Node{}, fmt.Errorf("no root node found")
	}

	if len(rootNodes) > 1 {
		return Node{}, fmt.Errorf("multiple root nodes found, DAG should have exactly one root")
	}

	return rootNodes[0], nil
}

type Node struct {
	Id       uuid.UUID `json:"id"`
	Question string    `json:"question"`
	Answers  []Answer  `json:"answers"`
}

type Answer struct {
	Id        uuid.UUID  `json:"id"`
	Statement string     `json:"answer"`
	NextNode  *uuid.UUID `json:"next_node"`
}

func (d DAG) MarshalJSON() ([]byte, error) {
	nodes := make([]Node, 0, len(d.Nodes))
	for _, node := range d.Nodes {
		nodes = append(nodes, node)
	}

	return json.Marshal(nodes)
}

func (d *DAG) UnmarshalJSON(data []byte) error {
	nodes := make([]Node, 0)

	err := json.Unmarshal(data, &nodes)
	if err != nil {
		return fmt.Errorf("error unmarshalling data: %w", err)
	}

	// Initialize the Nodes map if it's nil
	if d.Nodes == nil {
		d.Nodes = make(map[uuid.UUID]Node)
	}

	// Add all nodes to the map
	for _, node := range nodes {
		d.Nodes[node.Id] = node
	}

	return nil
}

func (d DAG) String() string {
	var sb strings.Builder
	for _, node := range d.Nodes {
		sb.WriteString("Question: " + node.Question + "\n")
		for _, answer := range node.Answers {
			sb.WriteString("\tAnswer: " + answer.Statement)
			if answer.NextNode != nil {
				nextNode, err := d.GetNode(*answer.NextNode)
				if err != nil {
					sb.WriteString(" -> [ERROR: " + err.Error() + "]")
				} else {
					sb.WriteString(" -> " + nextNode.Question)
				}
			} else {
				sb.WriteString(" -> [LEAF]")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Walk traverses the DAG starting from the given node ID, using fnAnswer to determine
// which answer to follow at each step until reaching a leaf node.
func (d DAG) Walk(nodeId uuid.UUID, fnAnswer func(Node) (Answer, error)) ([]Answer, error) {
	var path []Answer
	currentNodeId := nodeId

	for {
		// Get the current node
		currentNode, err := d.GetNode(currentNodeId)
		if err != nil {
			return path, fmt.Errorf("error getting node %s: %w", currentNodeId, err)
		}

		// If this is a leaf node (no answers), we're done
		if len(currentNode.Answers) == 0 {
			break
		}

		// Get the answer choice from the provided function
		selectedAnswer, err := fnAnswer(currentNode)
		if err != nil {
			return path, fmt.Errorf("error getting answer for node %s: %w", currentNodeId, err)
		}

		// Validate that the selected answer belongs to this node
		var validAnswer *Answer
		for _, answer := range currentNode.Answers {
			if answer.Id == selectedAnswer.Id {
				validAnswer = &answer
				break
			}
		}

		if validAnswer == nil {
			return path, fmt.Errorf("selected answer %s is not valid for node %s", selectedAnswer.Id, currentNodeId)
		}

		// Add the answer to the path
		path = append(path, *validAnswer)

		// If this answer has no next node, we've reached a leaf
		if validAnswer.NextNode == nil {
			break
		}

		// Move to the next node
		currentNodeId = *validAnswer.NextNode
	}

	return path, nil
}

func CLIFnAnswer(node Node) (Answer, error) {
	fmt.Printf("\n%s\n", node.Question)
	fmt.Println(strings.Repeat("-", len(node.Question)))

	// Display numbered options
	for i, answer := range node.Answers {
		fmt.Printf("%d. %s\n", i+1, answer.Statement)
	}

	// Prompt for user input
	fmt.Print("\nSelect your answer (enter the number): ")

	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return Answer{}, fmt.Errorf("invalid input: %w", err)
	}

	// Validate choice
	if choice < 1 || choice > len(node.Answers) {
		return Answer{}, fmt.Errorf("invalid choice: must be between 1 and %d", len(node.Answers))
	}

	// Return the selected answer
	selectedAnswer := node.Answers[choice-1]
	fmt.Printf("You selected: %s\n", selectedAnswer.Statement)

	return selectedAnswer, nil
}
