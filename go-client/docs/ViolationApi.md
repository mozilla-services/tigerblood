# \ViolationApi

All URIs are relative to *http://localhost/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ViolationsIpPut**](ViolationApi.md#ViolationsIpPut) | **Put** /violations/{ip} | upsert reputation for an IP by violation


# **ViolationsIpPut**
> ViolationsIpPut($ip, $violation)

upsert reputation for an IP by violation


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ip** | **string**| IP of reputation to update | 
 **violation** | **string**| violation type to report | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

