# Commons built up steps 

The following guide is intended to guide you through the process of bringing up a gen3 commons.


## First part, setting up the adminVM

1. Clone the repo
```bash
git clone https://github.com/uc-cdis/cloud-automation.git
```

2. If no proxy is needded then 
```bash
export GEN3_NOPROXY='no'
```
   If a proxy is required, then gen3 would assume cloud-proxy.internal.io:3128 is your proxy for http and https

3. Install dependencies, you must run this part as a sudo access user
```bash 
bash cloud-automation/gen3/bin/kube-setup-workvm.sh
```

4. kube-setup-workvm.sh add a few required configurations to the user's local bashrc file. To be able to use them, we may want to source it, otherwise we'll have to logout and in again.
```bash
source .bashrc
```

5. Edit the local aws config file by adding a profile additional to default, even if it's the same info as the default. It should look something like: 
```
  ubuntu@ip-172-31-40-144:~$ cat .aws/config 
  [default]
  output = json
  region = us-east-1
  credential_source = Ec2InstanceMetadata

  [profile cdistest]
  output = json
  region = us-east-1
  credential_source = Ec2InstanceMetadata
```

  It worth noting that additional information may be required in this file, everything would depend on your setup for the VM in question.



## Second part start gen3

1. Initialize the base module
```bash
gen3 workon cdistest commons-test
```

  Note: The third argument of the above command (cdistest) refers to the profile in the config file setup in step five of the fist part.
        The forth argument (commons-test) would be the name of commons you want to use, only lowercase letters and hyphen are permitted.

2. Go to the terraform workspace folder
```bash
gen3 cd
```

3. Edit the `config.tfvars` file with a text editor of prefference. 

  Variables to pay attention to:

`vpc_cidr_block` CIDR where the commons resources would reside. EX: 172.16.192.0/20. As for now, only /20 subnets are supported. Your VPC must have only RFC1918 or CG NAT CIDRs.

`dictionary_url` url where the dictionary schema is, it must be in json format

`portal_app`

`aws_cert_name` AWS ARN for the certificate to use on the Load Balancer that will be in front for HTTPS access.

`hostname` domain which the commons will respond to

`config_folder` folder for permissions. By default, commons would try to load a user.yaml file from s3://cdis-gen3-users/CONFIG_FOLDER/user.yaml. This bucket is not publicly accessible, you can later set a different one though. Just keep in mind that the folder with the name you are setting this var with needs to exist within the bucket and a user.yaml file within the folder in question. You can still set permissions based on a local file

`google_client_secret` and `google_client_id`  Google set of API key so you can set google authentication.

`csoc_managed` if you are going to set up your commons hooked up to a central control management account. By default it is set to yes, any other value would assume that you don't want this to happen. If you leave the default value, you must run the logging module first, otherwise terraform will fail.

`csoc_vpc_id` VPC id from where you are running gen3 commands, must be in the same region as where you are running gen3.

*Note:* If you are hooking up your commons with a cetralized control management account, you may need to add additional variables to this file with more information about said account.


4. Create a terraform plan
```bash
gen3 tfplan
```
  You may want to review what will be created by terraform by going through the outputed plan.

5. Apply the previously created plan
```bash
gen3 tfapply
```

6. Copy the newly commons-test_output folder created to the user's home folder. Keep in mind that you'll see the folder if you haven't `cd` onto a different folder after running `gen3 cd`
```bash
cp -r commons-test_output/ $HOME
```



## Third part, deploy the kubernetes cluster

1. Initialize the EKS module
```bash
gen3 workon cdistest commons-test_eks
```

  Note: The third argument of the above command (cdistest) refers to the profile in the config file setup in step five of the fist part.
        The forth argument would be the name of commons you want to use, only lowercase letters and hyphen are permitted. You must add `_eks` to the name in order to invoke the EKS module.

2. Go to the terraform workspace folder
```bash
gen3 cd
```

3. Edit the `config.tfvars` file with a text editor of prefference. 

  Variables to pay attention to:

`vpc_name` name of the commons it *MUST* be the same one used in part two.

`users_policy` this is the name of the policy that allows access to the user.yaml file mentioned in part two. This variable value should always be the same as the above one, but it might differ in very specific cases.

`instance_type` default set to t2.medium. Change if necessary.

`ec2_keyname` an existing Key Pair in EC2 for the workers for deployment. More keys can be added automatically if you specify them in $HOME/cloud-automation/files/authorized_keys/ops_team.


4. Create a terraform plan
```bash
gen3 tfplan
```
  You may want to review what will be created by terraform by going through the outputed plan.

5. Apply the previously created plan
```bash
gen3 tfapply
```

6. The EKS module creates a kubernetes configuration file (kubeconfig), copy it to the user's home folder.
```bash
cp commons-test_output_EKS/kubeconfig $HOME
```



## Fourth part, bring up services in kubernetes

1. Access the folder copied to the home folder
```bash
cd $HOME/commons-test_output/
```

2. Run `kube-up.sh`
```bash
bash kube-up.sh
```

4. Move the kubeconfig file we copied previously into a newly created folder that `kube-up.sh` created for us.
```bash
mv $HOME/kubconfig $HOME/commons-test/
```

3. Create a manifest folder
```bash
mkdir -p $HOME/cdis-manifest/commons-test.plantx-pla.net
```

  Note: The cdis-manifest folder is required, if you want to use your own manifest folder name you must make changes to the code, the file containing the line is `cloud-automation/gen3/lib/g3k_manifest.sh`.
        Moreover, a subfolder named the same as your hostname is required.

4. kube-up.sh added a few lines to our local bashrc file, let's load the up.
```bash
source $HOME/.bashrc
```

5. Verify that kubernetes is up. After sourcing our local bashrc file we should be able to talk to kubernetes:
```bash
kubectl get nodes
```

6. Roll services
```bash
gen3 roll all
```
  Note: it might take a few minutes to complete, let it run.

7. Get the newly created ELB endpoint so you can point your domain to it.
```bash
kubectl get service revproxy-service-elb -o json | jq -r .status.loadBalancer.ingress[].hostname
```

8. Go to your registrar and point the desired domain to the outcome of above command.


