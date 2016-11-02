# swagger_client.ViolationApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**violations_ip_put**](ViolationApi.md#violations_ip_put) | **PUT** /violations/{ip} | upsert reputation for an IP by violation


# **violations_ip_put**
> violations_ip_put(ip, violation)

upsert reputation for an IP by violation

### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.ViolationApi()
ip = 'ip_example' # str | IP of reputation to update
violation = 'violation_example' # str | violation type to report

try: 
    # upsert reputation for an IP by violation
    api_instance.violations_ip_put(ip, violation)
except ApiException as e:
    print "Exception when calling ViolationApi->violations_ip_put: %s\n" % e
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **str**| IP of reputation to update | 
 **violation** | **str**| violation type to report | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

