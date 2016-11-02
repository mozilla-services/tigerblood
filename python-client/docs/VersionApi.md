# swagger_client.VersionApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**version_get**](VersionApi.md#version_get) | **GET** /__version__ | 


# **version_get**
> Version version_get()



### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.VersionApi()

try: 
    api_response = api_instance.version_get()
    pprint(api_response)
except ApiException as e:
    print "Exception when calling VersionApi->version_get: %s\n" % e
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Version**](Version.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

