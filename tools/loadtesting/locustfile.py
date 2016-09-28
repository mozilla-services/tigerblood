from requests_hawk import HawkAuth
from locust import HttpLocust, TaskSet, task
import random
import os


class UserBehavior(TaskSet):
    hawk_auth = HawkAuth(
        id=os.environ['HAWK_ID'],
        key=os.environ['HAWK_KEY'])

    @task(1)
    def index(self):
        ip = "{}.{}.{}.{}".format(*(random.randint(0, 255) for _ in range(4)))
        self.client.get("/{}".format(ip), auth=self.hawk_auth)


class WebsiteUser(HttpLocust):
    task_set = UserBehavior
    min_wait = 0
    max_wait = 0
