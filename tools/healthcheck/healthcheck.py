import os
from requests_hawk import HawkAuth
import requests


def healthcheck(ip=None, url=None, hawk_id=None, hawk_key=None):
    "The function inserts an IP, updates it and deletes it, ensuring each operation was successful by looking up said IP afterwards."
    hawk_auth = HawkAuth(id=hawk_id or os.environ['HAWK_ID'],
                         key=hawk_key or os.environ['HAWK_KEY'])

    ip = ip or os.environ['TEST_IP']
    url = url or os.environ['SERVICE_URL']

    get = requests.get(url + ip, auth=hawk_auth)
    if get.status_code != 404:
        requests.delete(url + ip, auth=hawk_auth)

    post = requests.post(url, json={'ip': ip, 'reputation': 50}, auth=hawk_auth)
    assert post.status_code == 201
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.status_code == 200
    assert get.json() == {u'IP': ip, u'Reputation': 50}

    put = requests.put(url + ip, json={'ip': ip, 'reputation': 70}, auth=hawk_auth)
    assert put.status_code == 200
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.json() == {u'IP': ip, u'Reputation': 70}

    delete = requests.delete(url + ip, auth=hawk_auth)
    assert delete.status_code == 200
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.status_code == 404


def handler(event, context):
    "Entrypoint for AWS Lambda function."
    healthcheck()


if __name__ == '__main__':
    healthcheck()
