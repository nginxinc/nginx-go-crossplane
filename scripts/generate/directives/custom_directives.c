
#include <ngx_config.h>
#include <ngx_core.h>
#include <ngx_http.h>

static char *ngx_conf_set_nginxaas_scaling_group(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) { return NGX_CONF_OK; }
static char *ngx_conf_set_nginxaas_scaling_group_opts(ngx_conf_t *cf, ngx_command_t *cmd, void *conf) { return NGX_CONF_OK; }

static ngx_command_t custom_commands[] = {
    { ngx_string("nginxaas-scaling-group"), NGX_HTTP_UPS_CONF|NGX_CONF_TAKE1, ngx_conf_set_nginxaas_scaling_group, 0, 0, NULL },
    { ngx_string("nginxaas-scaling-group-opts"), NGX_HTTP_UPS_CONF|NGX_CONF_1MORE, ngx_conf_set_nginxaas_scaling_group_opts, 0, 0, NULL },
    ngx_null_command
};

static ngx_http_module_t custom_module_ctx = {
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
};

ngx_module_t custom_module = {
    NGX_MODULE_V1,
    &custom_module_ctx,
    custom_commands,
    NGX_HTTP_MODULE,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL,
    NGX_MODULE_V1_PADDING
};
