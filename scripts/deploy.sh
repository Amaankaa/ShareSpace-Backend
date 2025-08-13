#!/bin/bash

# ShareSpace Backend Deployment Script
# Usage: ./scripts/deploy.sh [staging|production]

set -e

ENVIRONMENT=${1:-staging}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if environment is valid
if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
    log_error "Invalid environment. Use 'staging' or 'production'"
    exit 1
fi

log_info "Starting deployment to $ENVIRONMENT environment..."

# Check if required files exist
check_requirements() {
    log_info "Checking deployment requirements..."
    
    if [[ ! -f "$PROJECT_ROOT/.env.$ENVIRONMENT" ]]; then
        log_error "Environment file .env.$ENVIRONMENT not found"
        exit 1
    fi
    
    if [[ ! -f "$PROJECT_ROOT/docker-compose.$ENVIRONMENT.yml" ]]; then
        log_error "Docker compose file for $ENVIRONMENT not found"
        exit 1
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    log_info "All requirements satisfied"
}

# Load environment variables
load_environment() {
    log_info "Loading environment variables for $ENVIRONMENT..."
    
    if [[ -f "$PROJECT_ROOT/.env.$ENVIRONMENT" ]]; then
        export $(cat "$PROJECT_ROOT/.env.$ENVIRONMENT" | grep -v '^#' | xargs)
        log_info "Environment variables loaded"
    else
        log_error "Environment file not found: .env.$ENVIRONMENT"
        exit 1
    fi
}

# Backup database (production only)
backup_database() {
    if [[ "$ENVIRONMENT" == "production" ]]; then
        log_info "Creating database backup..."
        
        BACKUP_DIR="$PROJECT_ROOT/backups"
        mkdir -p "$BACKUP_DIR"
        
        BACKUP_NAME="sharespace_backup_$(date +%Y%m%d_%H%M%S)"
        
        if docker ps | grep -q "sharespace_mongodb"; then
            docker exec sharespace_mongodb mongodump --out "/backups/$BACKUP_NAME"
            log_info "Database backup created: $BACKUP_NAME"
        else
            log_warn "MongoDB container not running, skipping backup"
        fi
    fi
}

# Build and deploy
deploy() {
    log_info "Building and deploying application..."
    
    cd "$PROJECT_ROOT"
    
    # Pull latest images
    log_info "Pulling latest Docker images..."
    docker-compose -f "docker-compose.$ENVIRONMENT.yml" pull
    
    # Stop existing containers
    log_info "Stopping existing containers..."
    docker-compose -f "docker-compose.$ENVIRONMENT.yml" down
    
    # Start new containers
    log_info "Starting new containers..."
    docker-compose -f "docker-compose.$ENVIRONMENT.yml" up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 30
    
    # Clean up unused images
    log_info "Cleaning up unused Docker images..."
    docker system prune -f
}

# Health check
health_check() {
    log_info "Running health check..."
    
    local max_attempts=10
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f "http://localhost:8080/health" &> /dev/null; then
            log_info "Health check passed"
            return 0
        fi
        
        log_warn "Health check failed (attempt $attempt/$max_attempts)"
        sleep 10
        ((attempt++))
    done
    
    log_error "Health check failed after $max_attempts attempts"
    return 1
}

# Rollback function
rollback() {
    log_error "Deployment failed. Rolling back..."
    
    # Stop current containers
    docker-compose -f "docker-compose.$ENVIRONMENT.yml" down
    
    # Start previous version (if backup exists)
    # This would require implementing version tagging
    log_info "Manual rollback required"
    exit 1
}

# Main deployment process
main() {
    log_info "ShareSpace Backend Deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Timestamp: $(date)"
    
    # Run deployment steps
    check_requirements
    load_environment
    backup_database
    
    # Deploy with error handling
    if deploy; then
        if health_check; then
            log_info "Deployment completed successfully!"
            
            # Send notification (if configured)
            if [[ -n "$SLACK_WEBHOOK" ]]; then
                curl -X POST -H 'Content-type: application/json' \
                    --data "{\"text\":\"✅ ShareSpace Backend deployed successfully to $ENVIRONMENT\"}" \
                    "$SLACK_WEBHOOK"
            fi
        else
            rollback
        fi
    else
        rollback
    fi
}

# Trap errors and run rollback
trap rollback ERR

# Run main function
main "$@"
