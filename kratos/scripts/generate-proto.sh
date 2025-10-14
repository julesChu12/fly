#!/bin/bash

# Script to generate Go code from protobuf definitions

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating gRPC code from protobuf definitions...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "protoc is not installed. Please install it first."
    echo "Installation instructions: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if required plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "protoc-gen-go is not installed. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "protoc-gen-go-grpc is not installed. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate code
PROTO_DIR="api/proto"
OUT_DIR="api/proto"

echo -e "${YELLOW}Generating code for order service...${NC}"
protoc \
    --go_out=${OUT_DIR} \
    --go_opt=paths=source_relative \
    --go-grpc_out=${OUT_DIR} \
    --go-grpc_opt=paths=source_relative \
    ${PROTO_DIR}/order/v1/order.proto

echo -e "${GREEN}✓ gRPC code generation completed successfully!${NC}"
echo -e "${GREEN}Generated files:${NC}"
echo "  - ${OUT_DIR}/order/v1/order.pb.go"
echo "  - ${OUT_DIR}/order/v1/order_grpc.pb.go"
