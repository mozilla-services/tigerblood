# swagger_client.ReputationApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ip_delete**](ReputationApi.md#ip_delete) | **DELETE** /{ip} | delete reputation for saved IP
[**ip_get**](ReputationApi.md#ip_get) | **GET** /{ip} | get IP reputation
[**ip_put**](ReputationApi.md#ip_put) | **PUT** /{ip} | update reputation for saved IP
[**root_post**](ReputationApi.md#root_post) | **POST** / | 


# **ip_delete**
> ip_delete(ip)

delete reputation for saved IP

### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.ReputationApi()
ip = 'ip_example' # str | IP of reputation to update

try: 
    # delete reputation for saved IP
    api_instance.ip_delete(ip)
except ApiException as e:
    print "Exception when calling ReputationApi->ip_delete: %s\n" % e
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **str**| IP of reputation to update | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ip_get**
> ip_get(ip)

get IP reputation

### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.ReputationApi()
ip = 'ip_example' # str | IP of reputation to return

try: 
    # get IP reputation
    api_instance.ip_get(ip)
except ApiException as e:
    print "Exception when calling ReputationApi->ip_get: %s\n" % e
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **str**| IP of reputation to return | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ip_put**
> ip_put(ip, reputation=reputation)

update reputation for saved IP

### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.ReputationApi()
ip = 'ip_example' # str | IP of reputation to update
reputation = swagger_client.Reputation1() # Reputation1 | reputation to set IP to (optional)

try: 
    # update reputation for saved IP
    api_instance.ip_put(ip, reputation=reputation)
except ApiException as e:
    print "Exception when calling ReputationApi->ip_put: %s\n" % e
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **str**| IP of reputation to update | 
 **reputation** | [**Reputation1**](Reputation1.md)| reputation to set IP to | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **root_post**
> root_post(reputation_and_ip_body=reputation_and_ip_body)



### Example 
```python
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.ReputationApi()
reputation_and_ip_body = swagger_client.Reputation() # Reputation | object with ip and reputation (optional)

try: 
    api_instance.root_post(reputation_and_ip_body=reputation_and_ip_body)
except ApiException as e:
    print "Exception when calling ReputationApi->root_post: %s\n" % e
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **reputation_and_ip_body** | [**Reputation**](Reputation.md)| object with ip and reputation | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

