#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
IMAGE_NAME="${IMAGE_NAME:-workouts-backend}"
CONTAINER_NAME="${CONTAINER_NAME:-workouts-backend}"
CONTAINER_CONFIG_FILE="${CONTAINER_CONFIG_FILE:-.container-config.json}"

# Function to get REGISTRY_ID
get_registry_id() {
    if [ -z "$REGISTRY_ID" ]; then
        echo -e "${YELLOW}REGISTRY_ID not specified, trying to get automatically...${NC}"
        REGISTRY_ID=$(yc container registry list --format=json 2>/dev/null | jq -r '.[0].id' 2>/dev/null)
        if [ -z "$REGISTRY_ID" ] || [ "$REGISTRY_ID" = "null" ]; then
            echo -e "${RED}Error: Failed to get registry ID automatically${NC}"
            echo -e "${YELLOW}Please specify registry ID via REGISTRY_ID variable${NC}"
            echo -e "${YELLOW}Example: REGISTRY_ID=your-registry-id ./scripts/deploy-revision.sh${NC}"
            exit 1
        else
            echo -e "${GREEN}Found registry with ID: $REGISTRY_ID${NC}"
        fi
    else
        echo -e "${GREEN}Using specified registry ID: $REGISTRY_ID${NC}"
    fi
}

# Script start
echo -e "${BLUE}Creating new serverless container revision...${NC}"

if [ -z "$CONTAINER_NAME" ]; then
    CONTAINER_NAME="workouts-backend"
    echo -e "${YELLOW}CONTAINER_NAME not specified, using default: $CONTAINER_NAME${NC}"
else
    echo -e "${GREEN}Using specified container name: $CONTAINER_NAME${NC}"
fi

get_registry_id

IMAGE_URL="cr.yandex/$REGISTRY_ID/$IMAGE_NAME:latest"
echo -e "${BLUE}Deploying revision with image: $IMAGE_URL${NC}"

# Get SERVICE_ACCOUNT_ID
if [ -n "$SERVICE_ACCOUNT_ID" ]; then
    echo -e "${GREEN}Using specified service account ID: $SERVICE_ACCOUNT_ID${NC}"
elif [ -n "$SERVICE_ACCOUNT_NAME" ]; then
    echo -e "${BLUE}Getting service account ID for: $SERVICE_ACCOUNT_NAME${NC}"
    SERVICE_ACCOUNT_ID=$(yc iam service-account get "$SERVICE_ACCOUNT_NAME" --format json 2>/dev/null | jq -r '.id' 2>/dev/null)
    if [ -z "$SERVICE_ACCOUNT_ID" ] || [ "$SERVICE_ACCOUNT_ID" = "null" ]; then
        echo -e "${RED}Error: Failed to get service account ID for '$SERVICE_ACCOUNT_NAME'${NC}"
        echo -e "${YELLOW}Please check that service account exists: yc iam service-account list${NC}"
        exit 1
    fi
    echo -e "${GREEN}Found service account ID: $SERVICE_ACCOUNT_ID${NC}"
else
    SERVICE_ACCOUNT_NAME="workouts-backend-admin"
    echo -e "${YELLOW}SERVICE_ACCOUNT_NAME not specified, using default: $SERVICE_ACCOUNT_NAME${NC}"
    echo -e "${BLUE}Getting service account ID for: $SERVICE_ACCOUNT_NAME${NC}"
    SERVICE_ACCOUNT_ID=$(yc iam service-account get "$SERVICE_ACCOUNT_NAME" --format json 2>/dev/null | jq -r '.id' 2>/dev/null)
    if [ -z "$SERVICE_ACCOUNT_ID" ] || [ "$SERVICE_ACCOUNT_ID" = "null" ]; then
        echo -e "${RED}Error: Failed to get service account ID for '$SERVICE_ACCOUNT_NAME'${NC}"
        echo -e "${YELLOW}Please check that service account exists: yc iam service-account list${NC}"
        echo -e "${YELLOW}Or specify SERVICE_ACCOUNT_NAME or SERVICE_ACCOUNT_ID explicitly${NC}"
        exit 1
    fi
    echo -e "${GREEN}Found service account ID: $SERVICE_ACCOUNT_ID${NC}"
fi

# Get revision configuration
echo -e "${BLUE}Getting latest revision configuration...${NC}"
if [ -f "$CONTAINER_CONFIG_FILE" ]; then
    echo -e "${BLUE}Found saved configuration file, using it...${NC}"
    CURRENT_REVISION_JSON=$(cat "$CONTAINER_CONFIG_FILE" 2>/dev/null)
else
    LATEST_REVISION_ID=$(yc serverless container revision list --container-name "$CONTAINER_NAME" --format json 2>/dev/null | jq -r '.[0].id // empty' 2>/dev/null)
    if [ -n "$LATEST_REVISION_ID" ] && [ "$LATEST_REVISION_ID" != "null" ]; then
        echo -e "${BLUE}Getting configuration from revision $LATEST_REVISION_ID...${NC}"
        CURRENT_REVISION_JSON=$(yc serverless container revision get --id "$LATEST_REVISION_ID" --format json 2>/dev/null)
    fi
fi

# Build deploy command
DEPLOY_CMD="yc serverless container revision deploy --container-name $CONTAINER_NAME --image $IMAGE_URL --service-account-id $SERVICE_ACCOUNT_ID"

if [ -n "$CURRENT_REVISION_JSON" ] && [ "$CURRENT_REVISION_JSON" != "null" ]; then
    echo -e "${BLUE}Copying configuration from revision...${NC}"

    CURRENT_CORES=$(echo "$CURRENT_REVISION_JSON" | jq -r '.resources.cores // empty' 2>/dev/null)
    CURRENT_MEMORY_BYTES=$(echo "$CURRENT_REVISION_JSON" | jq -r '.resources.memory // empty' 2>/dev/null)
    CURRENT_TIMEOUT=$(echo "$CURRENT_REVISION_JSON" | jq -r '.execution_timeout // empty' 2>/dev/null)
    CURRENT_CONCURRENCY=$(echo "$CURRENT_REVISION_JSON" | jq -r '.concurrency // empty' 2>/dev/null)

    if [ -n "$CURRENT_CORES" ] && [ "$CURRENT_CORES" != "null" ] && [ "$CURRENT_CORES" != "0" ]; then
        DEPLOY_CMD="$DEPLOY_CMD --cores $CURRENT_CORES"
    fi

    if [ -n "$CURRENT_MEMORY_BYTES" ] && [ "$CURRENT_MEMORY_BYTES" != "null" ]; then
        CURRENT_MEMORY=$(echo "$CURRENT_MEMORY_BYTES" | awk '{if ($1 >= 1073741824) printf "%.0fGB", $1/1073741824; else printf "%.0fMB", $1/1048576}')
        DEPLOY_CMD="$DEPLOY_CMD --memory $CURRENT_MEMORY"
    fi

    if [ -n "$CURRENT_TIMEOUT" ] && [ "$CURRENT_TIMEOUT" != "null" ]; then
        DEPLOY_CMD="$DEPLOY_CMD --execution-timeout $CURRENT_TIMEOUT"
    fi

    if [ -n "$CURRENT_CONCURRENCY" ] && [ "$CURRENT_CONCURRENCY" != "null" ] && [ "$CURRENT_CONCURRENCY" != "0" ]; then
        DEPLOY_CMD="$DEPLOY_CMD --concurrency $CURRENT_CONCURRENCY"
    fi

    ENV_JSON=$(echo "$CURRENT_REVISION_JSON" | jq -c '.image.environment // .environment // {}' 2>/dev/null)
    if [ -n "$ENV_JSON" ] && [ "$ENV_JSON" != "null" ] && [ "$ENV_JSON" != "{}" ]; then
        ENV_COUNT=$(echo "$ENV_JSON" | jq -r 'length // 0' 2>/dev/null)
        if [ "$ENV_COUNT" -gt 0 ] 2>/dev/null; then
            echo -e "${BLUE}Copying $ENV_COUNT environment variable(s)...${NC}"
            ENV_ARGS=$(echo "$ENV_JSON" | jq -r 'to_entries[] | "--environment \(.key)=\(.value)"' 2>/dev/null | tr '\n' ' ')
            if [ -n "$ENV_ARGS" ]; then
                DEPLOY_CMD="$DEPLOY_CMD $ENV_ARGS"
            fi
        fi
    else
        echo -e "${YELLOW}Warning: Environment variables not found in revision JSON.${NC}"
    fi

    SECRETS_JSON=$(echo "$CURRENT_REVISION_JSON" | jq -c '.secrets // []' 2>/dev/null)
    if [ -n "$SECRETS_JSON" ] && [ "$SECRETS_JSON" != "null" ] && [ "$SECRETS_JSON" != "[]" ]; then
        SECRET_COUNT=$(echo "$SECRETS_JSON" | jq -r 'length // 0' 2>/dev/null)
        if [ "$SECRET_COUNT" -gt 0 ] 2>/dev/null; then
            echo -e "${BLUE}Copying $SECRET_COUNT secret(s) from Lockbox...${NC}"
            SECRET_ARGS=$(echo "$SECRETS_JSON" | jq -r '.[] | "--secret id=\(.id),key=\(.key),environment-variable=\(.environment_variable)"' 2>/dev/null | tr '\n' ' ')
            if [ -n "$SECRET_ARGS" ]; then
                DEPLOY_CMD="$DEPLOY_CMD $SECRET_ARGS"
            fi
        fi
    else
        echo -e "${YELLOW}Warning: Secrets not found in revision JSON. They may be stored at container level.${NC}"
    fi
else
    echo -e "${YELLOW}Could not get revision details or revision JSON is empty, using defaults...${NC}"
    DEPLOY_CMD="$DEPLOY_CMD --cores 1 --memory 512MB --execution-timeout 30s --concurrency 10"
fi

# Execute deployment
echo -e "${BLUE}Deploying new revision...${NC}"
eval "$DEPLOY_CMD"
if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to deploy revision${NC}"
    echo -e "${YELLOW}Make sure container '$CONTAINER_NAME' exists${NC}"
    echo -e "${YELLOW}Check: yc serverless container list${NC}"
    exit 1
fi

echo -e "${GREEN}Revision successfully deployed!${NC}"
