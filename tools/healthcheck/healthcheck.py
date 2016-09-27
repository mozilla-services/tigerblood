from requests_hawk import HawkAuth
import requests
import os


def healthcheck(event, context):
    hawk_auth = HawkAuth(
        id=os.environ['HAWK_ID'],
        key=os.environ['HAWK_KEY'])

    ip = u"127.0.0.1"
    url = "https://tigerblood.stage.mozaws.net/"

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
