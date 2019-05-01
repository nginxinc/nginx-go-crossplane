package analyzer

import (
	"fmt"
	"strings"
)


// Statement struct used for analysing directives and other information
type Statement struct {
	directive string
	args      [1]string
	line      int

}

// Bits type for constant bit vars
type Bits uint

const (
	ngxDirectConf     Bits = 0x00010000 // main file (not used)
	ngxMainConf       Bits = 0x00040000 // main context
	ngxEventConf      Bits = 0x00080000 // events
	ngxMailMainConf   Bits = 0x00100000 // mail
	ngxMailSrvConf    Bits = 0x00200000 // mail > server
	ngxStreamMainConf Bits = 0x00400000 // stream
	ngxStreamSrvConf  Bits = 0x00800000 // stream > server
	ngxStreamUpsConf  Bits = 0x01000000 // stream > upstream
	ngxHTTPMainConf   Bits = 0x02000000 // http
	ngxHTTPSrvConf    Bits = 0x04000000 // http > server
	ngxHTTPLocConf    Bits = 0x08000000 // http > location
	ngxHTTPUpsConf    Bits = 0x10000000 // http > upstream
	ngxHTTPSifConf    Bits = 0x20000000 // http > server > if
	ngxHTTPLifConf    Bits = 0x40000000 // http > location > if
	ngxHTTPLmtConf    Bits = 0x80000000
	ngxConfTake12     Bits = (ngxConfTake1 | ngxConfTake2)
	ngxConfTake13     Bits = (ngxConfTake1 | ngxConfTake3)
	ngxConfTake23     Bits = (ngxConfTake2 | ngxConfTake3)
	ngxConfTake123    Bits = (ngxConfTake12 | ngxConfTake3)
	ngxConfTake1234   Bits = (ngxConfTake123 | ngxConfTake4)

	// bit masks for different directive argument styles
	ngxConfNoArgs Bits = 0x00000001 // 0 args
	ngxConfTake1  Bits = 0x00000002 // 1 args
	ngxConfTake2  Bits = 0x00000004 // 2 args
	ngxConfTake3  Bits = 0x00000008 // 3 args
	ngxConfTake4  Bits = 0x00000010 // 4 args
	ngxConfTake5  Bits = 0x00000020 // 5 args
	ngxConfTake6  Bits = 0x00000040 // 6 args
	ngxConfTake7  Bits = 0x00000080 // 7 args
	ngxConfBlock  Bits = 0x00000100 // followed by block
	ngxConfFlag   Bits = 0x00000200 // 'on' or 'off'
	ngxConfAny    Bits = 0x00000400 // >=0 args
	ngxConf1More  Bits = 0x00000800 // >=1 args
	ngxConf2More  Bits = 0x00001000 // >=2 args

	ngxAnyConf Bits = (ngxMainConf | ngxEventConf | ngxMailMainConf | ngxMailSrvConf |
		ngxStreamMainConf | ngxStreamSrvConf | ngxStreamUpsConf |
		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPUpsConf)
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"accept_mutex": {
		ngxEventConf, ngxConfFlag,
	},
	"accept_mutex_delay": {
		ngxEventConf, ngxConfTake1,
	},
	"access_log": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxHTTPLmtConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"add_after_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"add_before_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"add_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake23,
	},
	"add_trailer": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake23,
	},
	"addition_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"aio": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"aio_write": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"alias": {
		ngxHTTPLocConf, ngxConfTake1,
	},
	"allow": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ancient_browser": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"ancient_browser_value": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"auth_basic": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
	},
	"auth_basic_user_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"auth_request_set": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"autoindex": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"autoindex_exact_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"autoindex_format": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"autoindex_localtime": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"break": {
		ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfNoArgs,
	},
	"charset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"charset_map": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake2,
	},
	"charset_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"chunked_transfer_encoding": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"client_body_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"client_body_in_file_only": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"client_body_in_single_buffer": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"client_body_temp_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
	},
	"client_body_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"client_header_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"client_header_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"client_max_body_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"connection_pool_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"create_full_put_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"daemon": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"dav_access": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"dav_methods": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"debug_connection": {
		ngxEventConf, ngxConfTake1,
	},
	"debug_points": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"default_type": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"deny": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"directio": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"directio_alignment": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"disable_symlinks": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"empty_gif": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"env": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"error_log": {
		ngxMainConf, ngxConf1More,
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"error_page": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConf2More,
	},
	"etag": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"events": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"expires": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake12,
	},
	"fastcgi_bind": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"fastcgi_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"fastcgi_busy_buffers_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_background_update": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_cache_bypass": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_cache_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_lock": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_cache_lock_age": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_lock_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_max_range_offset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_methods": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_cache_min_uses": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_path": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"fastcgi_cache_revalidate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_cache_use_stale": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_cache_valid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_catch_stderr": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_force_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_hide_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_ignore_client_abort": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_ignore_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_index": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_intercept_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_keep_conn": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_max_temp_file_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_no_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"fastcgi_param": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake23,
	},
	"fastcgi_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"fastcgi_pass_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_pass_request_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_pass_request_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_request_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_send_lowat": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_split_path_info": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_store": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_store_access": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"fastcgi_temp_file_write_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_temp_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
	},
	"fastcgi_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"fastcgi_socket_keepalive": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
	},
	"flv": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"geo": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake12,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake12,
	},
	"geoip_city": {
		ngxHTTPMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_country": {
		ngxHTTPMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_org": {
		ngxHTTPMainConf, ngxConfTake12,
		ngxStreamMainConf, ngxConfTake12,
	},
	"geoip_proxy": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"geoip_proxy_recursive": {
		ngxHTTPMainConf, ngxConfFlag,
	},
	"google_perftools_profiles": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},

	"grpc_bind": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"grpc_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_hide_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ignore_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"grpc_intercept_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"grpc_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"grpc_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"grpc_pass_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_set_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"grpc_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"grpc_ssl_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_certificate_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_ciphers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_crl": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_password_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_protocols": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"grpc_ssl_server_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"grpc_ssl_session_reuse": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"grpc_ssl_trusted_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"grpc_ssl_verify": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"grpc_ssl_verify_depth": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},

	"gunzip": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"gunzip_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"gzip": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"gzip_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"gzip_comp_level": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"gzip_disable": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"gzip_http_version": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"gzip_min_length": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"gzip_proxied": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"gzip_static": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"gzip_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"gzip_vary": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"hash": {
		ngxHTTPUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},
	"http": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"http2_body_preread_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_chunk_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"http2_idle_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_max_concurrent_streams": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_max_field_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_max_header_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_max_requests": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_recv_buffer_size": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"http2_recv_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_push": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"http2_push_preload": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"http2_max_concurrent_pushes": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"http2_push": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_TAKE1,
	},
	"http2_push_preload": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
	},
	"http2_max_concurrent_pushes": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_CONF_TAKE1,
	},
	"if": {
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfBlock, ngxConf1More,
	},
	"if_modified_since": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"ignore_invalid_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"image_filter": {
		ngxHTTPLocConf, ngxConfTake123,
	},
	"image_filter_buffer": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"image_filter_interlace": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"image_filter_jpeg_quality": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"image_filter_sharpen": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"image_filter_transparency": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"image_filter_webp_quality": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
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
		ngxAnyConf, ngxConfTake1,
	},
	"index": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"internal": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"ip_hash": {
		ngxHTTPUpsConf, ngxConfNoArgs,
	},
	"keepalive": {
		ngxHTTPUpsConf, ngxConfTake1,
	},
	"keepalive_disable": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"keepalive_requests": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxHTTPUpsConf, ngxConfTake1,
	},
	"keepalive_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
		ngxHTTPUpsConf, ngxConfTake1,
	},
	"large_client_header_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake2,
	},
	"least_conn": {
		ngxHTTPUpsConf, ngxConfNoArgs,
		ngxStreamUpsConf, ngxConfNoArgs,
	},
	"limit_conn": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake2,
	},
	"limit_conn_log_level": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"limit_conn_status": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"limit_conn_zone": {
		ngxHTTPMainConf, ngxConfTake2,
		ngxStreamMainConf, ngxConfTake2,
	},
	"limit_except": {
		ngxHTTPLocConf, ngxConfBlock, ngxConf1More,
	},
	"limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"limit_rate_after": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"limit_req": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"limit_req_log_level": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"limit_req_status": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"limit_req_zone": {
		ngxHTTPMainConf, ngxConfTake3,
	},
	"lingering_close": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"lingering_time": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"lingering_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"listen": {
		ngxHTTPSrvConf, ngxConf1More,
		ngxMailSrvConf, ngxConf1More,
		ngxStreamSrvConf, ngxConf1More,
	},
	"load_module": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"location": {
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfBlock, ngxConfTake12,
	},
	"lock_file": {
		ngxMainConf, ngxDirectConf, ngxConfTake1,
	},
	"log_format": {
		ngxHTTPMainConf, ngxConf2More,
		ngxStreamMainConf, ngxConf2More,
	},
	"log_not_found": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"log_subrequest": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"mail": {
		ngxMainConf, ngxConfBlock, ngxConfNoArgs,
	},
	"map": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake2,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake2,
	},
	"map_hash_bucket_size": {
		ngxHTTPMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"map_hash_max_size": {
		ngxHTTPMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"master_process": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"max_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_bind": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"memcached_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_gzip_flag": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"memcached_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"memcached_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"memcached_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"memcached_socket_keepalive": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
	},
	"merge_slashes": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"min_delete_depth": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"mirror": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"mirror_request_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"modern_browser": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"modern_browser_value": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"mp4": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"mp4_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"mp4_max_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"msie_padding": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"msie_refresh": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"multi_accept": {
		ngxEventConf, ngxConfFlag,
	},
	"open_file_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"open_file_cache_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"open_file_cache_min_uses": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"open_file_cache_valid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"open_log_file_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1234,
	},
	"output_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"override_charset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"pcre_jit": {
		ngxMainConf, ngxDirectConf, ngxConfFlag,
	},
	"perl": {
		ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
	},
	"perl_modules": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"perl_require": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"perl_set": {
		ngxHTTPMainConf, ngxConfTake2,
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"postpone_output": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake12,
	},
	"proxy_buffer": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"proxy_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"proxy_busy_buffers_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_background_update": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_cache_bypass": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_cache_convert_head": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_cache_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_lock": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_cache_lock_age": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_lock_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_max_range_offset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_methods": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_cache_min_uses": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_cache_path": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"proxy_cache_revalidate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_cache_use_stale": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_cache_valid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_cookie_domain": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"proxy_cookie_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"proxy_download_rate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_force_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_headers_hash_bucket_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_headers_hash_max_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_hide_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_http_version": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_ignore_client_abort": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_ignore_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_intercept_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_max_temp_file_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_method": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_no_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxHTTPLmtConf, ngxConfTake1,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_pass_error_message": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"proxy_pass_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_pass_request_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_pass_request_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_protocol": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_protocol_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_redirect": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"proxy_request_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"proxy_responses": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_send_lowat": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_set_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_set_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"proxy_ssl": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_certificate_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_ciphers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_crl": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_password_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_protocols": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"proxy_ssl_server_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_session_reuse": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_trusted_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_ssl_verify": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_ssl_verify_depth": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_store": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_store_access": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"proxy_temp_file_write_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"proxy_temp_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
	},
	"proxy_timeout": {
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_upload_rate": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"random": {
		ngxHTTPUpsConf, ngxConfNoArgs, ngxConfTake12,
		ngxStreamUpsConf, ngxConfNoArgs, ngxConfTake12,
	},
	"random": {
		NGX_HTTP_UPS_CONF, NGX_CONF_NOARGS, NGX_CONF_TAKE12,
		NGX_STREAM_UPS_CONF, NGX_CONF_NOARGS, NGX_CONF_TAKE12,
	},
	"random_index": {
		ngxHTTPLocConf, ngxConfFlag,
	},
	"read_ahead": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"real_ip_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"real_ip_recursive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"recursive_error_pages": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"referer_hash_bucket_size": {
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"referer_hash_max_size": {
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"request_pool_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"reset_timedout_connection": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"resolver": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"resolver_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"return": {
		ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake12,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"rewrite": {
		ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake23,
	},
	"rewrite_log": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"root": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"satisfy": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_bind": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"scgi_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"scgi_busy_buffers_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_background_update": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_cache_bypass": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_cache_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_lock": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_cache_lock_age": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_lock_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_max_range_offset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_methods": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_cache_min_uses": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_cache_path": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"scgi_cache_revalidate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_cache_use_stale": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_cache_valid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_force_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_hide_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_ignore_client_abort": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_ignore_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_intercept_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_max_temp_file_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_no_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_param": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake23,
	},
	"scgi_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"scgi_pass_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_pass_request_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_pass_request_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_request_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_store": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_store_access": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"scgi_temp_file_write_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"scgi_temp_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
	},
	"secure_link": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"secure_link_md5": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"secure_link_secret": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"send_lowat": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"sendfile": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"sendfile_max_chunk": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"server": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxHTTPUpsConf, ngxConf1More,
		ngxMailMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxStreamMainConf, ngxConfBlock, ngxConfNoArgs,
		ngxStreamUpsConf, ngxConf1More,
	},
	"server_name": {
		ngxHTTPSrvConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
	},
	"server_name_in_redirect": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"server_names_hash_bucket_size": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"server_names_hash_max_size": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"server_tokens": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"set": {
		ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake2,
	},
	"set_real_ip_from": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"slice": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
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
	"smtp_client_buffer": {
		NGX_MAIL_MAIN_CONF, NGX_MAIL_SRV_CONF, NGX_CONF_TAKE1,
	},
	"smtp_greeting_delay": {
		NGX_MAIL_MAIN_CONF, NGX_MAIL_SRV_CONF, NGX_CONF_TAKE1,
	},
	"source_charset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"spdy_chunk_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"spdy_headers_comp": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"split_clients": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake2,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake2,
	},
	"ssi": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"ssi_last_modified": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"ssi_min_file_chunk": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"ssi_silent_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"ssi_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"ssi_value_length": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"ssl": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
	},
	"ssl_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"ssl_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_certificate_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_ciphers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_client_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_crl": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_dhparam": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_early_data": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"ssl_early_data": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_CONF_FLAG,
	},
	"ssl_ecdh_curve": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_prefer_server_ciphers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_preread": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_protocols": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConf1More,
		ngxMailMainConf, ngxMailSrvConf, ngxConf1More,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConf1More,
	},
	"ssl_session_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake12,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake12,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake12,
	},
	"ssl_session_ticket_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_session_tickets": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
		ngxMailMainConf, ngxMailSrvConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"ssl_session_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_stapling": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"ssl_stapling_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"ssl_stapling_responder": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
	},
	"ssl_stapling_verify": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"ssl_trusted_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_verify_client": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
		ngxMailMainConf, ngxMailSrvConf, ngxConfTake1,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"ssl_verify_depth": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfTake1,
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
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfNoArgs, ngxConfTake1,
	},
	"sub_filter": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"sub_filter_last_modified": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"sub_filter_once": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"sub_filter_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"subrequest_output_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"subrequest_output_buffer_size": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_TAKE1,
	},
	"tcp_nodelay": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"tcp_nopush": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
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
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf2More,
	},
	"types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfBlock, ngxConfNoArgs,
	},
	"types_hash_bucket_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"types_hash_max_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"underscores_in_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxConfFlag,
	},
	"uninitialized_variable_warn": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPSifConf, ngxHTTPLocConf, ngxHTTPLifConf, ngxConfFlag,
	},
	"upstream": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake1,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake1,
	},
	"use": {
		ngxEventConf, ngxConfTake1,
	},
	"user": {
		ngxMainConf, ngxDirectConf, ngxConfTake12,
	},
	"userid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_domain": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_expires": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_mark": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_p3p": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"userid_service": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_bind": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"uwsgi_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"uwsgi_busy_buffers_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_background_update": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_bypass": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_cache_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_lock": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_cache_lock_age": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_lock_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_max_range_offset": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_methods": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_cache_min_uses": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_cache_path": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"uwsgi_cache_revalidate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_cache_use_stale": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_cache_valid": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_connect_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_force_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_hide_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ignore_client_abort": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_ignore_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_intercept_errors": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_max_temp_file_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_modifier1": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_modifier2": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_next_upstream": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_next_upstream_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_next_upstream_tries": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_no_cache": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_param": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake23,
	},
	"uwsgi_pass": {
		ngxHTTPLocConf, ngxHTTPLifConf, ngxConfTake1,
	},
	"uwsgi_pass_header": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_pass_request_body": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_pass_request_headers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_read_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_request_buffering": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_send_timeout": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_socket_keepalive": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
	},
	"uwsgi_ssl_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_certificate_key": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_ciphers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_crl": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_password_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_protocols": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"uwsgi_ssl_server_name": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_session_reuse": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_trusted_certificate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_ssl_verify": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"uwsgi_ssl_verify_depth": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_store": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_store_access": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake123,
	},
	"uwsgi_temp_file_write_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"uwsgi_temp_path": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1234,
	},
	"valid_referers": {
		ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"variables_hash_bucket_size": {
		ngxHTTPMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"variables_hash_max_size": {
		ngxHTTPMainConf, ngxConfTake1,
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
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"xslt_last_modified": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"xslt_param": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"xslt_string_param": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"xslt_stylesheet": {
		ngxHTTPLocConf, ngxConf1More,
	},
	"xslt_types": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"zone": {
		ngxHTTPUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},

	// nginx+ directives {definitions inferred from docs}
  "api": {
		ngxHTTPLocConf, ngxConfNoArgs, ngxConfTake1,
	},
	"auth_jwt": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"auth_jwt_claim_set": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"auth_jwt_header_set": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"auth_jwt_key_file": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"auth_jwt_key_request": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"auth_jwt_leeway": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"auth_jwt_key_request": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_TAKE1,
	},
	"auth_jwt_leeway": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_TAKE1,
	},
	"f4f": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"f4f_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"fastcgi_cache_purge": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"health_check": {
		ngxHTTPLocConf, ngxConfAny,
		ngxStreamSrvConf, ngxConfAny,
	},
	"health_check_timeout": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"hls": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"hls_buffers": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake2,
	},
	"hls_forward_args": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"hls_fragment": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"hls_mp4_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"hls_mp4_max_buffer_size": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"js_access": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_content": {
		ngxHTTPLocConf, ngxHTTPLmtConf, ngxConfTake1,
	},
	"js_filter": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_include": {
		ngxHTTPMainConf, ngxConfTake1,
		ngxStreamMainConf, ngxConfTake1,
	},
	"js_path": {
		ngxHTTPMainConf, ngxConfTake1,
	},
	"js_path": {
		NGX_HTTP_MAIN_CONF, NGX_CONF_TAKE1,
	},
	"js_preread": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"js_set": {
		ngxHTTPMainConf, ngxConfTake2,
		ngxStreamMainConf, ngxConfTake2,
	},
	"keyval": {
		ngxHTTPMainConf, ngxConfTake3,
		ngxStreamMainConf, ngxConfTake3,
	},
	"keyval_zone": {
		ngxHTTPMainConf, ngxConf1More,
		ngxStreamMainConf, ngxConf1More,
	},
	"keyval": {
		NGX_HTTP_MAIN_CONF, NGX_CONF_TAKE3,
		NGX_STREAM_MAIN_CONF, NGX_CONF_TAKE3,
	},
	"keyval_zone": {
		NGX_HTTP_MAIN_CONF, NGX_CONF_1MORE,
		NGX_STREAM_MAIN_CONF, NGX_CONF_1MORE,
	},
	"least_time": {
		ngxHTTPUpsConf, ngxConfTake12,
		ngxStreamUpsConf, ngxConfTake12,
	},
	"limit_zone": {
		ngxHTTPMainConf, ngxConfTake3,
	},
	"match": {
		ngxHTTPMainConf, ngxConfBlock, ngxConfTake1,
		ngxStreamMainConf, ngxConfBlock, ngxConfTake1,
	},
	"memcached_force_ranges": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"mp4_limit_rate": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"mp4_limit_rate_after": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"ntlm": {
		ngxHTTPUpsConf, ngxConfNoArgs,
	},
	"proxy_cache_purge": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"proxy_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfFlag,
	},
	"proxy_requests": {
		ngxStreamMainConf, ngxStreamSrvConf, ngxConfTake1,
	},
	"proxy_socket_keepalive": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_FLAG,
	},
	"proxy_requests": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"queue": {
		ngxHTTPUpsConf, ngxConfTake12,
	},
	"scgi_cache_purge": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
	},
	"scgi_socket_keepalive": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfFlag,
	},
	"scgi_socket_keepalive": {
		NGX_HTTP_MAIN_CONF, NGX_HTTP_SRV_CONF, NGX_HTTP_LOC_CONF, NGX_CONF_FLAG,
	},
	"session_log": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake1,
	},
	"session_log_format": {
		ngxHTTPMainConf, ngxConf2More,
	},
	"session_log_zone": {
		ngxHTTPMainConf, ngxConfTake23, ngxConfTake4, ngxConfTake5, ngxConfTake6,
	},
	"state": {
		ngxHTTPUpsConf, ngxConfTake1,
		ngxStreamUpsConf, ngxConfTake1,
	},
	"status": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"status_format": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConfTake12,
	},
	"status_zone": {
		ngxHTTPSrvConf, ngxConfTake1,
		ngxStreamSrvConf, ngxConfTake1,
	},
	"sticky": {
		ngxHTTPUpsConf, ngxConf1More,
	},
	"sticky_cookie_insert": {
		ngxHTTPUpsConf, ngxConfTake1234,
	},
	"upstream_conf": {
		ngxHTTPLocConf, ngxConfNoArgs,
	},
	"uwsgi_cache_purge": {
		ngxHTTPMainConf, ngxHTTPSrvConf, ngxHTTPLocConf, ngxConf1More,
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
	"zone_sync": {
		NGX_STREAM_SRV_CONF, NGX_CONF_NOARGS,
	},
	"zone_sync_buffers": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE2,
	},
	"zone_sync_connect_retry_interval": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_connect_timeout": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_interval": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_recv_buffer_size": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_server": {
		NGX_STREAM_SRV_CONF, NGX_CONF_TAKE12,
	},
	"zone_sync_ssl": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_FLAG,
	},
	"zone_sync_ssl_certificate": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_certificate_key": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_ciphers": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_crl": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_name": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_password_file": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_protocols": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_1MORE,
	},
	"zone_sync_ssl_server_name": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_FLAG,
	},
	"zone_sync_ssl_trusted_certificate": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_ssl_verify": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_FLAG,
	},
	"zone_sync_ssl_verify_depth": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
	},
	"zone_sync_timeout": {
		NGX_STREAM_MAIN_CONF, NGX_STREAM_SRV_CONF, NGX_CONF_TAKE1,
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
	{"http"}:                             ngxHTTPMainConf,
	{"http", "server"}:                   ngxHTTPSrvConf,
	{"http", "location"}:                 ngxHTTPLocConf,
	{"http", "upstream"}:                 ngxHTTPUpsConf,
	{"http", "server", "if"}:             ngxHTTPSifConf,
	{"http", "location", "if"}:           ngxHTTPLifConf,
	{"http", "location", "limit_except"}: ngxHTTPLmtConf,
}


// Analyze Directives and args in nginc.conf file
func Analyze(fname string, stmt Statement, term string, ctx [3]string, strict bool, checkCtx bool, checkArg bool) error {
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

	args := stmt.Args
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
		if msk&ngxConfBlock != 0x00000000 && term != "{" {
			reason = fmt.Sprintf("diretive %v has no opening '{'", directive)
			continue
		}
		//if the directive is a block but shouldn't be according to the mask
		if msk&ngxConfBlock != 0x00000000 && term != ";" {
			reason = fmt.Sprintf("directive %v is not terminated by ';'", directive)
			continue
		}
		// use mask to check the directive's arguments
		if ((msk>>numArgs)&1 != 0x00000000 && numArgs <= 7) || //NOARGS to TAKE7

			(msk&ngxConfFlag != 0x00000000 && numArgs == 1 && validFlags(stmt.args[0])) ||
			(msk&ngxConfAny != 0x00000000) ||
			(msk&ngxConf1More != 0x00000000 && numArgs >= 1) ||
			(msk&ngxConf2More != 0x00000000 && numArgs >= 2) {
			return nil
		} else if msk&ngxConfFlag != 0x00000000 && numArgs == 1 && !validFlags(stmt.args[0]) {
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

func EnterBlockCTX(stmt Statement, ctx [3]string) [3]string {
	if len(ctx) != 0 && ctx[0] == "http" && stmt.Directive == "location" {
		return [3]string{"http", "location"}
	}
	for i, v := range ctx {
		if v == "" {
			ctx[i] = stmt.Directive
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
