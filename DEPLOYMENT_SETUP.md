# Deployment Setup Guide

## Prerequisites

1. **AWS Account** with:
   - ECR repository created
   - EC2 instance running (Ubuntu recommended)
   - IAM role for GitHub Actions

2. **GitHub Repository** with:
   - This code pushed to main branch
   - Secrets configured

## Step 1: Set Up AWS ECR Repository

```bash
# Create ECR repository
aws ecr create-repository \
  --repository-name mock-server \
  --region us-east-1
```

## Step 2: Set Up EC2 Instance

### Install Docker on EC2

```bash
# SSH into your EC2 instance
ssh -i your-key.pem ubuntu@<ec2-instance-ip>

# Update system
sudo apt-get update
sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add ubuntu user to docker group
sudo usermod -aG docker ubuntu

# Install AWS CLI
sudo apt-get install -y awscli

# Log out and back in for group changes to take effect
exit
```

## Step 3: Set Up GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add the following secrets:

| Secret Name | Value |
|-------------|-------|
| `AWS_ACCOUNT_ID` | Your AWS Account ID |
| `EC2_HOST` | Your EC2 instance IP or DNS |
| `EC2_USER` | ubuntu (or your AMI user) |
| `EC2_SSH_KEY` | Your EC2 SSH private key (full key content) |
| `AWS_ROLE_TO_ASSUME` | IAM role ARN for GitHub Actions (optional) |

## Step 4: Create IAM Role for GitHub Actions (Optional but Recommended)

### Using OIDC Provider

1. Add GitHub as an OIDC provider in IAM
2. Create a role with ECR push permissions
3. Set the trust relationship to allow GitHub Actions

```bash
# Example trust policy
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<account-id>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:<github-org>/<repo>:*"
        }
      }
    }
  ]
}
```

## Step 5: Configure Workflow

Edit `.github/workflows/deploy.yml` and update:

- `AWS_REGION` - Your AWS region
- `ECR_REPOSITORY` - Your ECR repository name

## Step 6: Deploy

1. Ensure EC2 security group allows:
   - SSH on port 22 (from GitHub Actions IPs or your IP)
   - HTTP on port 8080 (for accessing the application)

2. Push code to main branch:
```bash
git push origin main
```

3. Check GitHub Actions tab for workflow execution

## Monitoring

### View logs on EC2

```bash
# Check running containers
sudo docker ps

# View application logs
sudo docker logs -f mock-server

# Check container stats
sudo docker stats
```

### Verify deployment

```bash
# Test health endpoint
curl http://<ec2-instance-ip>:8080/health

# Test create student
curl -X POST http://<ec2-instance-ip>:8080/students \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Student","email":"test@example.com","age":20}'
```

## Troubleshooting

### Container won't start

```bash
# Check logs
sudo docker logs mock-server

# Check if port 8080 is in use
sudo netstat -tulpn | grep 8080

# Kill process if needed
sudo lsof -ti:8080 | xargs kill -9
```

### ECR login fails

```bash
# Manually login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
```

### SSH key authentication fails

- Ensure SSH key has correct permissions: `chmod 600 key.pem`
- Verify EC2 security group allows SSH access
- Check that EC2 instance is running and accessible

## Advanced: Using docker-compose on EC2

Create `/home/ubuntu/docker-compose.yml`:

```bash
sudo cat > /home/ubuntu/docker-compose.yml << 'EOF'
version: '3.8'

services:
  mock-server:
    image: <account-id>.dkr.ecr.us-east-1.amazonaws.com/mock-server:latest
    container_name: mock-server
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
    restart: always

EOF
```

Then in the deploy step, use:
```bash
cd /home/ubuntu && sudo docker-compose up -d
```

## Clean Up

To stop and remove the container:

```bash
sudo docker stop mock-server
sudo docker rm mock-server
```

To delete ECR repository:

```bash
aws ecr delete-repository --repository-name mock-server --force --region us-east-1
```
