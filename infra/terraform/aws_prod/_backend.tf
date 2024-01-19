terraform {
  backend "s3" {
    bucket = "ew-prod-terraform-backend"
    key    = "aws_prod"
    region = "us-east-2"
  }
}
