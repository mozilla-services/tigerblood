# coding: utf-8

"""
    Tigerblood

    IP Reputation Service API

    OpenAPI spec version: 1.0.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
"""

from __future__ import absolute_import

import sys
import os
import re

# python 2 and python 3 compatibility library
from six import iteritems

from ..configuration import Configuration
from ..api_client import ApiClient


class ReputationApi(object):
    """
    NOTE: This class is auto generated by the swagger code generator program.
    Do not edit the class manually.
    Ref: https://github.com/swagger-api/swagger-codegen
    """

    def __init__(self, api_client=None):
        config = Configuration()
        if api_client:
            self.api_client = api_client
        else:
            if not config.api_client:
                config.api_client = ApiClient()
            self.api_client = config.api_client

    def ip_delete(self, ip, **kwargs):
        """
        delete reputation for saved IP
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_delete(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to update (required)
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('callback'):
            return self.ip_delete_with_http_info(ip, **kwargs)
        else:
            (data) = self.ip_delete_with_http_info(ip, **kwargs)
            return data

    def ip_delete_with_http_info(self, ip, **kwargs):
        """
        delete reputation for saved IP
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_delete_with_http_info(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to update (required)
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['ip']
        all_params.append('callback')
        all_params.append('_return_http_data_only')

        params = locals()
        for key, val in iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method ip_delete" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'ip' is set
        if ('ip' not in params) or (params['ip'] is None):
            raise ValueError("Missing the required parameter `ip` when calling `ip_delete`")

        resource_path = '/{ip}'.replace('{format}', 'json')
        path_params = {}
        if 'ip' in params:
            path_params['ip'] = params['ip']

        query_params = {}

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None

        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.\
            select_header_accept(['application/json'])
        if not header_params['Accept']:
            del header_params['Accept']

        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.\
            select_header_content_type(['application/json'])

        # Authentication setting
        auth_settings = []

        return self.api_client.call_api(resource_path, 'DELETE',
                                            path_params,
                                            query_params,
                                            header_params,
                                            body=body_params,
                                            post_params=form_params,
                                            files=local_var_files,
                                            response_type=None,
                                            auth_settings=auth_settings,
                                            callback=params.get('callback'),
                                            _return_http_data_only=params.get('_return_http_data_only'))

    def ip_get(self, ip, **kwargs):
        """
        get IP reputation
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_get(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to return (required)
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('callback'):
            return self.ip_get_with_http_info(ip, **kwargs)
        else:
            (data) = self.ip_get_with_http_info(ip, **kwargs)
            return data

    def ip_get_with_http_info(self, ip, **kwargs):
        """
        get IP reputation
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_get_with_http_info(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to return (required)
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['ip']
        all_params.append('callback')
        all_params.append('_return_http_data_only')

        params = locals()
        for key, val in iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method ip_get" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'ip' is set
        if ('ip' not in params) or (params['ip'] is None):
            raise ValueError("Missing the required parameter `ip` when calling `ip_get`")

        resource_path = '/{ip}'.replace('{format}', 'json')
        path_params = {}
        if 'ip' in params:
            path_params['ip'] = params['ip']

        query_params = {}

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None

        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.\
            select_header_accept(['application/json'])
        if not header_params['Accept']:
            del header_params['Accept']

        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.\
            select_header_content_type(['application/json'])

        # Authentication setting
        auth_settings = []

        return self.api_client.call_api(resource_path, 'GET',
                                            path_params,
                                            query_params,
                                            header_params,
                                            body=body_params,
                                            post_params=form_params,
                                            files=local_var_files,
                                            response_type=None,
                                            auth_settings=auth_settings,
                                            callback=params.get('callback'),
                                            _return_http_data_only=params.get('_return_http_data_only'))

    def ip_put(self, ip, **kwargs):
        """
        update reputation for saved IP
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_put(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to update (required)
        :param Reputation1 reputation: reputation to set IP to
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('callback'):
            return self.ip_put_with_http_info(ip, **kwargs)
        else:
            (data) = self.ip_put_with_http_info(ip, **kwargs)
            return data

    def ip_put_with_http_info(self, ip, **kwargs):
        """
        update reputation for saved IP
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.ip_put_with_http_info(ip, callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param str ip: IP of reputation to update (required)
        :param Reputation1 reputation: reputation to set IP to
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['ip', 'reputation']
        all_params.append('callback')
        all_params.append('_return_http_data_only')

        params = locals()
        for key, val in iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method ip_put" % key
                )
            params[key] = val
        del params['kwargs']
        # verify the required parameter 'ip' is set
        if ('ip' not in params) or (params['ip'] is None):
            raise ValueError("Missing the required parameter `ip` when calling `ip_put`")

        resource_path = '/{ip}'.replace('{format}', 'json')
        path_params = {}
        if 'ip' in params:
            path_params['ip'] = params['ip']

        query_params = {}

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        if 'reputation' in params:
            body_params = params['reputation']

        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.\
            select_header_accept(['application/json'])
        if not header_params['Accept']:
            del header_params['Accept']

        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.\
            select_header_content_type(['application/json'])

        # Authentication setting
        auth_settings = []

        return self.api_client.call_api(resource_path, 'PUT',
                                            path_params,
                                            query_params,
                                            header_params,
                                            body=body_params,
                                            post_params=form_params,
                                            files=local_var_files,
                                            response_type=None,
                                            auth_settings=auth_settings,
                                            callback=params.get('callback'),
                                            _return_http_data_only=params.get('_return_http_data_only'))

    def root_post(self, **kwargs):
        """
        
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.root_post(callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param Reputation reputation_and_ip_body: object with ip and reputation
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """
        kwargs['_return_http_data_only'] = True
        if kwargs.get('callback'):
            return self.root_post_with_http_info(**kwargs)
        else:
            (data) = self.root_post_with_http_info(**kwargs)
            return data

    def root_post_with_http_info(self, **kwargs):
        """
        
        

        This method makes a synchronous HTTP request by default. To make an
        asynchronous HTTP request, please define a `callback` function
        to be invoked when receiving the response.
        >>> def callback_function(response):
        >>>     pprint(response)
        >>>
        >>> thread = api.root_post_with_http_info(callback=callback_function)

        :param callback function: The callback function
            for asynchronous request. (optional)
        :param Reputation reputation_and_ip_body: object with ip and reputation
        :return: None
                 If the method is called asynchronously,
                 returns the request thread.
        """

        all_params = ['reputation_and_ip_body']
        all_params.append('callback')
        all_params.append('_return_http_data_only')

        params = locals()
        for key, val in iteritems(params['kwargs']):
            if key not in all_params:
                raise TypeError(
                    "Got an unexpected keyword argument '%s'"
                    " to method root_post" % key
                )
            params[key] = val
        del params['kwargs']

        resource_path = '/'.replace('{format}', 'json')
        path_params = {}

        query_params = {}

        header_params = {}

        form_params = []
        local_var_files = {}

        body_params = None
        if 'reputation_and_ip_body' in params:
            body_params = params['reputation_and_ip_body']

        # HTTP header `Accept`
        header_params['Accept'] = self.api_client.\
            select_header_accept(['application/json'])
        if not header_params['Accept']:
            del header_params['Accept']

        # HTTP header `Content-Type`
        header_params['Content-Type'] = self.api_client.\
            select_header_content_type(['application/json'])

        # Authentication setting
        auth_settings = []

        return self.api_client.call_api(resource_path, 'POST',
                                            path_params,
                                            query_params,
                                            header_params,
                                            body=body_params,
                                            post_params=form_params,
                                            files=local_var_files,
                                            response_type=None,
                                            auth_settings=auth_settings,
                                            callback=params.get('callback'),
                                            _return_http_data_only=params.get('_return_http_data_only'))
