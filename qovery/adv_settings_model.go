package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfTypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"reflect"
	"strings"
)

type AdvSettingAttr struct {
	Description string
	Type        attr.Type
}

var noDescription = "No description for this advanced setting."

// To avoid missing field, we do some introspection here to get all active advanced settings fields base on Qovery client advanced settings structs
func advSettingsFromStruct(advSettings interface{}) map[string]AdvSettingAttr {
	v := reflect.ValueOf(advSettings)
	typeOfS := v.Type()

	var advSettingsMap = make(map[string]AdvSettingAttr)
	for i := 0; i < v.NumField(); i++ {
		tag := typeOfS.Field(i).Tag.Get("json")
		key := tag[:strings.Index(tag, ",")]
		_type := tfTypefromGoTtype(typeOfS.Field(i).Type)

		advSettingsMap[key] = AdvSettingAttr{
			Description: settingsDescription(key),
			Type:        _type,
		}
	}

	return advSettingsMap
}

func tfTypefromGoTtype(goType reflect.Type) attr.Type {
	switch goType.String() {
	case "*string", "string":
		return tfTypes.StringType
	case "*int", "int", "*int64", "int64", "*int32", "int32":
		return tfTypes.Int64Type
	case "*bool", "bool":
		return tfTypes.BoolType
	case "*[]string", "[]string":
		return tfTypes.SetType{ElemType: tfTypes.StringType}
	case "*map[string]string", "map[string]string":
		return tfTypes.MapType{ElemType: tfTypes.StringType}
	}

	return tfTypes.ObjectType{AttrTypes: nil}
}

func clusterSettingsDescription(setting string) string {
	switch setting {
	case "aws.cloudwatch.eks_logs_retention_days":
		return "Maximum retention days in Cloudwatch for EKS logs"
	case "aws.eks.ec2.metadata_imds":
		return "Specify the IMDS version you want to use. Possible values are Required (IMDS v2 only) and Optional (IMDS v1 and V2)"
	case "aws.iam.admin_group":
		return "Allows you to specify the IAM group name associated to the Qovery user"
	case "aws.vpc.enable_s3_flow_logs":
		return "Enable flow logs on the cluster VPC and store them in an s3 bucket"
	case "aws.vpc.flow_logs_retention_days":
		return "Set the number of retention days for flow logs"
	case "cloud_provider.container_registry.tags":
		return "Add additional tags on the cluster dedicated registry"
	case "database.mongodb.allowed_cidrs", "database.mysql.allowed_cidrs", "database.postgresql.allowed_cidrs", "database.redis.allowed_cidrs":
		return "List of allowed CIDRS"
	case "database.mongodb.deny_public_access":
		return "Deny public access to all MongoDB databases"
	case "database.mysql.deny_public_access":
		return "Deny public access to all MySQL databases"
	case "database.postgresql.deny_public_access":
		return "Deny public access to all PostgreSQL databases"
	case "database.redis.deny_public_access":
		return "Deny public access to all Redis databases"
	case "load_balancer.size":
		return "Allows you to specify the load balancer size in front of your cluster"
	case "loki.log_retention_in_week":
		return "Maximum Kubernetes pods (containers/application/jobs/cronjob) retention logs in weeks"
	case "registry.image_retention_time":
		return "Allows you to specify an amount in seconds after which images in the default registry are deleted"
	case "pleco.resources_ttl":
		return "Deprecated"
	}

	return noDescription
}

func autoScalingSettingsDescription(setting string) string {
	switch setting {
	case "hpa.cpu.average_utilization_percent":
		return "CPU usage autoscaling trigger value"
	}

	return noDescription
}

func deploymentSettingsDescription(setting string) string {
	switch setting {
	case "build.timeout_max_sec":
		return "Interval in seconds after which the application build times out"
	case "deployment.custom_domain_check_enabled":
		return "Allows you to specify the IAM group name associated to the Qovery user"
	case "deployment.delay_start_time_sec":
		return "Deprecated"
	case "deployment.termination_grace_period_seconds":
		return "Time in seconds the application is supposed to stop at maximum"
	}

	return noDescription
}

func jobSettingsDescription(setting string) string {
	switch setting {
	case "cronjob.concurrency_policy":
		return "Define if it is allowed to start another instance of the same job if the previous execution didn't finish yet"
	case "cronjob.failed_job_history_limit":
		return "Define the maximum number of failed job executions that should be returned in the job execution history"
	case "cronjob.success_job_history_limit":
		return "Define the maximum number of succeeded job executions that should be returned in the job execution history"
	case "job.delete_ttl_seconds_after_finished":
		return "Kubernetes will automatically cleanup completed jobs after the ttl"
	}

	return noDescription
}

func networkSettingsDescription(setting string) string {
	switch setting {
	case "network.ingress.cors_allow_headers":
		return "Specify which set of headers can be present in the client request"
	case "network.ingress.cors_allow_methods":
		return "Specify which set of methods can be used for the client request"
	case "network.ingress.cors_allow_origin":
		return "Specify which origin(s) can access a resource"
	case "network.ingress.enable_cors":
		return "Enable Cross-Origin Resource Sharing"
	case "network.ingress.enable_sticky_session":
		return "Enable Sticky session"
	case "network.ingress.keepalive_time_seconds":
		return "Limits the maximum time in seconds during which requests can be processed through one keepalive connection"
	case "network.ingress.keepalive_timeout_seconds":
		return "Sets a timeout in seconds during which an idle keepalive connection to an upstream server will stay open"
	case "network.ingress.proxy_body_size_mb":
		return "Set in megabytes a maximum size for resources that can be downloaded from your server"
	case "network.ingress.proxy_buffer_size_kb":
		return "Set in kilobytes a header buffer size used while reading the response header from upstream"
	case "network.ingress.proxy_connect_timeout_seconds":
		return "Defines a timeout in seconds for establishing a connection with a proxied server"
	case "network.ingress.proxy_read_timeout_seconds":
		return "Defines a timeout in seconds for reading a response from the proxied server"
	case "network.ingress.proxy_send_timeout_seconds":
		return "Sets a timeout in seconds for transmitting a request to the proxied server"
	case "network.ingress.send_timeout_seconds":
		return "Sets a timeout in seconds for transmitting a response to the client"
	case "network.ingress.whitelist_source_range":
		return "Specify which IP ranges are allowed to access your application (comma-separated list of CIDRs)"
	case "network.ingress.denylist_source_range":
		return "Specify which IP ranges are not allowed to access your application (comma-separated list of CIDRs)"
	case "network.ingress.basic_auth_env_var":
		return "Set the name of an environment variable to use as a basic authentication (login:crypted_password) from htpasswd command"
	}

	return noDescription
}

func probeSettingsDescription(setting string) string {
	switch setting {
	case "liveness_probe.type", "readiness_probe.type":
		return "Specify the type of probe: TCP, HTTP or NONE"
	case "liveness_probe.http_get.path", "readiness_probe.http_get.path":
		return "Path to access on the HTTP/HTTPS server to perform the health check"
	case "liveness_probe.initial_delay_seconds", "readiness_probe.initial_delay_seconds":
		return "Interval in seconds between the container start and the first check"
	case "liveness_probe.period_seconds", "readiness_probe.period_seconds":
		return "Interval in seconds between each check"
	case "liveness_probe.timeout_seconds", "readiness_probe.timeout_seconds":
		return "Interval in seconds after the probe times out"
	case "liveness_probe.success_threshold", "readiness_probe.success_threshold":
		return "Specify how many consecutive successes are needed to be considered successful after having failed"
	case "liveness_probe.failure_threshold", "readiness_probe.failure_threshold":
		return "Specify how many consecutive failures are needed to be considered failed after having succeeded"
	}

	return noDescription
}

func securitySettingsDescription(setting string) string {
	switch setting {
	case "security.service_account_name", "readiness_probe.type":
		return "Set an existing Kubernetes service account name"
	}

	return noDescription
}

func settingsDescription(setting string) string {
	if clusterSettingsDescription(setting) != noDescription {
		return clusterSettingsDescription(setting)
	}

	if autoScalingSettingsDescription(setting) != noDescription {
		return autoScalingSettingsDescription(setting)
	}

	if deploymentSettingsDescription(setting) != noDescription {
		return deploymentSettingsDescription(setting)
	}

	if jobSettingsDescription(setting) != noDescription {
		return jobSettingsDescription(setting)
	}

	if networkSettingsDescription(setting) != noDescription {
		return networkSettingsDescription(setting)
	}

	if probeSettingsDescription(setting) != noDescription {
		return probeSettingsDescription(setting)
	}

	if securitySettingsDescription(setting) != noDescription {
		return securitySettingsDescription(setting)
	}

	return noDescription
}

func GetApplicationSettingsDefault() map[string]AdvSettingAttr {
	return advSettingsFromStruct(qovery.ApplicationAdvancedSettings{})
}

func GetContainerSettingsDefault() map[string]AdvSettingAttr {
	return advSettingsFromStruct(qovery.ContainerAdvancedSettings{})
}

func GetJobSettingsDefault() map[string]AdvSettingAttr {
	return advSettingsFromStruct(qovery.JobAdvancedSettings{})
}

func GetClusterSettingsDefault() map[string]AdvSettingAttr {
	return advSettingsFromStruct(qovery.ClusterAdvancedSettings{})
}
