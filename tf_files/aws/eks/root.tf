terraform {
  backend "s3" {
    encrypt = "true"
  }
}

provider "aws" {}


module "eks" {
  source          = "../modules/eks"
  vpc_name        = "${var.vpc_name}"
  ec2_keyname     = "${var.ec2_keyname}"
  instance_type   = "${var.instance_type}"
  peering_cidr    = "${var.peering_cidr}"
  users_policy    = "${var.users_policy}"
#  csoc_managed    = "${var.csoc_managed}"
}

