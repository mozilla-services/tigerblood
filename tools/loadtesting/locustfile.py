from requests_hawk import HawkAuth
from locust import HttpLocust, TaskSet, task
import random
import os


test_violations = [
  'fxa:request.blockIp',
  'fxa:request.checkAuthenticated.block.devicesNotify',
  'fxa:request.checkAuthenticated.block.avatarUpload',
  'fxa:request.failedLoginAttempt.isOverBadLogins',
  'fxa:request.check.block.accountCreate',
  'fxa:request.check.block.accountLogin',
  'fxa:request.check.block.accountStatusCheck',
  'fxa:request.check.block.recoveryEmailResendCode',
  'fxa:request.check.block.recoveryEmailVerifyCode',
  'fxa:request.check.block.sendUnblockCode',
  'fxa:request.check.block.accountDestroy',
  'fxa:request.check.block.passwordChange',
  'test_violation'
]


def random_ipv4():
    return "{}.{}.{}.{}".format(*(random.randint(0, 255) for _ in range(4)))


class UserBehavior(TaskSet):
    hawk_auth = HawkAuth(
        id=os.environ['HAWK_ID'],
        key=os.environ['HAWK_KEY'])

    @task(1000)
    def get_ip(self):
        with self.client.get("{}".format(random_ipv4()),
                             auth=self.hawk_auth,
                             name='get_ip',
                             catch_response=True) as resp:
            if resp.status_code == 404:
                resp.success()

    @task(1)
    def report_ip(self):
        uri = "violations/{ip}".format(ip=random_ipv4())
        self.client.put(uri,
                        json={"Violation": random.choice(test_violations)},
                        auth=self.hawk_auth,
                        name='report_ip')


class WebsiteUser(HttpLocust):
    task_set = UserBehavior
    min_wait = 0
    max_wait = 0
