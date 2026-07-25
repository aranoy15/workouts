#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
IMAGE_NAME="${IMAGE_NAME:-workouts-backend}"

# Function to get REGISTRY_ID
get_registry_id() {
    if [ -z "$REGISTRY_ID" ]; then
        echo -e "${YELLOW}REGISTRY_ID not specified, trying to get automatically...${NC}"
        REGISTRY_ID=$(yc container registry list --format=json 2>/dev/null | jq -r '.[0].id' 2>/dev/null)
        if [ -z "$REGISTRY_ID" ] || [ "$REGISTRY_ID" = "null" ]; then
            echo -e "${RED}Error: Failed to get registry ID automatically${NC}"
            echo -e "${YELLOW}Please specify registry ID via REGISTRY_ID variable${NC}"
            echo -e "${YELLOW}Example: REGISTRY_ID=your-registry-id ./scripts/deploy.sh${NC}"
            exit 1
        else
            echo -e "${GREEN}Found registry with ID: $REGISTRY_ID${NC}"
        fi
    else
        echo -e "${GREEN}Using specified registry ID: $REGISTRY_ID${NC}"
    fi
}

# Step 1: Docker Build
echo -e "${BLUE}Building Docker image for linux/amd64...${NC}"
if docker buildx version > /dev/null 2>&1; then
    echo -e "${BLUE}Using Docker Buildx...${NC}"
    docker buildx build --platform linux/amd64 -t "$IMAGE_NAME" -f docker/Dockerfile . --load
else
    echo -e "${YELLOW}Buildx not available, using legacy builder with explicit platform...${NC}"
    docker build --platform linux/amd64 -t "$IMAGE_NAME" -f docker/Dockerfile .
fi
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Docker build failed${NC}"
    exit 1
fi
echo -e "${GREEN}Docker image built!${NC}"

# Step 2: Docker Push
echo -e "${BLUE}Pushing image to Yandex Cloud Registry...${NC}"
get_registry_id

echo -e "${BLUE}Tagging image...${NC}"
docker tag "$IMAGE_NAME" "cr.yandex/$REGISTRY_ID/$IMAGE_NAME:latest"

echo -e "${BLUE}Pushing image to registry...${NC}"
docker push "cr.yandex/$REGISTRY_ID/$IMAGE_NAME:latest"
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Docker push failed${NC}"
    exit 1
fi
echo -e "${GREEN}Image successfully pushed to registry!${NC}"

# Step 3: Deploy Revision (reuse deploy-revision.sh script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/deploy-revision.sh"
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Deployment failed${NC}"
    exit 1
fi

echo -e "${GREEN}Full deployment completed!${NC}"
