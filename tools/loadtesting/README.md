# Running the load tests

To run the load tests on your machine, you just need to have locust install. You can install it with `pip install locustio`. Then, to actually run the tests:
* Ensure the tigerblood server is running
* Run `locust --host http://localhost:{TIGERBLOOD_PORT}`
* Open the web ui at http://localhost:8089
* Select the number of users (concurrent connections) you want to simulate, and the hatch rate (users added per second).

To run the test on a distributed fashion on AWS, you're going to need to `pip install fabric boto` and set up your AWS credentials.
There are several fabric commands of interest:

## `spin_up_instances`

Use this command to automatically run EC2 instances to perform the load test. It will tag the instances so they can be identified later.

Parameters:

* ssh_key: Mandatory, the name of the AWS SSH key you want to use
* ec2_master_security_group: The EC2 security group that'll be assigned to the instances. TCP ports 5557 and 22 must be open.
* slave_count: Optional (defaults to 1). The number of slaves you want to perform the test with.
* aws_region: Optional (defaults to us-west-2).
* ami: Optional (defaults to Ubuntu's AMI for us-west-2). This has to be some flavor of debian.
* ec2_instance_size: Optional (defaults to t2.micro).

## set_hosts

Parameters:

* aws_region: Optional (defaults to us-west-2)

This command sets fabric's hosts for future tasks. You have to run this before all commands that require running anything on the servers.

## install_locust

Installs locust on all the servers. set_hosts must be passed to fabric first. For example:

`fab set_hosts install_locust -P -u ubuntu`

`-P` just makes the task run in parallel on all hosts.

## run_master

Runs the locust master on one of the servers

## run_slaves

Runs the locust slave on all but one server

Generally, you'll want to run:

`fab set_hosts run_master:"http://your-tigerblood.com" run_slaves:"http://your-tigerblood.com" -u ubuntu -P`

## stop

Stops locust

## spin_down_instances

This will terminate all EC2 instances tagged with the tag `spin_up_instances` sets.
