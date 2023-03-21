package model

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/terraform-provider-qovery/qovery/modifiers"
)

var ClusterSettingsDefault = map[string]AdvSettingAttr{
	"aws.cloudwatch.eks_logs_retention_days": {"Maximum retention days in Cloudwatch for EKS logs", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(90),
	}, types.Int64{Value: 90}},
	"aws.iam.admin_group": {"Allows you to specify the IAM group name associated to the Qovery user", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("Admins"),
	}, types.String{Value: "Admins"}},
	"aws.vpc.enable_s3_flow_logs": {"Enable flow logs on the cluster VPC and store them in an s3 bucket", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"aws.vpc.flow_logs_retention_days": {"Set the number of retention days for flow logs. Unlimited retention with value", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(365),
	}, types.Int64{Value: 365}},
	"cloud_provider.container_registry.tags": {"Add additional tags on the cluster dedicated registry", types.MapType{ElemType: types.StringType}, tfsdk.AttributePlanModifiers{
		modifiers.NewStringMapDefaultModifier(map[string]string{}),
	}, types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}}},
	"database.mongodb.allowed_cidrs": {"List of allowed CIDRS", types.SetType{ElemType: types.StringType}, tfsdk.AttributePlanModifiers{
		modifiers.NewStringSetDefaultModifier([]string{"0.0.0.0/0"}),
	}, types.Set{ElemType: types.StringType, Elems: []attr.Value{types.String{Value: "0.0.0.0/0"}}}},
	"database.mongodb.deny_public_access": {"Deny public access to all MongoDB databases", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"database.mysql.allowed_cidrs": {"List of allowed CIDRS", types.SetType{ElemType: types.StringType}, tfsdk.AttributePlanModifiers{
		modifiers.NewStringSetDefaultModifier([]string{"0.0.0.0/0"}),
	}, types.Set{ElemType: types.StringType, Elems: []attr.Value{types.String{Value: "0.0.0.0/0"}}}},
	"database.mysql.deny_public_access": {"Deny public access to all MySQL databases", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"database.postgresql.allowed_cidrs": {"List of allowed CIDRS", types.SetType{ElemType: types.StringType}, tfsdk.AttributePlanModifiers{
		modifiers.NewStringSetDefaultModifier([]string{"0.0.0.0/0"}),
	}, types.Set{ElemType: types.StringType, Elems: []attr.Value{types.String{Value: "0.0.0.0/0"}}}},
	"database.postgresql.deny_public_access": {"Deny public access to all PostgreSQL databases", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"database.redis.allowed_cidrs": {"List of allowed CIDRS", types.SetType{ElemType: types.StringType}, tfsdk.AttributePlanModifiers{
		modifiers.NewStringSetDefaultModifier([]string{"0.0.0.0/0"}),
	}, types.Set{ElemType: types.StringType, Elems: []attr.Value{types.String{Value: "0.0.0.0/0"}}}},
	"database.redis.deny_public_access": {"Deny public access to all Redis databases", types.BoolType, tfsdk.AttributePlanModifiers{
		modifiers.NewBoolDefaultModifier(false),
	}, types.Bool{Value: false}},
	"load_balancer.size": {"Allows you to specify the load balancer size in front of your cluster", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("lb-s"),
	}, types.String{Value: "lb-s"}},
	"loki.log_retention_in_week": {"Maximum Kubernetes pods (containers/application/jobs/cronjob) retention logs in weeks", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(12),
	}, types.Int64{Value: 12}},
	"registry.image_retention_time": {"Allows you to specify an amount in seconds after which images in the default registry are deleted", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(31536000),
	}, types.Int64{Value: 31536000}},
	"pleco.resources_ttl": {"Deprecated", types.Int64Type, tfsdk.AttributePlanModifiers{
		modifiers.NewInt64DefaultModifier(-1),
	}, types.Int64{Value: -1}},
	"aws.eks.ec2.metadata_imds": {"Specify the IMDS version you want to use. Possible values are Required (IMDS v2 only) and Optional (IMDS v1 and V2)", types.StringType, tfsdk.AttributePlanModifiers{
		modifiers.NewStringDefaultModifier("optional"),
	}, types.String{Value: "optional"}},
}
