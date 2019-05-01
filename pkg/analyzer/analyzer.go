package analyzer

import (
	"fmt"
	"strings"
)

type statement struct {
	directive string
	args      [1]string
	line      int
}
type Bits uint

const (
	NGX_DIRECT_CONF      Bits = 0x00010000 // main file (not used)
	NGX_MAIN_CONF        Bits = 0x00040000 // main context
	NGX_EVENT_CONF       Bits = 0x00080000 // events
	NGX_MAIL_MAIN_CONF   Bits = 0x00100000 // mail
	NGX_MAIL_SRV_CONF    Bits = 0x00200000 // mail > server
	NGX_STREAM_MAIN_CONF Bits = 0x00400000 // stream
	NGX_STREAM_SRV_CONF  Bits = 0x00800000 // stream > server
	NGX_STREAM_UPS_CONF  Bits = 0x01000000 // stream > upstream
	NGX_HTTP_MAIN_CONF   Bits = 0x02000000 // http
	NGX_HTTP_SRV_CONF    Bits = 0x04000000 // http > server
	NGX_HTTP_LOC_CONF    Bits = 0x08000000 // http > location
	NGX_HTTP_UPS_CONF    Bits = 0x10000000 // http > upstream
	NGX_HTTP_SIF_CONF    Bits = 0x20000000 // http > server > if
	NGX_HTTP_LIF_CONF    Bits = 0x40000000 // http > location > if
	NGX_HTTP_LMT_CONF    Bits = 0x80000000
	NGX_CONF_TAKE12      Bits = (NGX_CONF_TAKE1 | NGX_CONF_TAKE2)
	NGX_CONF_TAKE13      Bits = (NGX_CONF_TAKE1 | NGX_CONF_TAKE3)
	NGX_CONF_TAKE23      Bits = (NGX_CONF_TAKE2 | NGX_CONF_TAKE3)
	NGX_CONF_TAKE123     Bits = (NGX_CONF_TAKE12 | NGX_CONF_TAKE3)
	NGX_CONF_TAKE1234    Bits = (NGX_CONF_TAKE123 | NGX_CONF_TAKE4)

	// bit masks for different directive argument styles
	NGX_CONF_NOARGS Bits = 0x00000001 // 0 args
	NGX_CONF_TAKE1  Bits = 0x00000002 // 1 args
	NGX_CONF_TAKE2  Bits = 0x00000004 // 2 args
	NGX_CONF_TAKE3  Bits = 0x00000008 // 3 args
	NGX_CONF_TAKE4  Bits = 0x00000010 // 4 args
	NGX_CONF_TAKE5  Bits = 0x00000020 // 5 args
	NGX_CONF_TAKE6  Bits = 0x00000040 // 6 args
	NGX_CONF_TAKE7  Bits = 0x00000080 // 7 args
	NGX_CONF_BLOCK  Bits = 0x00000100 // followed by block
	NGX_CONF_FLAG   Bits = 0x00000200 // 'on' or 'off'
	NGX_CONF_ANY    Bits = 0x00000400 // >=0 args
	NGX_CONF_1MORE  Bits = 0x00000800 // >=1 args
	NGX_CONF_2MORE  Bits = 0x00001000 // >=2 args

	NGX_ANY_CONF Bits = (NGX_MAIN_CONF | NGX_EVENT_CONF | NGX_MAIL_MAIN_CONF | NGX_MAIL_SRV_CONF |
		NGX_STREAM_MAIN_CONF | NGX_STREAM_SRV_CONF | NGX_STREAM_UPS_CONF |
		NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_HTTP_UPS_CONF)
)

// Directives -
// This dict maps directives to lists of bit masks that define their behavior.
//Each bit mask describes these behaviors:
//  - how many arguments the directive can take
//  - whether or not it is a block directive
//  - whether this is a flag (takes one argument that's either "on" or "off")
//  - which contexts it's allowed to be in
// Since some directives can have different behaviors in different contexts, we
//  use lists of bit masks, each describing a valid way to use the directive.
//Definitions for directives that're available in the open source version of
//  nginx were taken directively from the source code. In fact, the variable
//  names for the bit masks defined above were taken from the nginx source code.
//Definitions for directives that're only available for nginx+ were inferred
//  from the documentation at http://nginx.org/en/docs/.
var Directives = map[string][]Bits{
	"absolute_redirect": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"accept_mutex": {
		ngxEventConf, ngxConfFlag,
	},
	"accept_mutex_delay": {
		ngxEventConf, ngxConfTake1,
	},
	"access_log": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxHttpLmtConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"add_after_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"add_before_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"add_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake23,
	},
	"add_trailer": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake23,
	},
	"addition_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"aio": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"aio_write": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"alias": {
		ngxHttpLocConf, ngxConfTake1,
	},
	"allow": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ancient_browser": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"ancient_browser_value": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"auth_basic": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
	},
	"auth_basic_user_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
	},
	"auth_http": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"auth_http_header": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake2,
	},
	"auth_http_pass_client_cert": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"auth_http_timeout": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"auth_request": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"auth_request_set": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"autoindex": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"autoindex_exact_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"autoindex_format": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"autoindex_localtime": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"break": {
		ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfNoArgs,
	},
	"charset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"charset_map": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake2,
	},
	"charset_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"chunked_transfer_encoding": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"client_body_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"client_body_in_file_only": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"client_body_in_single_buffer": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"client_body_temp_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
	},
	"client_body_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"client_header_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"client_header_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"client_max_body_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"connection_pool_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"create_full_put_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"daemon": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"dav_access": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"dav_methods": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"debug_connection": {
		ngxEventConf, ngxConfTake1,
	},
	"debug_points": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"default_type": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"deny": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"directio": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"directio_alignment": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"disable_symlinks": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"empty_gif": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"env": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"error_log": {
		ngxMainConf, ngxConf1More,
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"error_page": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConf2More,
	},
	"etag": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"events": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"expires": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake12,
	},
	"fastcgi_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"fastcgi_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"fastcgi_busy_buffers_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_background_update": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_cache_bypass": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_cache_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_lock": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_cache_lock_age": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_lock_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_max_range_offset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_methods": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_cache_min_uses": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_path": {
		ngxHttpMainConf, ngxConf2More,
	},
	"fastcgi_cache_revalidate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_cache_use_stale": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_cache_valid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_catch_stderr": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_force_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_hide_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_ignore_client_abort": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_ignore_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_index": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_intercept_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_keep_conn": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_max_temp_file_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_no_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"fastcgi_param": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake23,
	},
	"fastcgi_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"fastcgi_pass_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_pass_request_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_pass_request_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_request_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"fastcgi_send_lowat": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_split_path_info": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_store": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_store_access": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"fastcgi_temp_file_write_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_temp_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
	},
	"fastcgi_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"flv": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"geo": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake12,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake12,
	},
	"geoip_city": {
		ngxHttpMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_country": {
		ngxHttpMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_org": {
		ngxHttpMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_proxy": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"geoip_proxy_recursive": {
		ngxHttpMainConf, ngxConfFlag,
	},
	"google_perftools_profiles": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},

	"grpc_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"grpc_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_hide_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ignore_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"grpc_intercept_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"grpc_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"grpc_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"grpc_pass_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_set_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"grpc_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"grpc_ssl_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_certificate_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_ciphers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_crl": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_password_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_protocols": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"grpc_ssl_server_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"grpc_ssl_session_reuse": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"grpc_ssl_trusted_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"grpc_ssl_verify": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"grpc_ssl_verify_depth": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},

	"gunzip": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"gunzip_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"gzip": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"gzip_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"gzip_comp_level": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"gzip_disable": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"gzip_http_version": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"gzip_min_length": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"gzip_proxied": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"gzip_static": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"gzip_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"gzip_vary": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"hash": {
		ngxHttpUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},
	"http": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"http2_body_preread_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_chunk_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"http2_idle_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_max_concurrent_streams": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_max_field_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_max_header_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_max_requests": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_recv_buffer_size": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"http2_recv_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"http2_push": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"http2_push_preload": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"http2_max_concurrent_pushes": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"if": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConfBlock, ngxConf1More,
	},
	"if_modified_since": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"ignore_invalid_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"image_filter": {
		ngxHttpLocConf, ngxConfTake123,
	},
	"image_filter_buffer": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"image_filter_interlace": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"image_filter_jpeg_quality": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"image_filter_sharpen": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"image_filter_transparency": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"image_filter_webp_quality": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"imap_auth": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"imap_capabilities": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"imap_client_buffer": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"include": {
		NGX_ANY_CONF, ngxConfTake1,
	},
	"index": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"internal": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"ip_hash": {
		ngxHttpUpsConf, ngxConfNoArgs,
	},
	"keepalive": {
		ngxHttpUpsConf, ngxConfTake1,
	},
	"keepalive_disable": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"keepalive_requests": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxHttpUpsConf, ngxConfTake1,
	},
	"keepalive_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
		ngxHttpUpsConf, ngxConfTake1,
	},
	"large_client_header_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake2,
	},
	"least_conn": {
		ngxHttpUpsConf, ngxConfNoArgs,
		ngxStreamUpsConf, ngxConfNoArgs,
	},
	"limit_conn": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake2,
	},
	"limit_conn_log_level": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"limit_conn_status": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"limit_conn_zone": {
		ngxHttpMainConf, ngxConfTake2,
		ngxStreamMainConf, ngxConfTake2,
	},
	"limit_except": {
		ngxHttpLocConf, ngxConfBlock, ngxConf1More,
	},
	"limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"limit_rate_after": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"limit_req": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"limit_req_log_level": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"limit_req_status": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"limit_req_zone": {
		ngxHttpMainConf, ngxConfTake3,
	},
	"lingering_close": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"lingering_time": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"lingering_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"listen": {
		ngxHttpSrvConf, ngxConf1More,
		ngxMailSrvConf, ngxConf1More,
		ngxStreamSrvConf, ngxConf1More,
	},
	"load_module": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"location": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConfBlock, ngxConfTake12,
	},
	"lock_file": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"log_format": {
		ngxHttpMainConf, ngxConf2More,
		ngxStreamMainConf, ngxConf2More,
	},
	"log_not_found": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"log_subrequest": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"mail": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"map": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake2,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake2,
	},
	"map_hash_bucket_size": {
		ngxHttpMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"map_hash_max_size": {
		ngxHttpMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"master_process": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"max_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"memcached_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_gzip_flag": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"memcached_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"memcached_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"memcached_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"merge_slashes": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"min_delete_depth": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"mirror": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"mirror_request_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"modern_browser": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"modern_browser_value": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"mp4": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"mp4_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"mp4_max_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"msie_padding": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"msie_refresh": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"multi_accept": {
		ngxEventConf, ngxConfFlag,
	},
	"open_file_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"open_file_cache_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"open_file_cache_min_uses": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"open_file_cache_valid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"open_log_file_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1234,
	},
	"output_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"override_charset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"pcre_jit": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"perl": {
		ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
	},
	"perl_modules": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"perl_require": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"perl_set": {
		ngxHttpMainConf, ngxConfTake2,
	},
	"pid": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"pop3_auth": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"pop3_capabilities": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"port_in_redirect": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"postpone_output": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"preread_buffer_size": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"preread_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"protocol": {
		ngxMailSrvConf, ngxConfTake1,
	},
	"proxy_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake12,
	},
	"proxy_buffer": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"proxy_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"proxy_busy_buffers_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_background_update": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_cache_bypass": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_cache_convert_head": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_cache_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_lock": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_cache_lock_age": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_lock_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_max_range_offset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_methods": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_cache_min_uses": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_cache_path": {
		ngxHttpMainConf, ngxConf2More,
	},
	"proxy_cache_revalidate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_cache_use_stale": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_cache_valid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_cookie_domain": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"proxy_cookie_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"proxy_download_rate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_force_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_headers_hash_bucket_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_headers_hash_max_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_hide_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_http_version": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_ignore_client_abort": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_ignore_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_intercept_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_max_temp_file_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_method": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_no_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxHttpLmtConf, ngxConfTake1,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_pass_error_message": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"proxy_pass_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_pass_request_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_pass_request_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_protocol": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_protocol_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_redirect": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"proxy_request_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"proxy_responses": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_send_lowat": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_set_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_set_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"proxy_ssl": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_certificate_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_ciphers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_crl": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_password_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_protocols": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"proxy_ssl_server_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_session_reuse": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_trusted_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_verify": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_verify_depth": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_store": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_store_access": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"proxy_temp_file_write_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"proxy_temp_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
	},
	"proxy_timeout": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_upload_rate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"random": {
		ngxHttpUpsConf, ngxConfNoArgs, ngxConfTake12,
		ngxStreamUpsConf, ngxConfNoArgs, ngxConfTake12,
	},
	"random_index": {
		ngxHttpLocConf, ngxConfFlag,
	},
	"read_ahead": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"real_ip_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"real_ip_recursive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"recursive_error_pages": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"referer_hash_bucket_size": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"referer_hash_max_size": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"request_pool_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"reset_timedout_connection": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"resolver": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"resolver_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"return": {
		ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake12,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"rewrite": {
		ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake23,
	},
	"rewrite_log": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"root": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"satisfy": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"scgi_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"scgi_busy_buffers_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_background_update": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_cache_bypass": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_cache_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_lock": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_cache_lock_age": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_lock_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_max_range_offset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_methods": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_cache_min_uses": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_cache_path": {
		ngxHttpMainConf, ngxConf2More,
	},
	"scgi_cache_revalidate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_cache_use_stale": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_cache_valid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_force_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_hide_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_ignore_client_abort": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_ignore_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_intercept_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_max_temp_file_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_no_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_param": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake23,
	},
	"scgi_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"scgi_pass_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_pass_request_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_pass_request_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_request_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"scgi_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_store": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_store_access": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"scgi_temp_file_write_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"scgi_temp_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
	},
	"secure_link": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"secure_link_md5": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"secure_link_secret": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"send_lowat": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"sendfile": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"sendfile_max_chunk": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"server": {
		ngxHttpMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxHttpUpsConf, ngxConf1More,
		ngxMailMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxStreamMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxStreamUpsConf, ngxConf1More,
	},
	"server_name": {
		ngxHttpSrvConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"server_name_in_redirect": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"server_names_hash_bucket_size": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"server_names_hash_max_size": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"server_tokens": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"set": {
		ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake2,
	},
	"set_real_ip_from": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"slice": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"smtp_auth": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"smtp_capabilities": {
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
	},
	"smtp_client_buffer": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"smtp_greeting_delay": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"source_charset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"spdy_chunk_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"spdy_headers_comp": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"split_clients": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake2,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake2,
	},
	"ssi": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"ssi_last_modified": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"ssi_min_file_chunk": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"ssi_silent_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"ssi_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"ssi_value_length": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"ssl": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"ssl_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"ssl_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_certificate_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_ciphers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_client_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_crl": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_dhparam": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_early_data": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"ssl_ecdh_curve": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_engine": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"ssl_handshake_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_password_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_prefer_server_ciphers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_preread": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_protocols": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"ssl_session_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake12,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake12,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake12,
	},
	"ssl_session_ticket_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_session_tickets": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_session_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_stapling": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"ssl_stapling_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"ssl_stapling_responder": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
	},
	"ssl_stapling_verify": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"ssl_trusted_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_verify_client": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_verify_depth": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"starttls": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"stream": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"stub_status": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConfNoArgs, ngxConfTake1,
	},
	"sub_filter": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"sub_filter_last_modified": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"sub_filter_once": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"sub_filter_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"subrequest_output_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"tcp_nodelay": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"tcp_nopush": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"thread_pool": {
		ngxMainConf, ngxDirectConf, ngxConfTake23,
	},
	"timeout": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"timer_resolution": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"try_files": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConf2More,
	},
	"types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfBlock, ngxConfNoArgs,
	},
	"types_hash_bucket_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"types_hash_max_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"underscores_in_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxConfFlag,
	},
	"uninitialized_variable_warn": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpSifConf, ngxHttpLocConf, ngxHttpLifConf, ngxConfFlag,
	},
	"upstream": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake1,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake1,
	},
	"use": {
		ngxEventConf, ngxConfTake1,
	},
	"user": {
		ngxMainConf, ngxDirectConf, ngxConfTake12,
	},
	"userid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_domain": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_expires": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_mark": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_p3p": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"userid_service": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_bind": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"uwsgi_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"uwsgi_busy_buffers_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_background_update": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_bypass": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_cache_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_lock": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_cache_lock_age": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_lock_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_max_range_offset": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_methods": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_cache_min_uses": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_cache_path": {
		ngxHttpMainConf, ngxConf2More,
	},
	"uwsgi_cache_revalidate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_cache_use_stale": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_cache_valid": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_connect_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_force_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_hide_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ignore_client_abort": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_ignore_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_intercept_errors": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_max_temp_file_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_modifier1": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_modifier2": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_next_upstream": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_next_upstream_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_next_upstream_tries": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_no_cache": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_param": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake23,
	},
	"uwsgi_pass": {
		ngxHttpLocConf, ngxHttpLifConf, ngxConfTake1,
	},
	"uwsgi_pass_header": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_pass_request_body": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_pass_request_headers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_read_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_request_buffering": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_send_timeout": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_certificate_key": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_ciphers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_crl": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_password_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_protocols": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"uwsgi_ssl_server_name": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_session_reuse": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_trusted_certificate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_verify": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_verify_depth": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_store": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_store_access": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake123,
	},
	"uwsgi_temp_file_write_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"uwsgi_temp_path": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1234,
	},
	"valid_referers": {
		ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"variables_hash_bucket_size": {
		ngxHttpMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"variables_hash_max_size": {
		ngxHttpMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"worker_aio_requests": {
		ngxEventConf, ngxConfTake1,
	},
	"worker_connections": {
		ngxEventConf, ngxConfTake1,
	},
	"worker_cpu_affinity": {
		ngxMainConf, ngxDirectConf, ngxConf1More,
	},
	"worker_priority": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"worker_processes": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"worker_rlimit_core": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"worker_rlimit_nofile": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"worker_shutdown_timeout": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"working_directory": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"xclient": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"xml_entities": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"xslt_last_modified": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"xslt_param": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"xslt_string_param": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"xslt_stylesheet": {
		ngxHttpLocConf, ngxConf1More,
	},
	"xslt_types": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"zone": {
		ngxHttpUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},

	// nginx+ directives {definitions inferred from docs}
	"api": {
		ngxHttpLocConf, ngxConfNoArgs, ngxConfTake1,
	},
	"auth_jwt": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"auth_jwt_claim_set": {
		ngxHttpMainConf, ngxConf2More,
	},
	"auth_jwt_header_set": {
		ngxHttpMainConf, ngxConf2More,
	},
	"auth_jwt_key_file": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"auth_jwt_key_request": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"auth_jwt_leeway": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"f4f": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"f4f_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"fastcgi_cache_purge": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"health_check": {
		ngxHttpLocConf, ngxConfAny,
		ngxStreamSrvConf, ngxConfAny,
	},
	"health_check_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"hls": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"hls_buffers": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake2,
	},
	"hls_forward_args": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"hls_fragment": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"hls_mp4_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"hls_mp4_max_buffer_size": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"js_access": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_content": {
		ngxHttpLocConf, ngxHttpLmtConf, ngxConfTake1,
	},
	"js_filter": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_include": {
		ngxHttpMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"js_path": {
		ngxHttpMainConf, ngxConfTake1,
	},
	"js_preread": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_set": {
		ngxHttpMainConf, ngxConfTake2,
		ngxStreamMainConf, ngxConfTake2,
	},
	"keyval": {
		ngxHttpMainConf, ngxConfTake3,
		ngxStreamMainConf, ngxConfTake3,
	},
	"keyval_zone": {
		ngxHttpMainConf, ngxConf1More,
		ngxStreamMainConf, ngxConf1More,
	},
	"least_time": {
		ngxHttpUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},
	"limit_zone": {
		ngxHttpMainConf, ngxConfTake3,
	},
	"match": {
		ngxHttpMainConf, ngxConfBlock, ngxConfTake1,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake1,
	},
	"memcached_force_ranges": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"mp4_limit_rate": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"mp4_limit_rate_after": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"ntlm": {
		ngxHttpUpsConf, ngxConfNoArgs,
	},
	"proxy_cache_purge": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"proxy_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_requests": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"queue": {
		ngxHttpUpsConf, ngxConfTake12,
	},
	"scgi_cache_purge": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"scgi_socket_keepalive": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfFlag,
	},
	"session_log": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake1,
	},
	"session_log_format": {
		ngxHttpMainConf, ngxConf2More,
	},
	"session_log_zone": {
		ngxHttpMainConf, ngxConfTake23, ngxConfTake4, ngxConfTake5, ngxConfTake6,
	},
	"state": {
		ngxHttpUpsConf, ngxConfTake1,
		ngxStreamUpsConf, ngxConfTake1,
	},
	"status": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"status_format": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConfTake12,
	},
	"status_zone": {
		ngxHttpSrvConf, ngxConfTake1,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"sticky": {
		ngxHttpUpsConf, ngxConf1More,
	},
	"sticky_cookie_insert": {
		ngxHttpUpsConf, ngxConfTake1234,
	},
	"upstream_conf": {
		ngxHttpLocConf, ngxConfNoArgs,
	},
	"uwsgi_cache_purge": {
		ngxHttpMainConf, ngxHttpSrvConf, ngxHttpLocConf, ngxConf1More,
	},
	"zone_sync": {
		ngxStreamSrvConf, ngxConfNoArgs,
	},
	"zone_sync_buffers": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake2,
	},
	"zone_sync_connect_retry_interval": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_connect_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_interval": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_recv_buffer_size": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_server": {
		ngxStreamSrvConf, ngxConfTake12,
	},
	"zone_sync_ssl": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"zone_sync_ssl_certificate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_certificate_key": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_ciphers": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_crl": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_name": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_password_file": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_protocols": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"zone_sync_ssl_server_name": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"zone_sync_ssl_trusted_certificate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_ssl_verify": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"zone_sync_ssl_verify_depth": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"zone_sync_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
}

// Context - contexts to a key to its bitmasks in Mask
var Context = map[[3]string]Bits{
	{}:                                   ngxMainConf,
	{"events"}:                           ngxEventConf,
	{"mail"}:                             ngxMailMainConf,
	{"mail", "server"}:                   ngxMailSrvConf,
	{"stream"}:                           ngxStreamMainConf,
	{"stream", "server"}:                 ngxStreamSrvConf,
	{"stream", "upstream"}:               ngxStreamUpsConf,
	{"http"}:                             ngxHttpMainConf,
	{"http", "server"}:                   ngxHttpSrvConf,
	{"http", "location"}:                 ngxHttpLocConf,
	{"http", "upstream"}:                 ngxHttpUpsConf,
	{"http", "server", "if"}:             ngxHttpSifConf,
	{"http", "location", "if"}:           ngxHttpLifConf,
	{"http", "location", "limit_except"}: ngxHttpLmtConf,
}

func analyze(fname string, stmt statement, term string, ctx [3]string, strict bool, checkCtx bool, checkArg bool) error {
	directive := stmt.directive
	dir := checkDirective(directive, Directives)

	if strict && !dir {
		return fmt.Errorf("unknown directive %v", directive)
	}

	ct := checkContext(ctx, Context)
	// if we don't know where this directive is allowed and how
	// many arguments it can take then don't bother analyzing it
	if !ct || !dir {
		return fmt.Errorf("context or directive in invalid")
	}

	args := stmt.args
	// makes numArgs an unsigned int for bit shifting later
	numArgs := uint(len(args))

	masks := Directives[directive]
	// if this directive can't be used in this context then throw an error
	if checkCtx {
		for _, mask := range masks {
			bitmask := Context[ctx]
			if mask&bitmask != 0x00000000 {
				masks = append(masks, mask)
			}
		}

		if len(masks) == 0 {
			return fmt.Errorf("%v directive is not allowed here", directive)
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
		if msk&NGX_CONF_BLOCK != 0x00000000 && term != "{" {
			reason = fmt.Sprintf("diretive %v has no opening '{'", directive)
			continue
		}
		//if the directive is a block but shouldn't be according to the mask
		if msk&NGX_CONF_BLOCK != 0x00000000 && term != ";" {
			reason = fmt.Sprintf("directive %v is not terminated by ';'", directive)
			continue
		}
		// use mask to check the directive's arguments
		if ((msk>>numArgs)&1 != 0x00000000 && numArgs <= 7) || //NOARGS to TAKE7
			(msk&NGX_CONF_FLAG != 0x00000000 && numArgs == 1 && validFlags(stmt.args[0])) ||
			(msk&NGX_CONF_ANY != 0x00000000) ||
			(msk&NGX_CONF_1MORE != 0x00000000 && numArgs >= 1) ||
			(msk&NGX_CONF_2MORE != 0x00000000 && numArgs >= 2) {
			return nil
		} else if msk&NGX_CONF_FLAG != 0x00000000 && numArgs == 1 && !validFlags(stmt.args[0]) {
			reason = fmt.Sprintf("invalid value %v in %v directive, it must be 'on' or 'off'", stmt.args[0], stmt.directive)
			continue
		} else {
			reason = fmt.Sprintf("invalid number of arguements in %v", directive)
		}
	}
	if reason == "" {
		return nil
	}
	return fmt.Errorf(reason)
}

func checkContext(cont [3]string, contexts map[[3]string]Bits) bool {
	if _, ok := contexts[cont]; ok {
		return true
	}
	return false
}

func checkDirective(dir string, direct map[string][]Bits) bool {
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

func registerExternalDirectives(directives map[string][]Bits) {
	for d, b := range directives {
		Directives[d] = []Bits{}
		for _, v := range b {
			if v != 0x00000000 {
				Directives[d] = append(Directives[d], v)
			}
		}
	}
}
