from fabric.api import run, env, sudo, put, task, runs_once, parallel, roles
import boto.ec2
import time


@task
def spin_up_instances(ssh_key,
                      ec2_master_security_group,
                      slave_count=1,
                      aws_region='us-west-2',
                      ami='ami-d732f0b7',
                      ec2_instance_size='t2.micro'):
    conn = boto.ec2.connect_to_region(aws_region)
    slave_count = int(slave_count)
    reservation = conn.run_instances(
        ami,
        min_count=slave_count + 1,
        max_count=slave_count + 1,
        key_name=ssh_key,
        instance_type=ec2_instance_size,
        security_groups=[ec2_master_security_group],
    )
    while not all(i.update() != 'pending' for i in reservation.instances):
        time.sleep(10)
    for instance in reservation.instances:
        instance.add_tag("App", "tigerblood-load-testing")


@task
@runs_once
def set_hosts(aws_region='us-west-2'):
    conn = boto.ec2.connect_to_region(aws_region)
    reservations = conn.get_all_instances(
        filters={"tag:App": "tigerblood-load-testing"})
    hosts = [i.ip_address for r in reservations
             for i in r.instances if i.state == 'running']
    env.roledefs = {
        'master': [hosts[0]],
        'slaves': hosts[1:]
    }
    env.hosts = hosts
    print("Loaded hosts: ", env.roledefs)


@task
def install_locust():
    sudo("rm -Rf /tmp/*")
    sudo("apt-get -qq update -y")
    sudo("apt-get -q install tmux gcc python-dev python-zmq python-pip -y")
    run("pip install locustio requests_hawk --user")
    put('locustfile.py', 'locustfile.py')


@task
@parallel
@roles('master')
def run_master(host):
    print("Running master on", env.host)
    run("tmux new -d -s locust '~/.local/bin/locust --master --host {}'"
        .format(host))


@task
@parallel
@roles('slaves')
def run_slaves(host):
    run("tmux new -d -s locust '~/.local/bin/locust "
        "--slave --master-host={} --host {}'"
        .format(env.roledefs['master'][0], host))


@task
def stop():
    run("tmux kill-session -t locust || true")


@task
def spin_down_instances(aws_region='us-west-2'):
    conn = boto.ec2.connect_to_region(aws_region)
    reservations = conn.get_all_instances(
        filters={"tag:App": "tigerblood-load-testing"})
    hosts = [i.id for r in reservations
             for i in r.instances if i.state == 'running']
    conn.terminate_instances(instance_ids=hosts)
