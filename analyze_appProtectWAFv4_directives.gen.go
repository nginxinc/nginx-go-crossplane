/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

// Code generated by generator; DO NOT EDIT.
// All the definitions are extracted from the source code
// Each bit mask describes these behaviors:
//   - how many arguments the directive can take
//   - whether or not it is a block directive
//   - whether this is a flag (takes one argument that's either "on" or "off")
//   - which contexts it's allowed to be in

package crossplane

var appProtectWAFv4Directives = map[string][]uint{
    "app_protect_app_name": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
    },
    "app_protect_compressed_requests_action": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_config_set_timeout": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_cookie_seed": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_cpu_thresholds": {
        ngxHTTPMainConf | ngxConfTake2,
    },
    "app_protect_custom_log_attribute": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
    },
    "app_protect_enable": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfFlag,
    },
    "app_protect_enforcer_address": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_enforcer_memory_limit_mb": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_failure_mode_action": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_global_settings": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_logging_str": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
    },
    "app_protect_physical_memory_util_thresholds": {
        ngxHTTPMainConf | ngxConfTake2,
    },
    "app_protect_policy_file": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake1,
    },
    "app_protect_reconnect_period_seconds": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_request_buffer_overflow_action": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_response_enforcement_disable": {
        ngxHTTPMainConf | ngxConfTake1,
    },
    "app_protect_security_log": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfTake2,
    },
    "app_protect_security_log_enable": {
        ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxConfFlag,
    },
    "app_protect_streaming_buffer_watermarks": {
        ngxHTTPMainConf | ngxConfTake2,
    },
    "app_protect_user_defined_signatures": {
        ngxHTTPMainConf | ngxConfTake1,
    },
}

// MatchAppProtectWAFv4 is a MatchFunc for App Protect v4 module.
func MatchAppProtectWAFv4(directive string) ([]uint, bool) {
    m, ok := appProtectWAFv4Directives[directive]
    return m, ok
}
