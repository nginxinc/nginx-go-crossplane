package analyzer

import (
	"errors"
	"strings"
)

type statement struct {
	directive string
	args      [1]string
	line      int
}

var DIRECTIVES = map[string][]string{
	"absolute_redirect": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"accept_mutex": []string{
		"NGX_EVENT_CONF", "NGX_CONF_FLAG"},
	"accept_mutex_delay": []string{
		"NGX_EVENT_CONF", "NGX_CONF_TAKE1"},
	"access_log": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"add_after_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"add_before_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"add_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE23"},
	"add_trailer": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE23"},
	"addition_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"aio": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"aio_write": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"alias": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"allow": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ancient_browser": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"ancient_browser_value": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"auth_basic": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1"},
	"auth_basic_user_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1"},
	"auth_http": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"auth_http_header": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE2"},
	"auth_http_pass_client_cert": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG"},
	"auth_http_timeout": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"auth_request": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"auth_request_set": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"autoindex": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"autoindex_exact_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"autoindex_format": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"autoindex_localtime": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"break": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_NOARGS"},
	"charset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"charset_map": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE2"},
	"charset_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"chunked_transfer_encoding": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"client_body_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"client_body_in_file_only": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"client_body_in_single_buffer": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"client_body_temp_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234"},
	"client_body_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"client_header_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"client_header_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"client_max_body_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"connection_pool_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"create_full_put_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"daemon": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_FLAG"},
	"dav_access": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"dav_methods": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"debug_connection": []string{
		"NGX_EVENT_CONF", "NGX_CONF_TAKE1"},
	"debug_points": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"default_type": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"deny": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"directio": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"directio_alignment": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"disable_symlinks": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"empty_gif": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"env": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"error_log": []string{
		"NGX_MAIN_CONF", "NGX_CONF_1MORE", "NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"error_page": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_2MORE"},
	"etag": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"events": []string{
		"NGX_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS"},
	"expires": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE12"},
	"fastcgi_bind": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"fastcgi_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"fastcgi_busy_buffers_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_background_update": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_cache_bypass": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_cache_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_lock": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_cache_lock_age": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_lock_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_max_range_offset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_methods": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_cache_min_uses": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE"},
	"fastcgi_cache_revalidate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_cache_use_stale": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_cache_valid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_catch_stderr": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_connect_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_force_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_hide_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_ignore_client_abort": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_ignore_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_index": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_intercept_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_keep_conn": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_max_temp_file_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_next_upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_next_upstream_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_next_upstream_tries": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_no_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"fastcgi_param": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE23"},
	"fastcgi_pass": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_pass_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_pass_request_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_pass_request_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_read_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_request_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"fastcgi_send_lowat": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_split_path_info": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_store": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_store_access": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"fastcgi_temp_file_write_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_temp_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234"},
	"flv": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"geo": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE12"},
	"geoip_city": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE12"},
	"geoip_country": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE12"},
	"geoip_org": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE12"},
	"geoip_proxy": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"geoip_proxy_recursive": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_FLAG"},
	"google_perftools_profiles": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"gunzip": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"gunzip_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"gzip": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"gzip_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"gzip_comp_level": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"gzip_disable": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"gzip_http_version": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"gzip_min_length": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"gzip_proxied": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"gzip_static": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"gzip_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"gzip_vary": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"hash": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_UPS_CONF", "NGX_CONF_TAKE12"},
	"http": []string{
		"NGX_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS"},
	"http2_body_preread_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_chunk_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"http2_idle_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_max_concurrent_streams": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_max_field_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_max_header_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_max_requests": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"http2_recv_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"http2_recv_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"if": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_BLOCK", "NGX_CONF_1MORE"},
	"if_modified_since": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"ignore_invalid_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG"},
	"image_filter": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"image_filter_buffer": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"image_filter_interlace": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"image_filter_jpeg_quality": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"image_filter_sharpen": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"image_filter_transparency": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"image_filter_webp_quality": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"imap_auth": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"imap_capabilities": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"imap_client_buffer": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"include": []string{
		"NGX_ANY_CONF", "NGX_CONF_TAKE1"},
	"index": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"internal": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"ip_hash": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_NOARGS"},
	"keepalive": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE1"},
	"keepalive_disable": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"keepalive_requests": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"keepalive_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"large_client_header_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE2"},
	"least_conn": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_NOARGS,NGX_STREAM_UPS_CONF", "NGX_CONF_NOARGS"},
	"limit_conn": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE2"},
	"limit_conn_log_level": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"limit_conn_status": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"limit_conn_zone": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE2,NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE2"},
	"limit_except": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_BLOCK", "NGX_CONF_1MORE"},
	"limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"limit_rate_after": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"limit_req": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"limit_req_log_level": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"limit_req_status": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"limit_req_zone": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE3"},
	"lingering_close": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"lingering_time": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"lingering_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"listen": []string{
		"NGX_HTTP_SRV_CONF", "NGX_CONF_1MORE,NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"load_module": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"location": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE12"},
	"lock_file": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"log_format": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE,NGX_STREAM_MAIN_CONF", "NGX_CONF_2MORE"},
	"log_not_found": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"log_subrequest": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"mail": []string{
		"NGX_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS"},
	"map": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE2", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE2"},
	"map_hash_bucket_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1,NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE1"},
	"map_hash_max_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1,NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE1"},
	"master_process": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_FLAG"},
	"max_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_bind": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"memcached_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_connect_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_gzip_flag": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_next_upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"memcached_next_upstream_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_next_upstream_tries": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_pass": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"memcached_read_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"memcached_send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"merge_slashes": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG"},
	"min_delete_depth": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"mirror": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"mirror_request_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"modern_browser": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"modern_browser_value": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"mp4": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"mp4_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"mp4_max_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"msie_padding": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"msie_refresh": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"multi_accept": []string{
		"NGX_EVENT_CONF", "NGX_CONF_FLAG"},
	"open_file_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"open_file_cache_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"open_file_cache_min_uses": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"open_file_cache_valid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"open_log_file_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1234"},
	"output_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"override_charset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"pcre_jit": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_FLAG"},
	"perl": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1"},
	"perl_modules": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"perl_require": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"perl_set": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE2"},
	"pid": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"pop3_auth": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"pop3_capabilities": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"port_in_redirect": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"postpone_output": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"preread_buffer_size": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"preread_timeout": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"protocol": []string{
		"NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_bind": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE12"},
	"proxy_buffer": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"proxy_busy_buffers_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_background_update": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_cache_bypass": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_cache_convert_head": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_cache_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_lock": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_cache_lock_age": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_lock_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_max_range_offset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_methods": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_cache_min_uses": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_cache_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE"},
	"proxy_cache_revalidate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_cache_use_stale": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_cache_valid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_connect_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_cookie_domain": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"proxy_cookie_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"proxy_download_rate": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_force_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_headers_hash_bucket_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_headers_hash_max_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_hide_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_http_version": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_ignore_client_abort": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_ignore_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_intercept_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_max_temp_file_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_method": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_next_upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_next_upstream_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_next_upstream_tries": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_no_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"proxy_pass": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_pass_error_message": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_pass_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_pass_request_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_pass_request_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_protocol": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_protocol_timeout": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_read_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_redirect": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"proxy_request_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"proxy_responses": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_send_lowat": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_set_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_set_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"proxy_ssl": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_ssl_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_certificate_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_ciphers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_crl": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_name": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_password_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_protocols": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"proxy_ssl_server_name": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_ssl_session_reuse": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_ssl_trusted_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_ssl_verify": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"proxy_ssl_verify_depth": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_store": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_store_access": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"proxy_temp_file_write_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"proxy_temp_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234"},
	"proxy_timeout": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"proxy_upload_rate": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"random_index": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"read_ahead": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"real_ip_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"real_ip_recursive": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"recursive_error_pages": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"referer_hash_bucket_size": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"referer_hash_max_size": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"request_pool_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"reset_timedout_connection": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"resolver": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"resolver_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"return": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"rewrite": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE23"},
	"rewrite_log": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"root": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"satisfy": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_bind": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"scgi_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"scgi_busy_buffers_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_background_update": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_cache_bypass": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_cache_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_lock": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_cache_lock_age": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_lock_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_max_range_offset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_methods": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_cache_min_uses": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_cache_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE"},
	"scgi_cache_revalidate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_cache_use_stale": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_cache_valid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_connect_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_force_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_hide_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_ignore_client_abort": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_ignore_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_intercept_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_max_temp_file_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_next_upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_next_upstream_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_next_upstream_tries": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_no_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"scgi_param": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE23"},
	"scgi_pass": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"scgi_pass_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_pass_request_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_pass_request_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_read_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_request_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"scgi_send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_store": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_store_access": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"scgi_temp_file_write_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"scgi_temp_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234"},
	"secure_link": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"secure_link_md5": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"secure_link_secret": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"send_lowat": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"sendfile": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"sendfile_max_chunk": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"server": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS", "NGX_HTTP_UPS_CONF", "NGX_CONF_1MORE", "NGX_MAIL_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS", "NGX_STREAM_UPS_CONF", "NGX_CONF_1MORE"},
	"server_name": []string{
		"NGX_HTTP_SRV_CONF", "NGX_CONF_1MORE,NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"server_name_in_redirect": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"server_names_hash_bucket_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"server_names_hash_max_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1"},
	"server_tokens": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"set": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE2"},
	"set_real_ip_from": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"slice": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"smtp_auth": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"smtp_capabilities": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE"},
	"source_charset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"spdy_chunk_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"spdy_headers_comp": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"split_clients": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE2", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE2"},
	"ssi": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"ssi_last_modified": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"ssi_min_file_chunk": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"ssi_silent_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"ssi_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"ssi_value_length": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"ssl": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_certificate_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_ciphers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_client_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_crl": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_dhparam": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_ecdh_curve": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_engine": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"ssl_handshake_timeout": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_password_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_prefer_server_ciphers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_preread": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_protocols": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_1MORE", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_1MORE", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_1MORE"},
	"ssl_session_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE12", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE12"},
	"ssl_session_ticket_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_session_tickets": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_session_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_stapling": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_stapling_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_stapling_responder": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_stapling_verify": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG"},
	"ssl_trusted_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_verify_client": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"ssl_verify_depth": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1", "NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"starttls": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"stream": []string{
		"NGX_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS"},
	"stub_status": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS", "NGX_CONF_TAKE1"},
	"sub_filter": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"sub_filter_last_modified": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"sub_filter_once": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"sub_filter_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"tcp_nodelay": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG", "NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_FLAG"},
	"tcp_nopush": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"thread_pool": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE23"},
	"timeout": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_TAKE1"},
	"timer_resolution": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"try_files": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_2MORE"},
	"types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_BLOCK", "NGX_CONF_NOARGS"},
	"types_hash_bucket_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"types_hash_max_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"underscores_in_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_CONF_FLAG"},
	"uninitialized_variable_warn": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_SIF_CONF", "NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_FLAG"},
	"upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE1"},
	"use": []string{
		"NGX_EVENT_CONF", "NGX_CONF_TAKE1"},
	"user": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE12"},
	"userid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_domain": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_expires": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_mark": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_name": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_p3p": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"userid_service": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_bind": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"uwsgi_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"uwsgi_busy_buffers_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_background_update": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_bypass": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_cache_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_lock": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_cache_lock_age": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_lock_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_max_range_offset": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_methods": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_cache_min_uses": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_cache_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE"},
	"uwsgi_cache_revalidate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_cache_use_stale": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_cache_valid": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_connect_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_force_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_hide_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ignore_client_abort": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_ignore_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_intercept_errors": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_max_temp_file_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_modifier1": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_modifier2": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_next_upstream": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_next_upstream_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_next_upstream_tries": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_no_cache": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_param": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE23"},
	"uwsgi_pass": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LIF_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_pass_header": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_pass_request_body": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_pass_request_headers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_read_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_request_buffering": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_send_timeout": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_certificate_key": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_ciphers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_crl": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_name": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_password_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_protocols": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"uwsgi_ssl_server_name": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_ssl_session_reuse": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_ssl_trusted_certificate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_ssl_verify": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"uwsgi_ssl_verify_depth": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_store": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_store_access": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE123"},
	"uwsgi_temp_file_write_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"uwsgi_temp_path": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1234"},
	"valid_referers": []string{
		"NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"variables_hash_bucket_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE1"},
	"variables_hash_max_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE1"},
	"worker_aio_requests": []string{
		"NGX_EVENT_CONF", "NGX_CONF_TAKE1"},
	"worker_connections": []string{
		"NGX_EVENT_CONF", "NGX_CONF_TAKE1"},
	"worker_cpu_affinity": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_1MORE"},
	"worker_priority": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"worker_processes": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"worker_rlimit_core": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"worker_rlimit_nofile": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"worker_shutdown_timeout": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"working_directory": []string{
		"NGX_MAIN_CONF", "NGX_DIRECT_CONF", "NGX_CONF_TAKE1"},
	"xclient": []string{
		"NGX_MAIL_MAIN_CONF", "NGX_MAIL_SRV_CONF", "NGX_CONF_FLAG"},
	"xml_entities": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"xslt_last_modified": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"xslt_param": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"xslt_string_param": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"xslt_stylesheet": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"xslt_types": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"zone": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_UPS_CONF", "NGX_CONF_TAKE12"},
	"auth_jwt": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"auth_jwt_claim_set": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE2"},
	"auth_jwt_header_set": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE2"},
	"auth_jwt_key_file": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"f4f": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"f4f_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"fastcgi_cache_purge": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"health_check": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_ANY", "NGX_STREAM_SRV_CONF", "NGX_CONF_ANY"},
	"health_check_timeout": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"hls": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"hls_buffers": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE2"},
	"hls_forward_args": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"hls_fragment": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"hls_mp4_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"hls_mp4_max_buffer_size": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"js_access": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"js_content": []string{
		"NGX_HTTP_LOC_CONF", "NGX_HTTP_LMT_CONF", "NGX_CONF_TAKE1"},
	"js_filter": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"js_include": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE1"},
	"js_preread": []string{
		"NGX_STREAM_MAIN_CONF", "NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"js_set": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE2", "NGX_STREAM_MAIN_CONF", "NGX_CONF_TAKE2"},
	"least_time": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE12", "NGX_STREAM_UPS_CONF", "NGX_CONF_TAKE12"},
	"limit_zone": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE3"},
	"match": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE1", "NGX_STREAM_MAIN_CONF", "NGX_CONF_BLOCK", "NGX_CONF_TAKE1"},
	"memcached_force_ranges": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_FLAG"},
	"mp4_limit_rate": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"mp4_limit_rate_after": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"ntlm": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_NOARGS"},
	"proxy_cache_purge": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"queue": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE12"},
	"scgi_cache_purge": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
	"session_log": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE1"},
	"session_log_format": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_2MORE"},
	"session_log_zone": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_CONF_TAKE23", "NGX_CONF_TAKE4", "NGX_CONF_TAKE5", "NGX_CONF_TAKE6"},
	"state": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE1", "NGX_STREAM_UPS_CONF", "NGX_CONF_TAKE1"},
	"status": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"status_format": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_TAKE12"},
	"status_zone": []string{
		"NGX_HTTP_SRV_CONF", "NGX_CONF_TAKE1,NGX_STREAM_SRV_CONF", "NGX_CONF_TAKE1"},
	"sticky": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_1MORE"},
	"sticky_cookie_insert": []string{
		"NGX_HTTP_UPS_CONF", "NGX_CONF_TAKE1234"},
	"upstream_conf": []string{
		"NGX_HTTP_LOC_CONF", "NGX_CONF_NOARGS"},
	"uwsgi_cache_purge": []string{
		"NGX_HTTP_MAIN_CONF", "NGX_HTTP_SRV_CONF", "NGX_HTTP_LOC_CONF", "NGX_CONF_1MORE"},
}
var MASKS = map[string]uint{
	// bit masks for different directive locations
	"NGX_DIRECT_CONF":      0x00010000, // main file (not used)
	"NGX_MAIN_CONF":        0x00040000, // main context
	"NGX_EVENT_CONF":       0x00080000, // events
	"NGX_MAIL_MAIN_CONF":   0x00100000, // mail
	"NGX_MAIL_SRV_CONF":    0x00200000, // mail > server
	"NGX_STREAM_MAIN_CONF": 0x00400000, // stream
	"NGX_STREAM_SRV_CONF":  0x00800000, // stream > server
	"NGX_STREAM_UPS_CONF":  0x01000000, // stream > upstream
	"NGX_HTTP_MAIN_CONF":   0x02000000, // http
	"NGX_HTTP_SRV_CONF":    0x04000000, // http > server
	"NGX_HTTP_LOC_CONF":    0x08000000, // http > location
	"NGX_HTTP_UPS_CONF":    0x10000000, // http > upstream
	"NGX_HTTP_SIF_CONF":    0x20000000, // http > server > if
	"NGX_HTTP_LIF_CONF":    0x40000000, // http > location > if
	"NGX_HTTP_LMT_CONF":    0x80000000,

	// bit masks for different directive argument styles
	"NGX_CONF_NOARGS": 0x00000001, // 0 args
	"NGX_CONF_TAKE1":  0x00000002, // 1 args
	"NGX_CONF_TAKE2":  0x00000004, // 2 args
	"NGX_CONF_TAKE3":  0x00000008, // 3 args
	"NGX_CONF_TAKE4":  0x00000010, // 4 args
	"NGX_CONF_TAKE5":  0x00000020, // 5 args
	"NGX_CONF_TAKE6":  0x00000040, // 6 args
	"NGX_CONF_TAKE7":  0x00000080, // 7 args
	"NGX_CONF_BLOCK":  0x00000100, // followed by block
	"NGX_CONF_FLAG":   0x00000200, // 'on' or 'off'
	"NGX_CONF_ANY":    0x00000400, // >=0 args
	"NGX_CONF_1MORE":  0x00000800, // >=1 args
	"NGX_CONF_2MORE":  0x00001000, // >=2 args

}
var CONTEXT = map[[3]string]string{
	[3]string{}:                                   "NGX_MAIN_CONF",
	[3]string{"events"}:                           "NGX_EVENT_CONF",
	[3]string{"mail"}:                             "NGX_MAIL_MAIN_CONF",
	[3]string{"mail", "server"}:                   "NGX_MAIL_SRV_CONF",
	[3]string{"stream"}:                           "NGX_STREAM_MAIN_CONF",
	[3]string{"stream", "server"}:                 "NGX_STREAM_SRV_CONF",
	[3]string{"stream", "upstream"}:               "NGX_STREAM_UPS_CONF",
	[3]string{"http"}:                             "NGX_HTTP_MAIN_CONF",
	[3]string{"http", "server"}:                   "NGX_HTTP_SRV_CONF",
	[3]string{"http", "location"}:                 "NGX_HTTP_LOC_CONF",
	[3]string{"http", "upstream"}:                 "NGX_HTTP_UPS_CONF",
	[3]string{"http", "server", "if"}:             "NGX_HTTP_SIF_CONF",
	[3]string{"http", "location", "if"}:           "NGX_HTTP_LIF_CONF",
	[3]string{"http", "location", "limit_except"}: "NGX_HTTP_LMT_CONF",
}

func analyze(fname string, stmt statement, term string, ctx [3]string, strict bool, checkCtx bool, checkArg bool) error {

	directive := stmt.directive
	dir := checkDirective(directive, DIRECTIVES)

	// if strict and directive isn't recognized then throw error
	if strict && !dir {
		return errors.New("unknown directive " + directive)
	}

	ct := checkContext(ctx, CONTEXT)
	// if we don't know where this directive is allowed and how
	// many arguments it can take then don't bother analyzing it
	if !ct || !dir {
		return errors.New("problem here")
	}

	args := stmt.args
	//  makes numArgs an unsigned int for bit shifting later
	numArgs := uint(len(args))

	masks := DIRECTIVES[directive]
	// if this directive can't be used in this context then throw an error
	if checkCtx {
		for _, mask := range masks {
			bitmask := CONTEXT[ctx]
			if MASKS[mask]&MASKS[bitmask] != 0x00000000 {
				masks = append(masks, mask)
			}
		}

		if len(masks) == 0 {
			return errors.New(directive + " directive is not allowed here")
		}
	}

	if !checkArg {
		return nil
	}

	validFlags := func(x string) bool {
		x = strings.ToLower(x)
		for _, v := range [2]string{"on", "off"} {
			if x == v {
				return true
			}
		}
		return false
	}
	// do this in reverse because we only throw errors at the end if no masks
	// are valid, and typically the first bit mask is what the parser expects
	reason := ""
	for i := len(masks) - 1; i >= 0; i-- {
		msk := masks[i]
		// if the directive isn't a block but should be according to the mask
		if MASKS[msk]&MASKS["NGX_CONF_BLOCK"] != 0x00000000 && term != "{" {
			reason = "directive " + directive + " has no opening '{'"
			continue
		}
		//if the directive is a block but shouldn't be according to the mask
		if MASKS[msk]&MASKS["NGX_CONF_BLOCK"] != 0x00000000 && term != ";" {
			reason = "directive " + directive + " is not terminated by ';'"
			continue
		}
		// use mask to check the directive's arguments
		if ((MASKS[msk]>>numArgs)&1 != 0x00000000 && numArgs <= 7) || //NOARGS to TAKE7
			(MASKS[msk]&MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs == 1 && validFlags(stmt.args[0])) ||
			(MASKS[msk]&MASKS["NGX_CONF_ANY"] != 0x00000000 && numArgs >= 0) ||
			(MASKS[msk]&MASKS["NGX_CONF_1MORE"] != 0x00000000 && numArgs >= 1) ||
			(MASKS[msk]&MASKS["NGX_CONF_2MORE"] != 0x00000000 && numArgs >= 2) {
			return nil
		} else if MASKS[msk]&MASKS["NGX_CONF_FLAG"] != 0x00000000 && numArgs == 1 && !validFlags(stmt.args[0]) {
			reason = "invalid value " + stmt.args[0] + " in " + stmt.directive + " directive, it must be 'on' or 'off'"
			continue
		} else {
			reason = "invalid number of arguements in " + directive

		}
	}
	if reason == "" {
		return nil
	}
	return errors.New(reason)
}

func checkContext(cont [3]string, contexts map[[3]string]string) bool {
	if _, ok := contexts[cont]; ok {
		return true
	}
	return false
}

func checkDirective(dir string, direct map[string][]string) bool {
	for d := range direct {
		if d == dir {
			return true
		}
	}
	return false
}

func enterBlockCTX(stmt statement, ctx [3]string) [3]string {
	if len(ctx) != 0 && ctx[0] == "http" && stmt.directive == "location" {
		return [3]string{"http", "location"}
	}
	for i, v := range ctx {
		if v == "" {
			ctx[i] = stmt.directive
			break
		}
	}
	return ctx
}

func registerExternalDirectives(directives map[string][]string) {
	for d, b := range directives {
		DIRECTIVES[d] = []string{}
		for _, v := range b {
			if MASKS[v] != 0x00000000 {
				DIRECTIVES[d] = append(DIRECTIVES[d], v)
			}
		}
	}
}
