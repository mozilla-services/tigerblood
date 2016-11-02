# swagger_client.HeartbeatApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**heartbeat_get**](HeartbeatApi.md#heartbeat_get) | **GET** /__heartbeat__ | 


# **heartbeat_get**
> heartbeat_get()



### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.HeartbeatApi()

try: 
    api_instance.heartbeat_get()
except ApiException as e:
    print "Exception when calling HeartbeatApi->heartbeat_get: %s\n" % e
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

