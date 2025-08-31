# Jurigen API Documentation

## Overview

Jurigen provides a REST API for managing Legal Case Context through Directed Acyclic Graphs (DAGs). The API enables legal professionals to traverse decision trees, collect case context, and build comprehensive legal assessments.

## Base URL

```
http://localhost:8080/v1
```

## Authentication

The API supports Bearer token authentication via the `Authorization` header:

```bash
Authorization: Bearer <your-token>
```

*Note: Authentication is optional and can be disabled by setting `authFn` to `nil` in the server configuration.*

## OpenAPI/Swagger Documentation

### 📖 Interactive Documentation

Access the Swagger UI at: **http://localhost:8080/swagger/**

### 📄 Raw Specifications

- **JSON**: `docs/swagger/swagger.json`
- **YAML**: `docs/swagger/swagger.yaml`

## API Endpoints

### `GET /v1/dags/{dagId}`

Retrieve a complete Legal Case DAG structure.

**Parameters:**
- `dagId` (path, required): UUID of the DAG to retrieve

**Response 200 - Success:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "nodes": [
    {
      "id": "8b007ce4-b676-5fb3-9f93-f5f6c41cb655",
      "question": "Were you discriminated against in the workplace?",
      "answers": [
        {
          "id": "fc28c4b6-d185-cf56-a7e4-dead499ff1e8",
          "answer": "Yes, age discrimination occurred",
          "user_context": "Manager explicitly mentioned my age during termination",
          "metadata": {
            "confidence": 0.9,
            "severity": "high",
            "tags": ["age_discrimination", "wrongful_termination"],
            "sources": ["HR_Email.pdf", "Witness_Statement.pdf"],
            "damages_estimate": 75000,
            "evidence_quality": "strong",
            "actions_needed": ["gather_employment_records", "file_eeoc_complaint"]
          }
        }
      ]
    }
  ]
}
```

**Error Responses:**
- `400`: Invalid DAG ID format
- `404`: DAG not found  
- `500`: Internal server error

## Data Models

### DAGPresenter

Represents a complete Legal Case DAG structure.

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier for the Legal Case DAG |
| `nodes` | Array<NodePresenter> | Question nodes that make up the decision tree |

### NodePresenter

Represents a question node in the Legal Case DAG.

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier for the question node |
| `question` | string | The legal question being asked |
| `answers` | Array<AnswerPresenter> | Available answer options |

### AnswerPresenter

Represents an answer option with optional legal context.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique identifier for the answer |
| `answer` | string | Yes | The answer statement or response |
| `user_context` | string | No | Free-form user notes and context |
| `metadata` | object | No | Structured metadata for legal assessment |

### Metadata Structure

The `metadata` field supports flexible legal case data:

#### **Assessment & Confidence**
```json
{
  "confidence": 0.9,           // User confidence (0.0-1.0)
  "legal_strength": 0.8,       // Attorney assessment (0.0-1.0)
  "severity": "high"           // low/medium/high/critical
}
```

#### **Evidence & Documentation**
```json
{
  "sources": ["HR_Email.pdf", "Witness_Statement.pdf"],
  "evidence_quality": "strong", // weak/medium/strong/compelling
  "witnesses": ["Jane Doe", "John Smith"],
  "document_ids": ["DOC001", "DOC002"]
}
```

#### **Case Management**
```json
{
  "tags": ["age_discrimination", "wrongful_termination"],
  "case_type": "employment_discrimination",
  "actions_needed": ["gather_records", "file_complaint"],
  "priority": 8,               // 1-10 urgency scale
  "deadline": "2024-03-01"
}
```

#### **Financial & Damages**
```json
{
  "damages_estimate": 75000,
  "medical_expenses": 15000,
  "lost_wages": 8000,
  "settlement_minimum": 40000
}
```

#### **Timeline & Chronology**
```json
{
  "date_occurred": "2024-01-15T10:30:00Z",
  "timeline_order": 3,         // Sequential order in case timeline
  "frequency": "daily",        // how often incident occurred
  "duration": "6_months"       // how long situation lasted
}
```

## Usage Examples

### Basic DAG Retrieval

```bash
curl -X GET "http://localhost:8080/v1/dags/550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer your-token"
```

### Using with jq for Processing

```bash
# Get all questions in a DAG
curl -s "http://localhost:8080/v1/dags/550e8400-e29b-41d4-a716-446655440000" | \
  jq '.nodes[].question'

# Extract high-confidence answers
curl -s "http://localhost:8080/v1/dags/550e8400-e29b-41d4-a716-446655440000" | \
  jq '.nodes[].answers[] | select(.metadata.confidence > 0.8)'

# Get all evidence sources
curl -s "http://localhost:8080/v1/dags/550e8400-e29b-41d4-a716-446655440000" | \
  jq '.nodes[].answers[].metadata.sources[]?' | sort | uniq
```

## Development Workflow

### Generate Documentation

```bash
# Generate OpenAPI docs from code annotations
make swagger

# Or use the script directly
./scripts/generate-swagger.sh
```

### Serve Documentation Locally

```bash
# Option 1: Through the API server
go run main.go server
# Then visit: http://localhost:8080/swagger/

# Option 2: Standalone documentation server
make swagger-serve
# Then visit: http://localhost:8081/swagger/
```

### Update Documentation

1. **Add/modify Swagger annotations** in your Go code
2. **Regenerate docs**: `make swagger`
3. **Restart server** to pick up changes

## Integration Examples

### Frontend Integration

```javascript
// Fetch DAG structure
const response = await fetch('/v1/dags/your-dag-id', {
  headers: {
    'Authorization': 'Bearer your-token'
  }
});
const dag = await response.json();

// Process questions and answers
dag.nodes.forEach(node => {
  console.log(`Question: ${node.question}`);
  node.answers.forEach(answer => {
    console.log(`  Answer: ${answer.answer}`);
    if (answer.user_context) {
      console.log(`  Context: ${answer.user_context}`);
    }
    if (answer.metadata?.confidence) {
      console.log(`  Confidence: ${answer.metadata.confidence}`);
    }
  });
});
```

### Python Integration

```python
import requests
import json

# Fetch DAG
response = requests.get(
    'http://localhost:8080/v1/dags/your-dag-id',
    headers={'Authorization': 'Bearer your-token'}
)
dag = response.json()

# Extract high-priority items
high_priority = [
    answer for node in dag['nodes'] 
    for answer in node['answers']
    if answer.get('metadata', {}).get('priority', 0) >= 8
]

print(f"Found {len(high_priority)} high-priority items")
```

## Error Handling

All errors follow a consistent format:

```json
{
  "message": "Human-readable error description",
  "error": "Technical error details"
}
```

Common HTTP status codes:
- `200`: Success
- `400`: Bad Request (invalid input)
- `401`: Unauthorized (invalid/missing token)
- `404`: Not Found (DAG doesn't exist)
- `500`: Internal Server Error

## Development & Contribution

To add new endpoints:

1. **Add route** in `pkg/adapter/http/router.go`
2. **Add handler** with Swagger annotations
3. **Add presenter models** with documentation
4. **Regenerate docs**: `make swagger`
5. **Test endpoints** via Swagger UI

For more details, see the project's main README and CONTRIBUTING guidelines.
