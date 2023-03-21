package model

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var NetworkSettingsDefault = map[string]AdvSettingAttr{
	"network.ingress.cors_allow_headers": {"Allows you to specify which set of headers can be present in the client request", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"),
	}, types.String{Value: "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization"}},
	"network.ingress.cors_allow_methods": {"Allows you to specify which set of methods can be used for the client request", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("GET, PUT, POST, DELETE, PATCH, OPTIONS"),
	}, types.String{Value: "GET, PUT, POST, DELETE, PATCH, OPTIONS"}},
	"network.ingress.cors_allow_origin": {"Allows you to specify which origin(s) can access a resource", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("*"),
	}, types.String{Value: "*"}},
	"network.ingress.enable_cors": {"Allows you to enable Cross-Origin Resource Sharing", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"network.ingress.enable_sticky_session": {"Allows you to enable Sticky session", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"network.ingress.keepalive_time_seconds": {"Limits the maximum time in seconds during which requests can be processed through one keepalive connection", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(3600),
	}, types.Int64{Value: 3600}},
	"network.ingress.keepalive_timeout_seconds": {"Sets a timeout in seconds during which an idle keepalive connection to an upstream server will stay open.", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
	"network.ingress.proxy_body_size_mb": {"Allows you to set in megabytes a maximum size for resources that can be downloaded from your server", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(100),
	}, types.Int64{Value: 100}},
	"network.ingress.proxy_buffer_size_kb": {"Allows you to set in kilobytes a header buffer size used while reading the response header from upstream", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(4),
	}, types.Int64{Value: 4}},
	"network.ingress.proxy_connect_timeout_seconds": {"Defines a timeout in seconds for establishing a connection with a proxied server", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
	"network.ingress.proxy_read_timeout_seconds": {"Defines a timeout in seconds for reading a response from the proxied server", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
	"network.ingress.proxy_send_timeout_seconds": {"Sets a timeout in seconds for transmitting a request to the proxied server", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
	"network.ingress.send_timeout_seconds": {"Sets a timeout in seconds for transmitting a response to the client", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(60),
	}, types.Int64{Value: 60}},
	"network.ingress.whitelist_source_range": {"Allows you to specify which IP ranges are allowed to access your application (comma-separated list of CIDRs)", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("0.0.0.0/0"),
	}, types.String{Value: "0.0.0.0/0"}},
	"network.ingress.denylist_source_range": {"Allows you to specify which IP ranges are not allowed to access your application (comma-separated list of CIDRs)", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier(""),
	}, types.String{Value: ""}},
	"network.ingress.basic_auth_env_var": {"Set the name of an environment variable to use as a basic authentication (login:crypted_password) from htpasswd command", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier(""),
	}, types.String{Value: ""}},
}
