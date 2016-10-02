from requests_hawk import HawkAuth
import requests
import json


def healthcheck(event, context):
    with open('config.json') as config_file:
        config = json.load(config_file)
    hawk_auth = HawkAuth(
        id=config['hawk_id'],
        key=config['hawk_key'])

    ip = config['ip']
    url = config['url']

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
