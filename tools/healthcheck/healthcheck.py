import os
from requests_hawk import HawkAuth
import requests


def healthcheck(ip=None, url=None, hawk_id=None, hawk_key=None):
    "The function inserts an IP, updates it and deletes it, ensuring each operation was successful by looking up said IP afterwards."
    hawk_auth = HawkAuth(id=hawk_id or os.environ['HAWK_ID'],
                         key=hawk_key or os.environ['HAWK_KEY'])

    ip = ip or os.environ['TEST_IP']
    url = url or os.environ['SERVICE_URL']

    # clear test IP reputation
    get = requests.get(url + ip, auth=hawk_auth)
    if get.status_code != 404:
        requests.delete(url + ip, auth=hawk_auth)

    # check adding reputation
    post = requests.post(url, json={'ip': ip, 'reputation': 50}, auth=hawk_auth)
    assert post.status_code == 201, "instead of 201 add rep returned: %d %r %r" % (post.status_code, post.headers, post.text)
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.status_code == 200, "instead of 200 get added rep returned: %d %r %r" % (get.status_code, get.headers, get.text)
    add_expected_body = {u'IP': ip, u'Reputation': 50}
    assert get.json() == add_expected_body, "instead of %r get updated rep returned: %d %r %r" % (add_expected_body, get.status_code, get.headers, get.json())

    # check updating reputation
    put = requests.put(url + ip, json={'ip': ip, 'reputation': 70}, auth=hawk_auth)
    assert put.status_code == 200, "instead of 200 put reputation returned: %d %r %r" % (put.status_code, put.headers, put.text)
    get = requests.get(url + ip, auth=hawk_auth)
    update_expected_body = {u'IP': ip, u'Reputation': 70}
    assert get.json() == update_expected_body, "instead of %r get updated rep returned: %d %r %r" % (update_expected_body, get.status_code, get.headers, get.json())

    # check deleting reputation
    delete = requests.delete(url + ip, auth=hawk_auth)
    assert delete.status_code == 200, "instead of 200 del rep returned: %d %r %r" % (delete.status_code, delete.headers, delete.text)
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.status_code == 404, "instead of 404 get deleted rep returned: %d %r %r" % (get.status_code, get.headers, get.text)

    # check upserting reputation by violation
    put = requests.put(url + 'violations/' + ip, json={'ip': ip, 'Violation': 'test_violation'}, auth=hawk_auth)
    assert put.status_code == 204, "instead of 204 put violation returned: %d %r %r" % (put.status_code, put.headers, put.text)
    get = requests.get(url + ip, auth=hawk_auth)
    assert get.status_code == 200, "instead of 200 get violated rep returned: %d %r %r" % (get.status_code, get.headers, get.text)
    get_json = get.json()
    assert get_json['IP'] == ip, "instead of %s get violated rep returned ip: %s" % (ip, get_json['IP'])
    assert get_json['Reputation'] < 72, "instead of 72 get violated rep returned rep: %s" % get_json['Reputation']


def handler(event, context):
    "Entrypoint for AWS Lambda function."
    healthcheck()


if __name__ == '__main__':
    healthcheck()
