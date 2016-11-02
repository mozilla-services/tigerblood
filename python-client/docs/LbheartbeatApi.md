# swagger_client.LbheartbeatApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**lbheartbeat_get**](LbheartbeatApi.md#lbheartbeat_get) | **GET** /__lbheartbeat__ | 


# **lbheartbeat_get**
> lbheartbeat_get()



### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.LbheartbeatApi()

try: 
    api_instance.lbheartbeat_get()
except ApiException as e:
    print "Exception when calling LbheartbeatApi->lbheartbeat_get: %s\n" % e
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

