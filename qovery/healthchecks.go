package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
)

type HealthChecks struct {
	ReadinessProbe *Probe `tfsdk:"readiness_probe"`
	LivenessProbe  *Probe `tfsdk:"liveness_probe"`
}

type ProbeType struct {
	Tcp  *ProbeTcp  `tfsdk:"tcp"`
	Http *ProbeHttp `tfsdk:"http"`
	Grpc *ProbeGrpc `tfsdk:"grpc"`
	Exec *ProbeExec `tfsdk:"exec"`
}

type Probe struct {
	InitialDelaySeconds types.Int64 `tfsdk:"initial_delay_seconds"`
	PeriodSeconds       types.Int64 `tfsdk:"period_seconds"`
	TimeoutSeconds      types.Int64 `tfsdk:"timeout_seconds"`
	SuccessThreshold    types.Int64 `tfsdk:"success_threshold"`
	FailureThreshold    types.Int64 `tfsdk:"failure_threshold"`
	Type                ProbeType   `tfsdk:"type"`
}
type ProbeTcp struct {
	Port types.Int64  `tfsdk:"port"`
	Host types.String `tfsdk:"host"`
}

type ProbeHttp struct {
	Port   types.Int64  `tfsdk:"port"`
	Path   types.String `tfsdk:"path"`
	Scheme types.String `tfsdk:"scheme"`
}

type ProbeGrpc struct {
	Port    types.Int64  `tfsdk:"port"`
	Service types.String `tfsdk:"service"`
}

type ProbeExec struct {
	command types.List `tfsdk:"command"`
}

func healthchecksSchemaAttributes(required bool) tfsdk.Attribute {
	return tfsdk.Attribute{
		Description: "Configuration for the healthchecks that are going to be executed against your service",
		Required:    required,
		Optional:    !required,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"readiness_probe": {
				Description: "Configuration for the readiness probe, in order to know when your service is ready to receive traffic. Failing the probe means your service will stop receiving traffic.",
				Attributes:  tfsdk.SingleNestedAttributes(probeSchemaAttributes()),
				Optional:    true,
			},
			"liveness_probe": {
				Description: "Configuration for the liveness probe, in order to know when your service is working correctly. Failing the probe means your service being killed/ask to be restarted.",
				Attributes:  tfsdk.SingleNestedAttributes(probeSchemaAttributes()),
				Optional:    true,
			},
		}),
	}
}

func probeSchemaAttributes() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"initial_delay_seconds": {
			Description: "Number of seconds to wait before the first execution of the probe to be trigerred",
			Type:        types.Int64Type,
			Required:    true,
		},
		"period_seconds": {
			Description: "Number of seconds before each execution of the probe",
			Type:        types.Int64Type,
			Required:    true,
		},
		"timeout_seconds": {
			Description: "Number of seconds within which the check need to respond before declaring it as a failure",
			Type:        types.Int64Type,
			Required:    true,
		},
		"success_threshold": {
			Description: "Number of time the probe should success before declaring a failed probe as ok again",
			Type:        types.Int64Type,
			Required:    true,
		},
		"failure_threshold": {
			Description: "Number of time the an ok probe should fail before declaring it as failed",
			Type:        types.Int64Type,
			Required:    true,
		},
		"type": {
			Description: "Kind of check to run for this probe. There can only be one configured at a time",
			Required:    true,
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"tcp": {
					Description: "Check that the given port accepting connection",
					Optional:    true,
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"port": {
							Description: "The port number to try to connect to",
							Required:    true,
							Type:        types.Int64Type,
						},
						"host": {
							Description: "Optional. If the host need to be different than localhost/pod ip",
							Type:        types.StringType,
							Optional:    true,
							Computed:    true,
						},
					}),
				},
				"http": {
					Description: "Check that the given port respond to HTTP call (should return a 2xx response code)",
					Optional:    true,
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"port": {
							Description: "The port number to try to connect to",
							Required:    true,
							Type:        types.Int64Type,
						},
						"path": {
							Description: "The path that the HTTP GET request. By default it is `/`",
							Type:        types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"scheme": {
							Description: "if the HTTP GET request should be done in HTTP or HTTPS. Default is HTTP",
							Type:        types.StringType,
							Optional:    true,
							Computed:    true,
						},
					}),
				},
				"grpc": {
					Description: "Check that the given port respond to GRPC call",
					Optional:    true,
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"port": {
							Description: "The port number to try to connect to",
							Required:    true,
							Type:        types.Int64Type,
						},
						"service": {
							Description: "The grpc service to connect to. It needs to implement grpc health protocol. https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe",
							Type:        types.StringType,
							Optional:    true,
						},
					}),
				},
				"exec": {
					Description: "Check that the given command return an exit 0. Binary should be present in the image",
					Optional:    true,
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"command": {
							Description: "The command and its arguments to exec",
							Required:    true,
							Type: types.ListType{
								ElemType: types.StringType,
							},
						},
					}),
				},
			}),
		},
	}
}

func (p *ProbeTcp) toProbeTcpRequest() qovery.NullableProbeTypeTcp {
	if p == nil {
		return *qovery.NewNullableProbeTypeTcp(nil)
	}

	return *qovery.NewNullableProbeTypeTcp(&qovery.ProbeTypeTcp{
		Port: ToInt32Pointer(p.Port),
		Host: ToNullableString(p.Host),
	})
}

func (p *ProbeHttp) toProbeHttpRequest() qovery.NullableProbeTypeHttp {
	if p == nil {
		return qovery.NullableProbeTypeHttp{}
	}

	return *qovery.NewNullableProbeTypeHttp(&qovery.ProbeTypeHttp{
		Port:   ToInt32Pointer(p.Port),
		Path:   ToStringPointer(p.Path),
		Scheme: ToStringPointer(p.Scheme),
	})
}

func (p *ProbeGrpc) toProbeGrpcRequest() qovery.NullableProbeTypeGrpc {
	if p == nil {
		return qovery.NullableProbeTypeGrpc{}
	}

	return *qovery.NewNullableProbeTypeGrpc(&qovery.ProbeTypeGrpc{
		Port:    ToInt32Pointer(p.Port),
		Service: ToNullableString(p.Service),
	})
}

func (p *ProbeExec) toProbeExecRequest() qovery.NullableProbeTypeExec {
	if p == nil {
		return qovery.NullableProbeTypeExec{}
	}

	return *qovery.NewNullableProbeTypeExec(&qovery.ProbeTypeExec{
		Command: ToStringArray(p.command),
	})
}

func (p *Probe) toProbeRequest() *qovery.Probe {
	if p == nil {
		return nil
	}

	return &qovery.Probe{
		InitialDelaySeconds: ToInt32Pointer(p.InitialDelaySeconds),
		PeriodSeconds:       ToInt32Pointer(p.PeriodSeconds),
		TimeoutSeconds:      ToInt32Pointer(p.TimeoutSeconds),
		SuccessThreshold:    ToInt32Pointer(p.SuccessThreshold),
		FailureThreshold:    ToInt32Pointer(p.FailureThreshold),
		Type: &qovery.ProbeType{
			Exec: p.Type.Exec.toProbeExecRequest(),
			Tcp:  p.Type.Tcp.toProbeTcpRequest(),
			Http: p.Type.Http.toProbeHttpRequest(),
			Grpc: p.Type.Grpc.toProbeGrpcRequest(),
		},
	}
}

func (h HealthChecks) toHealthchecksRequest() qovery.Healthcheck {
	return qovery.Healthcheck{
		ReadinessProbe: h.ReadinessProbe.toProbeRequest(),
		LivenessProbe:  h.LivenessProbe.toProbeRequest(),
	}
}

func convertProbeResponseToDomain(p *qovery.Probe) *Probe {
	if p == nil {
		return nil
	}

	var tcp *ProbeTcp
	if p.Type.Tcp.Get() != nil {
		tcp = &ProbeTcp{
			Port: FromInt32Pointer(p.Type.Tcp.Get().Port),
			Host: FromStringPointer(p.Type.Tcp.Get().Host.Get()),
		}
	}

	var http *ProbeHttp
	if p.Type.Http.Get() != nil {
		http = &ProbeHttp{
			Port:   FromInt32Pointer(p.Type.Http.Get().Port),
			Path:   FromStringPointer(p.Type.Http.Get().Path),
			Scheme: FromStringPointer(p.Type.Http.Get().Scheme),
		}
	}

	var grpc *ProbeGrpc
	if p.Type.Grpc.Get() != nil {
		grpc = &ProbeGrpc{
			Port:    FromInt32Pointer(p.Type.Grpc.Get().Port),
			Service: FromNullableString(p.Type.Grpc.Get().Service),
		}
	}

	var exec *ProbeExec
	if p.Type.Exec.Get() != nil {
		exec = &ProbeExec{
			command: FromStringArray(p.Type.Exec.Get().Command),
		}
	}

	return &Probe{
		InitialDelaySeconds: FromInt32Pointer(p.InitialDelaySeconds),
		PeriodSeconds:       FromInt32Pointer(p.PeriodSeconds),
		TimeoutSeconds:      FromInt32Pointer(p.TimeoutSeconds),
		SuccessThreshold:    FromInt32Pointer(p.SuccessThreshold),
		FailureThreshold:    FromInt32Pointer(p.FailureThreshold),
		Type: ProbeType{
			Tcp:  tcp,
			Http: http,
			Grpc: grpc,
			Exec: exec,
		},
	}
}

func convertHealthchecksResponseToDomain(r qovery.Healthcheck) HealthChecks {
	return HealthChecks{
		ReadinessProbe: convertProbeResponseToDomain(r.ReadinessProbe),
		LivenessProbe:  convertProbeResponseToDomain(r.LivenessProbe),
	}
}
