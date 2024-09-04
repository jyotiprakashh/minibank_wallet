variable "region" {
  description = "The AWS region to deploy resources in"
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "The CIDR block for the VPC"
  default     = "10.0.0.0/16"
}

variable "db_instance_class" {
  description = "The instance type for the PostgreSQL database"
  default     = "db.t2.micro"
}

variable "db_name" {
  description = "The name of the PostgreSQL database"
  default     = "bank"
}

variable "db_username" {
  description = "The username for the PostgreSQL database"
  default     = "jyoti"
}

variable "db_password" {
  description = "The password for the PostgreSQL database"
  default     = "12345789"
}

variable "db_allocated_storage" {
  description = "The allocated storage for the PostgreSQL database"
  default     = 20
}
