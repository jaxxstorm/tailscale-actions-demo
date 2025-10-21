# Tailscale Actions Demo - Infrastructure

This directory contains the Terraform configuration for deploying the demo application to AWS ECS.

## Architecture

- **VPC**: 3 availability zones with public and private subnets
- **RDS PostgreSQL**: Database in private subnets
- **ECS Fargate**: Container orchestration for the application
- **Internal ALB**: Load balancer in private subnets
- **Secrets Manager**: Stores database credentials and Tailscale auth key
- **CloudWatch Logs**: Centralized logging for ECS tasks
- **GitHub Container Registry**: Docker image storage

## Prerequisites

1. AWS credentials configured
2. Terraform installed (>= 1.0)
3. Tailscale auth key
4. GitHub Personal Access Token with `write:packages` permission

## Deployment

### 1. Initialize Terraform

```bash
cd infra
terraform init
```

### 2. Create terraform.tfvars

```hcl
name                = "tailscale-actions-demo"
tailscale_auth_key  = "tskey-auth-xxxxx"
app_image          = "ghcr.io/jaxxstorm/tailscale-actions-demo:latest"

# Optional overrides
ecs_desired_count   = 2
ecs_task_cpu       = "256"
ecs_task_memory    = "512"
log_retention_days = 7

tags = {
  Environment = "demo"
  Project     = "tailscale-actions-demo"
}
```

### 3. Deploy Infrastructure

```bash
terraform plan
terraform apply
```

### 4. Build and Push Docker Image

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Build and push
cd ../app
docker build -t ghcr.io/jaxxstorm/tailscale-actions-demo:latest .
docker push ghcr.io/jaxxstorm/tailscale-actions-demo:latest
```

### 5. Run Database Migrations

After the infrastructure is deployed, run migrations:

```bash
# Get database details
terraform output db_instance_address
terraform output db_secret_arn

# Get password from Secrets Manager
aws secretsmanager get-secret-value --secret-id <secret-arn> --query SecretString --output text | jq -r '.password'

# Run migrations
cd ../app
export DB_URL="postgres://dbadmin:PASSWORD@RDS_ENDPOINT:5432/appdb?sslmode=require"
migrate -path migrations/initial -database "$DB_URL" up
migrate -path migrations/new_product -database "$DB_URL" up
```

## GitHub Actions CI/CD

The repository includes a GitHub Actions workflow (`.github/workflows/deploy.yml`) that:

1. Builds the Docker image on push to `main`
2. Pushes to GitHub Container Registry
3. Runs database migrations
4. Deploys to ECS

### Required Secrets

Configure these in your GitHub repository:

- `AWS_ROLE_ARN`: IAM role ARN for OIDC authentication
- `TAILSCALE_AUTH_KEY`: Tailscale authentication key (if not in tfvars)

### Setting up AWS OIDC for GitHub Actions

```bash
# Create OIDC provider (one-time setup)
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# Create IAM role with trust policy for GitHub Actions
# Add policies for ECS, RDS, Secrets Manager access
```

## Accessing the Application

The application is deployed behind an **internal ALB** and is only accessible from within the VPC or via Tailscale.

```bash
# Get ALB DNS name
terraform output app_url

# Access from bastion host or via Tailscale
curl http://<alb-dns-name>/health
curl http://<alb-dns-name>/api/products
```

## Outputs

- `app_url`: Internal URL to access the application
- `alb_dns_name`: DNS name of the load balancer
- `db_instance_endpoint`: RDS database endpoint
- `ecs_cluster_name`: Name of the ECS cluster
- `cloudwatch_log_group`: Log group for application logs

## Monitoring

### CloudWatch Logs

```bash
# View logs
aws logs tail /ecs/tailscale-actions-demo-app --follow

# View specific task logs
aws ecs list-tasks --cluster tailscale-actions-demo-cluster --service tailscale-actions-demo-app
aws logs tail /ecs/tailscale-actions-demo-app --follow --filter-pattern "ERROR"
```

### ECS Service Status

```bash
aws ecs describe-services \
  --cluster tailscale-actions-demo-cluster \
  --services tailscale-actions-demo-app
```

## Cleanup

```bash
terraform destroy
```

Note: If you have deletion protection enabled on the RDS instance or ALB, you'll need to disable it first.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                    GitHub Actions                    │
│  Build → Push to GHCR → Migrate DB → Deploy ECS    │
└────────────────────┬────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────┐
│                      AWS VPC                         │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │           Private Subnets                     │  │
│  │                                               │  │
│  │  ┌─────────────┐      ┌──────────────────┐  │  │
│  │  │  Internal   │      │   ECS Fargate    │  │  │
│  │  │     ALB     │─────▶│   (App Tasks)    │  │  │
│  │  └─────────────┘      └────────┬─────────┘  │  │
│  │                                 │            │  │
│  │                                 ▼            │  │
│  │                        ┌──────────────────┐  │  │
│  │                        │  RDS PostgreSQL  │  │  │
│  │                        └──────────────────┘  │  │
│  └──────────────────────────────────────────────┘  │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │  Secrets Manager                              │  │
│  │  • DB Password                                │  │
│  │  • Tailscale Auth Key                         │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Terraform Files

- `vpc.tf`: VPC configuration with public/private subnets
- `rds.tf`: PostgreSQL database
- `ecs.tf`: ECS cluster, task definition, service, ALB
- `ecs_iam.tf`: IAM roles for ECS execution and tasks
- `variables.tf`: Input variables
- `provider.tf`: AWS provider configuration

## Components

- **VPC**: Custom VPC with public and private subnets
- **EC2 Auto Scaling Group**: Tailscale subnet router instances
- **RDS PostgreSQL**: Database for your Python applications
- **IAM Roles**: Permissions for SSM and EC2 operations
- **Security Groups**: Network access control

## SSM Session Manager Connectivity

### Prerequisites

For SSM Session Manager to work, ensure:

1. **IAM Role**: Instance must have the `AmazonSSMManagedInstanceCore` policy attached
2. **Network Connectivity**: Instance needs to reach SSM endpoints via:
   - Public subnet with Internet Gateway (current setup), OR
   - VPC endpoints for SSM services in private subnets

3. **SSM Agent**: Must be running on the instance (pre-installed on Amazon Linux 2023)

### Troubleshooting SSM

If you can't connect via SSM:

```bash
# Check if instance is registered with SSM
aws ssm describe-instance-information --filters "Key=InstanceIds,Values=i-xxxxx"

# Check SSM agent status (via SSH or user data logs)
sudo systemctl status amazon-ssm-agent

# View SSM agent logs
sudo journalctl -u amazon-ssm-agent -f

# Restart SSM agent if needed
sudo systemctl restart amazon-ssm-agent
```

### Common Issues

1. **Instance not appearing in SSM**: 
   - Check IAM role has `AmazonSSMManagedInstanceCore` policy
   - Verify outbound HTTPS (443) is allowed in security group
   - Ensure instance can reach internet (public IP or NAT gateway)

2. **Connection timeout**:
   - Check VPC route table has route to Internet Gateway
   - Verify security group allows outbound traffic
   - Confirm SSM endpoints are reachable

3. **In private subnets**: Add VPC endpoints for:
   - `com.amazonaws.<region>.ssm`
   - `com.amazonaws.<region>.ssmmessages`
   - `com.amazonaws.<region>.ec2messages`

## Database Connection

The PostgreSQL database is deployed in private subnets and accessible from within the VPC.

### Connection Details

After deployment, retrieve connection info:

```bash
# Get database endpoint
terraform output db_instance_endpoint

# Get secret ARN (contains password)
terraform output db_secret_arn

# Retrieve credentials from Secrets Manager
aws secretsmanager get-secret-value --secret-id <secret-arn> --query SecretString --output text | jq
```

### Python Connection Example

See `app/db_example.py` for a complete example:

```python
# Set environment variable
export DB_SECRET_ARN=$(terraform output -raw db_secret_arn)

# Install dependencies
pip install -r app/requirements.txt

# Run example
python app/db_example.py
```

## Deployment

```bash
# Initialize Terraform
terraform init

# Review plan
terraform plan

# Apply configuration
terraform apply

# Connect to instance via SSM
aws ssm start-session --target <instance-id>
```

## Variables

Key variables you can customize in `terraform.tfvars`:

```hcl
name               = "my-tailscale-router"
tailscale_auth_key = "tskey-auth-xxxxx"
advertise_tags     = ["tag:my-tag"]

# Database
db_name            = "myapp"
db_username        = "admin"
db_instance_class  = "db.t3.small"

# EC2
instance_type      = "t3.medium"
enable_aws_ssm     = true
```

## Security

- Database password is automatically generated and stored in AWS Secrets Manager
- Instances use IMDSv2 for enhanced security
- Database is in private subnets, not publicly accessible
- Security groups restrict access appropriately

## Cleanup

```bash
terraform destroy
```

**Note**: If `db_deletion_protection = true`, you'll need to disable it first or use:

```bash
terraform apply -var="db_deletion_protection=false"
terraform destroy
```
