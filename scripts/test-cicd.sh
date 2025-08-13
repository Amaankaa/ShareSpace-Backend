#!/bin/bash

# CI/CD Implementation Verification Script
# This script tests all components of the CI/CD pipeline locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}✅ [PASS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠️  [WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}❌ [FAIL]${NC} $1"
}

log_test() {
    echo -e "${BLUE}🧪 [TEST]${NC} $1"
}

echo "🚀 ShareSpace CI/CD Implementation Verification"
echo "=============================================="

# Test 1: Check if required files exist
log_test "Checking CI/CD configuration files..."

check_file() {
    if [[ -f "$1" ]]; then
        log_info "Found: $1"
        return 0
    else
        log_error "Missing: $1"
        return 1
    fi
}

check_dir() {
    if [[ -d "$1" ]]; then
        log_info "Found directory: $1"
        return 0
    else
        log_error "Missing directory: $1"
        return 1
    fi
}

# Check GitHub Actions
check_file ".github/workflows/ci-cd.yml"

# Check Docker files
check_file "Dockerfile"
check_file "docker-compose.yml"
check_file "docker-compose.prod.yml"
check_file ".dockerignore"

# Check deployment scripts
check_file "scripts/deploy.sh"
check_file "scripts/mongo-init.js"

# Check environment files
check_file ".env.example"
check_file ".env.staging"

# Check documentation
check_file "docs/CICD_SETUP.md"

echo ""

# Test 2: Validate Docker setup
log_test "Testing Docker configuration..."

if command -v docker &> /dev/null; then
    log_info "Docker is installed"
    
    # Test Docker build
    log_test "Testing Docker build..."
    if docker build -t sharespace-test . &> /dev/null; then
        log_info "Docker build successful"
        docker rmi sharespace-test &> /dev/null
    else
        log_error "Docker build failed"
    fi
else
    log_warn "Docker not installed - install for full testing"
fi

if command -v docker-compose &> /dev/null; then
    log_info "Docker Compose is installed"
    
    # Validate docker-compose files
    log_test "Validating docker-compose.yml..."
    if docker-compose config &> /dev/null; then
        log_info "docker-compose.yml is valid"
    else
        log_error "docker-compose.yml has syntax errors"
    fi
    
    log_test "Validating docker-compose.prod.yml..."
    if docker-compose -f docker-compose.prod.yml config &> /dev/null; then
        log_info "docker-compose.prod.yml is valid"
    else
        log_error "docker-compose.prod.yml has syntax errors"
    fi
else
    log_warn "Docker Compose not installed - install for full testing"
fi

echo ""

# Test 3: Check Go application
log_test "Testing Go application..."

if command -v go &> /dev/null; then
    log_info "Go is installed"
    
    # Test build
    log_test "Testing Go build..."
    if go build -o test-app ./Delivery &> /dev/null; then
        log_info "Go build successful"
        rm -f test-app
    else
        log_error "Go build failed"
    fi
    
    # Test dependencies
    log_test "Checking Go dependencies..."
    if go mod verify &> /dev/null; then
        log_info "Go modules are valid"
    else
        log_error "Go module issues detected"
    fi
else
    log_error "Go not installed - required for development"
fi

echo ""

# Test 4: Check environment configuration
log_test "Checking environment configuration..."

if [[ -f ".env" ]]; then
    log_info "Local .env file exists"
else
    log_warn "No local .env file - copy from .env.example for development"
fi

# Check for required environment variables in example
required_vars=("MONGODB_URI" "JWT_SECRET" "REFRESH_SECRET")
for var in "${required_vars[@]}"; do
    if grep -q "^$var=" .env.example; then
        log_info "Required variable $var found in .env.example"
    else
        log_error "Missing required variable $var in .env.example"
    fi
done

echo ""

# Test 5: GitHub repository checks
log_test "Checking GitHub repository setup..."

if git rev-parse --git-dir &> /dev/null; then
    log_info "Git repository initialized"
    
    # Check for remote
    if git remote -v | grep -q "origin"; then
        log_info "Git remote 'origin' configured"
        
        # Check if it's a GitHub repository
        if git remote get-url origin | grep -q "github.com"; then
            log_info "GitHub repository detected"
        else
            log_warn "Remote is not a GitHub repository"
        fi
    else
        log_warn "No git remote configured - needed for GitHub Actions"
    fi
    
    # Check current branch
    current_branch=$(git branch --show-current)
    log_info "Current branch: $current_branch"
    
    if [[ "$current_branch" == "main" || "$current_branch" == "develop" ]]; then
        log_info "On deployment branch ($current_branch)"
    else
        log_warn "Not on main/develop branch - CI/CD triggers on these branches"
    fi
else
    log_error "Not a git repository"
fi

echo ""

# Test 6: Security checks
log_test "Running security checks..."

# Check for secrets in code
log_test "Checking for potential secrets in code..."
if grep -r -i "password\|secret\|key" --include="*.go" --exclude-dir=".git" . | grep -v "// " | grep -v "example" | head -5; then
    log_warn "Potential secrets found in code - review and use environment variables"
else
    log_info "No obvious secrets found in code"
fi

# Check .gitignore
if [[ -f ".gitignore" ]]; then
    if grep -q ".env" .gitignore; then
        log_info ".env files are properly ignored by git"
    else
        log_warn ".env files should be added to .gitignore"
    fi
else
    log_warn "No .gitignore file found"
fi

echo ""

# Test 7: Test local development setup
log_test "Testing local development setup..."

if [[ -f ".env.example" ]]; then
    log_test "Creating test environment file..."
    cp .env.example .env.test
    
    # Add test values
    sed -i 's/your-super-secret-jwt-key-at-least-32-characters/test-jwt-secret-key-for-local-testing-32-chars/g' .env.test
    sed -i 's/your-super-secret-refresh-key-at-least-32-characters/test-refresh-secret-key-for-local-testing-32-chars/g' .env.test
    sed -i 's/mongodb:\/\/localhost:27017\/sharespace/mongodb:\/\/localhost:27017\/sharespace_test/g' .env.test
    
    log_info "Test environment file created (.env.test)"
    
    # Clean up
    rm -f .env.test
fi

echo ""

# Summary
echo "📊 CI/CD Implementation Summary"
echo "==============================="

echo ""
echo "🔧 Next Steps to Complete CI/CD Setup:"
echo ""
echo "1. 📝 Configure GitHub Secrets:"
echo "   - Go to GitHub repo → Settings → Secrets and variables → Actions"
echo "   - Add: DOCKER_USERNAME, DOCKER_PASSWORD"
echo "   - Add server credentials for deployment"
echo ""
echo "2. 🐳 Test Docker Build:"
echo "   docker build -t sharespace-backend ."
echo ""
echo "3. 🧪 Test Local Development:"
echo "   cp .env.example .env"
echo "   # Fill in your values"
echo "   docker-compose up -d"
echo ""
echo "4. 🚀 Test CI/CD Pipeline:"
echo "   git add ."
echo "   git commit -m 'feat: add CI/CD pipeline'"
echo "   git push origin develop  # Test staging deployment"
echo ""
echo "5. 📊 Monitor GitHub Actions:"
echo "   - Go to GitHub repo → Actions tab"
echo "   - Watch the workflow execution"
echo ""
echo "6. 🏥 Add Health Endpoint:"
echo "   - Implement /health endpoint in your router"
echo "   - See docs/HEALTH_ENDPOINT.md for details"
echo ""

echo "✨ CI/CD pipeline is ready for testing!"
echo "   Push to 'develop' branch to test staging deployment"
echo "   Push to 'main' branch to test production deployment"
