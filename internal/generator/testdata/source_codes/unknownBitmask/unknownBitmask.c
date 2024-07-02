static ngx_command_t my_directives[] = {

    { ngx_string("my_directive_1"),
      NGX_HTTP_MAIN_CONF|FAKE_BITMASK, // unknown bitmask FAKE_BITMASK
      0,
      0,
      0,
      NULL }, 
    { ngx_string("my_directive_2"),
      NGX_HTTP_MAIN_CONF|NGX_CONF_FLAG,
      0,
      0,
      0,
      NULL },
    { ngx_string("my_directive_3"),
      NGX_HTTP_MAIN_CONF|NGX_HTTP_SRV_CONF|NGX_CONF_NOARGS,
      0,
      0,
      0,
      NULL },

    ngx_null_command
};