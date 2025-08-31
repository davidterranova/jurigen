# DAG

## Metadata usage examples for legal cases

### 🔍 Evidence & Documentation
```
metadata := map[string]interface{}{
    "sources":          []string{"Email_HR_2024-01-15.pdf", "Witness_John_Doe"},
    "evidence_quality": "strong",
    "document_ids":     []string{"DOC001", "DOC002"},
    "attachments":      []string{"contract.pdf", "email_thread.txt"},
}
```

### ⚖️ Legal Assessment
```
metadata := map[string]interface{}{
    "confidence":          0.8,        // 0.0 to 1.0 scale
    "legal_strength":      0.7,        // Case strength estimate  
    "damages_estimate":    50000,      // Monetary estimate
    "statute_limitations": "2025-01-15", // Critical deadline
    "jurisdiction":        "california",  // Legal jurisdiction
}
```

### 🏷️ Categorization & Organization
```
metadata := map[string]interface{}{
    "tags":      []string{"discrimination", "wrongful_termination", "urgent"},
    "severity":  "high",        // low/medium/high/critical
    "case_type": "employment",  // employment/civil/criminal
    "priority":  8,            // 1-10 scale
    "category":  "workplace_harassment",
}
```

### 📅 Timeline & Action Items
```
metadata := map[string]interface{}{
    "date_occurred":   "2024-01-15T10:30:00Z",
    "date_discovered": "2024-02-01T14:00:00Z", 
    "actions_needed":  []string{"gather_records", "interview_witnesses"},
    "deadline":        "2024-03-01",
    "assigned_to":     "attorney_smith",
    "status":          "in_progress",
}
```

### 📋 Action Items & Case Management
```
metadata := map[string]interface{}{
    "actions_needed": []string{
        "gather_employment_records",
        "interview_witnesses",
        "file_eeoc_complaint",
        "medical_expert_consultation",
    },
    "deadline":        "2024-03-01",
    "assigned_to":     "attorney_smith",
    "status":          "in_progress",
    "next_hearing":    "2024-04-15T09:00:00Z",
}
```

### 💰 Financial & Damages Information
```
metadata := map[string]interface{}{
    "damages_estimate":    50000,
    "medical_expenses":    15000,
    "lost_wages":         8000,
    "future_losses":      25000,
    "punitive_damages":   true,
    "settlement_minimum": 40000,
    "insurance_coverage": 100000,
}
```