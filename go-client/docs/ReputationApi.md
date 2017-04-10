# \ReputationApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**IpDelete**](ReputationApi.md#IpDelete) | **Delete** /{ip} | delete reputation for saved IP
[**IpGet**](ReputationApi.md#IpGet) | **Get** /{ip} | get IP reputation
[**IpPut**](ReputationApi.md#IpPut) | **Put** /{ip} | update reputation for saved IP
[**RootPost**](ReputationApi.md#RootPost) | **Post** / | 


# **IpDelete**
> IpDelete($ip)

delete reputation for saved IP


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **string**| IP of reputation to update | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **IpGet**
> IpGet($ip)

get IP reputation


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **string**| IP of reputation to return | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **IpPut**
> IpPut($ip, $reputation)

update reputation for saved IP


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **string**| IP of reputation to update | 
 **reputation** | [**Reputation1**](Reputation1.md)| reputation to set IP to | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RootPost**
> RootPost($reputationAndIPBody)




### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **reputationAndIPBody** | [**Reputation**](Reputation.md)| object with ip and reputation | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

