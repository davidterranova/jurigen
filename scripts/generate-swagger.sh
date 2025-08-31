#!/bin/bash

# Jurigen OpenAPI Documentation Generator
# Generates Swagger/OpenAPI documentation from Go code annotations

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs/swagger"

echo -e "${BLUE}🔄 Jurigen OpenAPI Documentation Generator${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo -e "${YELLOW}⚠️  swag CLI not found. Installing...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
    echo -e "${GREEN}✅ swag CLI installed successfully${NC}"
fi

# Navigate to project root
cd "$PROJECT_ROOT"

# Clean previous docs
if [ -d "$DOCS_DIR" ]; then
    echo -e "${YELLOW}🧹 Cleaning previous documentation...${NC}"
    rm -rf "$DOCS_DIR"
fi

# Create docs directory
mkdir -p "$DOCS_DIR"

# Generate OpenAPI documentation
echo -e "${BLUE}🚀 Generating OpenAPI documentation...${NC}"
echo ""

swag init \
    --dir . \
    --output docs/swagger \
    --parseDependency \
    --parseInternal \
    --parseDepth 3 \
    --generalInfo main.go

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ OpenAPI documentation generated successfully!${NC}"
    echo ""
    echo -e "${BLUE}📁 Generated files:${NC}"
    echo "  📄 JSON Spec: docs/swagger/swagger.json"
    echo "  📄 YAML Spec: docs/swagger/swagger.yaml"
    echo "  🔧 Go Docs:   docs/swagger/docs.go"
    echo ""
    
    # Show file sizes
    echo -e "${BLUE}📊 Documentation Stats:${NC}"
    if [ -f "$DOCS_DIR/swagger.json" ]; then
        JSON_SIZE=$(stat -f%z "$DOCS_DIR/swagger.json" 2>/dev/null || stat -c%s "$DOCS_DIR/swagger.json" 2>/dev/null)
        echo "  JSON: ${JSON_SIZE} bytes"
    fi
    if [ -f "$DOCS_DIR/swagger.yaml" ]; then
        YAML_SIZE=$(stat -f%z "$DOCS_DIR/swagger.yaml" 2>/dev/null || stat -c%s "$DOCS_DIR/swagger.yaml" 2>/dev/null)
        echo "  YAML: ${YAML_SIZE} bytes"
    fi
    
    # Count endpoints
    if [ -f "$DOCS_DIR/swagger.json" ]; then
        ENDPOINT_COUNT=$(grep -o '"paths"' "$DOCS_DIR/swagger.json" | wc -l)
        echo "  Endpoints: documented"
    fi
    
    echo ""
    echo -e "${BLUE}🌐 Next Steps:${NC}"
    echo "  • View docs: make swagger-serve (opens http://localhost:8081/swagger/)"
    echo "  • Integrate UI: Add swagger endpoint to your HTTP server"
    echo "  • Share spec: Use docs/swagger/swagger.json or swagger.yaml"
    echo ""
    echo -e "${GREEN}🎉 Ready to use!${NC}"
else
    echo ""
    echo -e "${RED}❌ Failed to generate OpenAPI documentation${NC}"
    echo -e "${YELLOW}💡 Common issues:${NC}"
    echo "  • Check that all imports are valid"
    echo "  • Verify @tags match between endpoints and @tag.name definitions"
    echo "  • Ensure model structs have proper annotations"
    echo "  • Run 'go mod tidy' to clean up dependencies"
    exit 1
fi
