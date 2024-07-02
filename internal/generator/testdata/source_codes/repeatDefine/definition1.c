static ngx_command_t my_directives[] = {

    { ngx_string("my_directive_1"),
      NGX_HTTP_MAIN_CONF|NGX_CONF_TAKE2,
      0,
      0,
      0,
      NULL }, /*comments will be igored by generator*/
    { ngx_string("my_directive_2"),
      NGX_HTTP_MAIN_CONF|NGX_CONF_FLAG,/*comments will be igored by generator*/
      0,
      0,
      0,
      NULL },
    { ngx_string("my_directive_3"),
      NGX_HTTP_MAIN_CONF|NGX_HTTP_SRV_CONF|NGX_CONF_NOARGS,//comments will be igored by generator
      0,
      0,
      0,
      NULL },

    ngx_null_command
};