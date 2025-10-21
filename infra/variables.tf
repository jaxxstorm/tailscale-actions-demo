

variable "name" {
  description = "Name of the subnet router"
  type        = string
}

variable "enable_aws_ssm" {
  description = "Enable AWS SSM for the instance"
  type        = bool
  default     = true
  
}

variable "tags" {
  description = "EC2 tags to apply to the instance"
  type        = map(string)
  default     = {}
}

variable "architecture" {
  description = "Architecture of the instance"
  type        = string
  default     = "x86_64"
}

variable "instance_type" {
  description = "Instance type to use for the subnet routers"
  default     = "t3.medium"
  type        = string
}

variable "ebs_root_volume_size" {
  description = "Size of the root volume in GB"
  default     = 20
  type        = number
}


variable "tailscale_auth_key" {
  description = "Tailscale authentication key"
  type        = string
  sensitive   = true
}

variable "advertise_tags" {
  description = "Tags to advertise for the subnet routers"
  type        = list(string)
  default     = []
}

variable "hostname" {
  description = "The hostname for the Tailscale ec2 instances"
  type = string
  default = "subnet-router"
}
# Database variables
variable "db_name" {
  description = "The name of the database to create"
  type        = string
  default     = "appdb"
}

variable "db_username" {
  description = "Master username for the database"
  type        = string
  default     = "dbadmin"
}

variable "db_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "16.3"
}

variable "db_instance_class" {
  description = "The instance type of the RDS instance"
  type        = string
  default     = "db.t3.micro"
}

variable "db_allocated_storage" {
  description = "The allocated storage in gigabytes"
  type        = number
  default     = 20
}

variable "db_backup_retention_period" {
  description = "The days to retain backups for"
  type        = number
  default     = 7
}

variable "db_skip_final_snapshot" {
  description = "Determines whether a final DB snapshot is created before deletion"
  type        = bool
  default     = false
}

variable "db_deletion_protection" {
  description = "Enable deletion protection on the RDS instance"
  type        = bool
  default     = false
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key file"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "ssh_allowed_cidrs" {
  description = "CIDR blocks allowed to SSH to instances"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # Consider restricting this to your IP
}

# ECS Variables
variable "app_image" {
  description = "Docker image for the application (from GitHub Container Registry)"
  type        = string
  default     = "ghcr.io/jaxxstorm/tailscale-actions-demo:latest"
}

variable "ecs_task_cpu" {
  description = "CPU units for the ECS task (1 vCPU = 1024)"
  type        = string
  default     = "256"
}

variable "ecs_task_memory" {
  description = "Memory for the ECS task in MB"
  type        = string
  default     = "512"
}

variable "ecs_desired_count" {
  description = "Desired number of ECS tasks"
  type        = number
  default     = 2
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 7
}

variable "alb_deletion_protection" {
  description = "Enable deletion protection for the ALB"
  type        = bool
  default     = false
}

variable "github_token" {
  description = "GitHub Personal Access Token for pulling from GHCR"
  type        = string
  sensitive   = true
}
