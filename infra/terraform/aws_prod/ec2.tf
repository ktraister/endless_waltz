data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "ultron_cluster" {
  count           = var.instance_count
  ami             = "ami-05fb0b8c1424f266b"
  instance_type   = "t3.medium"
  subnet_id       = module.vpc.public_subnets[count.index]
  vpc_security_group_ids = [aws_security_group.prod_clusters.id]
  key_name        = "aws"

  tags = {
    Name = "Ultron_cluster-${count.index}"
  }
}

resource "aws_instance" "epion_cluster" {
  count           = 1
  ami             = data.aws_ami.ubuntu.id
  instance_type   = "t3.medium"
  subnet_id       = module.vpc.public_subnets[count.index]
  vpc_security_group_ids = [aws_security_group.prod_clusters.id]
  key_name        = "aws"

  tags = {
    Name = "Epion_cluster-${count.index}"
  }
}

resource "aws_security_group" "prod_clusters" {
  name        = "prod_clusters_sg"
  description = "Allow prod k3s clusters inbound and outbound traffic"
  vpc_id      = module.vpc.vpc_id
  tags = {
    Name = "prod_clusters_sg"
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_22" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = var.home_plate_ip
  from_port         = 22
  ip_protocol       = "tcp"
  to_port           = 22
  description       = "ssh from home plate"
}

resource "aws_vpc_security_group_ingress_rule" "allow_6443" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = var.home_plate_ip
  from_port         = 6443
  ip_protocol       = "tcp"
  to_port           = 6443
  description       = "Kubectl from home plate"
}

resource "aws_vpc_security_group_ingress_rule" "allow_80" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  ip_protocol       = "tcp"
  to_port           = 80
}

resource "aws_vpc_security_group_ingress_rule" "allow_443" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  ip_protocol       = "tcp"
  to_port           = 443
}

resource "aws_vpc_security_group_ingress_rule" "allow_8080" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 8080
  ip_protocol       = "tcp"
  to_port           = 8080
}

resource "aws_security_group_rule" "allow_intracluster_traffic" {
  security_group_id = aws_security_group.prod_clusters.id
  source_security_group_id = aws_security_group.prod_clusters.id
  type              = "ingress"
  protocol          = "all"
  to_port           = 0
  from_port         = 0
}

resource "aws_vpc_security_group_egress_rule" "allow_all_traffic" {
  security_group_id = aws_security_group.prod_clusters.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1" # semantically equivalent to all ports
}

