
module "squid_proxy" {
  source               = "../squid"
  csoc_cidr            = "${var.csoc_cidr}"
  env_vpc_name         = "${var.vpc_name}"
  env_public_subnet_id = "${aws_subnet.public.id}"
  env_vpc_cidr         = "${aws_vpc.main.cidr_block}"
  env_vpc_id           = "${aws_vpc.main.id}"
  ssh_key_name         = "${var.ssh_key_name}"
  env_instance_profile = "${aws_iam_instance_profile.cluster_logging_cloudwatch.name}"
  env_log_group        = "${aws_cloudwatch_log_group.main_log_group.name}"
}

module "data-bucket" {
  source               = "../upload-data-bucket"
  vpc_name             = "${var.vpc_name}"
  cloudwatchlogs_group = "${aws_cloudwatch_log_group.main_log_group.arn}"
  environment          = "${var.vpc_name}"
}

module "fence-bot-user" {
  source               = "../fence-bot-user"
  vpc_name             = "${var.vpc_name}"
  bucket_name          = "${module.data-bucket.data-bucket_name}"
}


