#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
CONTAINER_NAME="${CONTAINER_NAME:-workouts-backend}"
CONTAINER_CONFIG_FILE="${CONTAINER_CONFIG_FILE:-.container-config.json}"

# Script start
echo -e "${BLUE}Saving container configuration...${NC}"

LATEST_REVISION_ID=$(yc serverless container revision list --container-name "$CONTAINER_NAME" --format json 2>/dev/null | jq -r '.[0].id // empty' 2>/dev/null)

if [ -n "$LATEST_REVISION_ID" ] && [ "$LATEST_REVISION_ID" != "null" ]; then
    echo -e "${BLUE}Getting revision $LATEST_REVISION_ID configuration...${NC}"
    yc serverless container revision get --id "$LATEST_REVISION_ID" --format json > "$CONTAINER_CONFIG_FILE" 2>/dev/null

    if [ -f "$CONTAINER_CONFIG_FILE" ]; then
        FILE_SIZE=$(wc -c < "$CONTAINER_CONFIG_FILE" 2>/dev/null | tr -d ' ')
        if [ "$FILE_SIZE" -gt 10 ] 2>/dev/null; then
            if jq '.' "$CONTAINER_CONFIG_FILE" > /dev/null 2>&1; then
                SECRET_COUNT=$(jq -r '.secrets | length // 0' "$CONTAINER_CONFIG_FILE" 2>/dev/null)
                echo -e "${GREEN}Configuration saved to $CONTAINER_CONFIG_FILE${NC}"
                echo -e "${GREEN}Found $SECRET_COUNT secret(s) in configuration${NC}"
                echo -e "${YELLOW}Note: Environment variables are not returned by API for security reasons.${NC}"
                echo -e "${YELLOW}If you need to add environment variables, edit $CONTAINER_CONFIG_FILE manually.${NC}"
            else
                echo -e "${RED}Failed to save configuration: Invalid JSON${NC}"
                rm -f "$CONTAINER_CONFIG_FILE"
                exit 1
            fi
        else
            echo -e "${RED}Failed to save configuration: File is empty${NC}"
            rm -f "$CONTAINER_CONFIG_FILE"
            exit 1
        fi
    else
        echo -e "${RED}Failed to save configuration${NC}"
        exit 1
    fi
else
    echo -e "${RED}No revisions found${NC}"
    exit 1
fi
