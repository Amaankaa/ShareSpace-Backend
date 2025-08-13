# ShareSpace Backend CI/CD Setup Guide

This guide will help you set up a complete CI/CD pipeline for the ShareSpace backend using GitHub Actions, Docker, and various deployment options.

## 🏗️ Architecture Overview

```
GitHub Repository
    ↓
GitHub Actions (CI/CD)
    ↓
Docker Hub (Container Registry)
    ↓
Deployment Targets (Staging/Production)
```

## 📋 Prerequisites

### 1. GitHub Repository Setup
- Push your code to a GitHub repository
- Enable GitHub Actions in repository settings

### 2. Docker Hub Account
- Create account at [Docker Hub](https://hub.docker.com)
- Create a repository for your application

### 3. Server Requirements (for deployment)
- Ubuntu 20.04+ or similar Linux distribution
- Docker and Docker Compose installed
- SSH access configured

## 🔧 Configuration Steps

### Step 1: GitHub Secrets Configuration

Add the following secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

#### Docker Configuration
```
DOCKER_USERNAME=your-dockerhub-username
DOCKER_PASSWORD=your-dockerhub-password-or-token
```

#### Staging Environment
```
STAGING_SSH_KEY=your-staging-server-private-key
STAGING_USER=your-staging-server-username
STAGING_HOST=your-staging-server-ip-or-domain
STAGING_URL=https://staging.yourdomain.com
```

#### Production Environment
```
PRODUCTION_SSH_KEY=your-production-server-private-key
PRODUCTION_USER=your-production-server-username
PRODUCTION_HOST=your-production-server-ip-or-domain
PRODUCTION_URL=https://yourdomain.com
```

#### Optional: Notifications
```
SLACK_WEBHOOK=your-slack-webhook-url
```

### Step 2: Environment Files Setup

1. Copy `.env.example` to `.env.staging` and `.env.production`
2. Fill in the appropriate values for each environment
3. **Never commit these files to version control**

### Step 3: Server Setup

#### Install Docker and Docker Compose
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

#### Create Application Directory
```bash
sudo mkdir -p /opt/sharespace
sudo chown $USER:$USER /opt/sharespace
cd /opt/sharespace
```

#### Setup SSH Keys
```bash
# Generate SSH key pair (if not exists)
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"

# Add public key to authorized_keys
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys

# Copy private key content to GitHub secrets
cat ~/.ssh/id_rsa
```

### Step 4: Domain and SSL Setup (Production)

#### Configure Domain
1. Point your domain to your server's IP address
2. Update DNS A records

#### Setup SSL with Let's Encrypt
```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Generate SSL certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## 🚀 Deployment Options

### Option 1: Automatic Deployment (Recommended)

The GitHub Actions workflow automatically deploys:
- `develop` branch → Staging environment
- `main` branch → Production environment

### Option 2: Manual Deployment

```bash
# Clone repository
git clone https://github.com/yourusername/sharespace-backend.git
cd sharespace-backend

# Deploy to staging
./scripts/deploy.sh staging

# Deploy to production
./scripts/deploy.sh production
```

### Option 3: Docker Compose Only

```bash
# For development
docker-compose up -d

# For production
docker-compose -f docker-compose.prod.yml up -d
```

## 📊 Monitoring Setup

### Prometheus + Grafana (Optional)

The Docker Compose files include monitoring services:

1. **Prometheus**: Metrics collection (port 9090)
2. **Grafana**: Visualization (port 3000)

Access Grafana at `http://your-server:3000` with admin/admin credentials.

### Log Management

Logs are automatically managed with rotation:
- Application logs: `./logs/`
- Nginx logs: `./nginx/logs/`
- Container logs: Docker's built-in log rotation

## 🔍 Health Checks

The pipeline includes several health checks:

1. **Build Health**: Code compilation and tests
2. **Security Scan**: Vulnerability detection
3. **Deployment Health**: Service availability
4. **Smoke Tests**: Basic functionality verification

## 🛠️ Troubleshooting

### Common Issues

#### 1. Docker Build Fails
```bash
# Check Docker daemon
sudo systemctl status docker

# Check disk space
df -h

# Clean up Docker
docker system prune -a
```

#### 2. SSH Connection Issues
```bash
# Test SSH connection
ssh -o StrictHostKeyChecking=no user@server

# Check SSH key permissions
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
```

#### 3. Database Connection Issues
```bash
# Check MongoDB container
docker logs sharespace_mongodb

# Check network connectivity
docker network ls
docker network inspect sharespace_sharespace-network
```

### Rollback Procedure

#### Automatic Rollback
The deployment script includes automatic rollback on failure.

#### Manual Rollback
```bash
# Stop current version
docker-compose -f docker-compose.prod.yml down

# Restore from backup
docker exec sharespace_mongodb mongorestore /backups/backup_name

# Deploy previous version
docker-compose -f docker-compose.prod.yml up -d
```

## 📈 Performance Optimization

### Production Optimizations

1. **Resource Limits**: Set appropriate CPU/memory limits
2. **Load Balancing**: Use multiple API instances
3. **Caching**: Implement Redis caching
4. **CDN**: Use CDN for static assets
5. **Database**: Optimize MongoDB indexes

### Scaling Considerations

```yaml
# docker-compose.prod.yml
services:
  api:
    deploy:
      replicas: 3  # Scale API instances
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
```

## 🔒 Security Best Practices

1. **Secrets Management**: Use GitHub secrets, never commit credentials
2. **Network Security**: Use private networks, firewall rules
3. **Container Security**: Run as non-root user, scan for vulnerabilities
4. **SSL/TLS**: Always use HTTPS in production
5. **Regular Updates**: Keep dependencies and base images updated

## 📞 Support

For issues with the CI/CD setup:

1. Check GitHub Actions logs
2. Review server logs: `docker-compose logs`
3. Verify environment configuration
4. Test manual deployment steps

Remember to customize all configuration files with your actual values before deployment!
