# DAG Validation API

## Overview

The DAG Validation API endpoint allows you to validate DAG structures without saving them to the system. This is useful for:

- Pre-validation before creating or updating DAGs
- Testing DAG structures during development
- Integration with external tools that need to validate legal case decision trees

## Endpoint

```
POST /v1/dags/validate
```

## Authentication

Requires API key authentication via `X-API-Key` header.

## Request Format

```json
{
  "dag": {
    "id": "uuid",
    "title": "DAG Title",
    "nodes": [
      {
        "id": "uuid",
        "question": "Question text?",
        "answers": [
          {
            "id": "uuid",
            "answer": "Answer text",
            "next_node": "uuid" // Optional, null for leaf answers
          }
        ]
      }
    ]
  }
}
```

## Response Format

### Success Response (200 OK)

The endpoint always returns `200 OK`, even for invalid DAGs. Check the `is_valid` field to determine validation status.

```json
{
  "is_valid": true|false,
  "errors": [
    {
      "code": "ERROR_CODE",
      "message": "Human readable error message",
      "node_id": "uuid", // Optional
      "answer_id": "uuid", // Optional
      "severity": "error"
    }
  ],
  "warnings": [
    {
      "code": "WARNING_CODE", 
      "message": "Human readable warning message",
      "node_id": "uuid", // Optional
      "answer_id": "uuid" // Optional
    }
  ],
  "statistics": {
    "total_nodes": 5,
    "root_nodes": 1,
    "leaf_nodes": 2,
    "total_answers": 12,
    "max_depth": 3,
    "has_cycles": false,
    "root_node_ids": ["uuid"],
    "leaf_node_ids": ["uuid", "uuid"],
    "cycle_paths": [] // Array of cycle paths if cycles detected
  }
}
```

## Validation Rules

The validator checks for:

### ✅ **Single Root Node**
- DAG must have exactly one root node (not referenced by any answer)
- Error codes: `DAG_NO_ROOT`, `DAG_MULTIPLE_ROOTS`

### ✅ **Acyclic Structure** 
- DAG must not contain cycles (no circular references)
- Error code: `DAG_HAS_CYCLES`
- Provides detailed cycle paths for debugging

### ✅ **Structure Integrity**
- Valid UUID formats for all IDs
- Non-empty required fields (title, questions, answer statements)
- Valid node-answer relationships
- All `next_node` references point to existing nodes

### ✅ **Statistical Analysis**
- Calculates DAG depth, node counts, and structure metrics
- Identifies root and leaf nodes
- Reports structural characteristics

## Error Codes

| Code | Description |
|------|-------------|
| `DAG_NULL` | DAG cannot be null |
| `DAG_INVALID_ID` | DAG ID is invalid or empty |
| `DAG_EMPTY_TITLE` | DAG title cannot be empty |
| `DAG_NO_NODES` | DAG must contain at least one node |
| `DAG_NO_ROOT` | No root node found (circular references) |
| `DAG_MULTIPLE_ROOTS` | Multiple root nodes found |
| `DAG_HAS_CYCLES` | DAG contains cycles |
| `NODE_ID_MISMATCH` | Node map key doesn't match node ID |
| `NODE_EMPTY_QUESTION` | Node has empty question |
| `ANSWER_INVALID_ID` | Answer has invalid ID |
| `ANSWER_EMPTY_STATEMENT` | Answer has empty statement |
| `ANSWER_INVALID_REFERENCE` | Answer references non-existent node |

## Examples

### Valid DAG

```bash
curl -X POST http://localhost:8080/v1/dags/validate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "dag": {
      "id": "12345678-1234-1234-1234-123456789abc",
      "title": "Employment Discrimination Case",
      "nodes": [
        {
          "id": "aaaaaaaa-1111-1111-1111-111111111111",
          "question": "Was there workplace discrimination?",
          "answers": [
            {
              "id": "bbbbbbbb-2222-2222-2222-222222222222",
              "answer": "Yes",
              "next_node": "cccccccc-3333-3333-3333-333333333333"
            },
            {
              "id": "dddddddd-4444-4444-4444-444444444444",
              "answer": "No"
            }
          ]
        },
        {
          "id": "cccccccc-3333-3333-3333-333333333333",
          "question": "What type of discrimination?",
          "answers": []
        }
      ]
    }
  }'
```

**Response:**
```json
{
  "is_valid": true,
  "statistics": {
    "total_nodes": 2,
    "root_nodes": 1,
    "leaf_nodes": 1,
    "total_answers": 2,
    "max_depth": 1,
    "has_cycles": false,
    "root_node_ids": ["aaaaaaaa-1111-1111-1111-111111111111"],
    "leaf_node_ids": ["cccccccc-3333-3333-3333-333333333333"]
  }
}
```

### Invalid DAG (Cyclic)

```bash
curl -X POST http://localhost:8080/v1/dags/validate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "dag": {
      "id": "12345678-1234-1234-1234-123456789abc",
      "title": "Cyclic DAG",
      "nodes": [
        {
          "id": "aaaaaaaa-1111-1111-1111-111111111111",
          "question": "Node 1?",
          "answers": [
            {
              "id": "bbbbbbbb-2222-2222-2222-222222222222",
              "answer": "Go to Node 2",
              "next_node": "cccccccc-3333-3333-3333-333333333333"
            }
          ]
        },
        {
          "id": "cccccccc-3333-3333-3333-333333333333", 
          "question": "Node 2?",
          "answers": [
            {
              "id": "dddddddd-4444-4444-4444-444444444444",
              "answer": "Go back to Node 1",
              "next_node": "aaaaaaaa-1111-1111-1111-111111111111"
            }
          ]
        }
      ]
    }
  }'
```

**Response:**
```json
{
  "is_valid": false,
  "errors": [
    {
      "code": "DAG_NO_ROOT",
      "message": "DAG has no root node - this indicates a circular reference",
      "severity": "error"
    },
    {
      "code": "DAG_HAS_CYCLES", 
      "message": "DAG contains 1 cycle(s). A valid DAG must be acyclic",
      "severity": "error"
    }
  ],
  "statistics": {
    "total_nodes": 2,
    "root_nodes": 0,
    "leaf_nodes": 0, 
    "total_answers": 2,
    "max_depth": 0,
    "has_cycles": true,
    "cycle_paths": [
      "[aaaaaaaa-1111-1111-1111-111111111111 cccccccc-3333-3333-3333-333333333333 aaaaaaaa-1111-1111-1111-111111111111]"
    ]
  }
}
```

## Integration Notes

- The validation endpoint is independent of DAG storage
- Results are not cached or persisted
- Use this endpoint for pre-validation before calling update endpoints
- The same validation logic is used during actual DAG updates
- Supports the same authentication mechanisms as other DAG endpoints

## Performance

- Validation is performed in-memory without database access
- Typical response time: < 100ms for DAGs with 50+ nodes
- No rate limiting applied (relies on general API rate limits)
- Suitable for integration into form validation and CI/CD pipelines
