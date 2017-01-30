# Healthcheck lambda function

The function inserts an IP, updates it and deletes it, ensuring each operation was successful by looking up said IP afterwards.

Requires the environment vars:

* HAWK_ID
* HAWK_KEY
* SERVICE_URL
* TEST_IP
