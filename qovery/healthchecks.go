package qovery

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	Command types.List `tfsdk:"command"`
}

func healthchecksSchemaAttributes(required bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Configuration for the healthchecks that are going to be executed against your service",
		Required:    required,
		Optional:    !required,
		Attributes: map[string]schema.Attribute{
			"readiness_probe": schema.SingleNestedAttribute{
				Description: "Configuration for the readiness probe, in order to know when your service is ready to receive traffic. Failing the probe means your service will stop receiving traffic.",
				Optional:    true,
				Attributes:  probeSchemaAttributes(),
			},
			"liveness_probe": schema.SingleNestedAttribute{
				Description: "Configuration for the liveness probe, in order to know when your service is working correctly. Failing the probe means your service being killed/ask to be restarted.",
				Optional:    true,
				Attributes:  probeSchemaAttributes(),
			},
		},
	}
}

func probeSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"initial_delay_seconds": schema.Int64Attribute{
			Description: "Number of seconds to wait before the first execution of the probe to be trigerred",
			Required:    true,
		},
		"period_seconds": schema.Int64Attribute{
			Description: "Number of seconds before each execution of the probe",
			Required:    true,
		},
		"timeout_seconds": schema.Int64Attribute{
			Description: "Number of seconds within which the check need to respond before declaring it as a failure",
			Required:    true,
		},
		"success_threshold": schema.Int64Attribute{
			Description: "Number of time the probe should success before declaring a failed probe as ok again",
			Required:    true,
		},
		"failure_threshold": schema.Int64Attribute{
			Description: "Number of time the an ok probe should fail before declaring it as failed",
			Required:    true,
		},
		"type": schema.SingleNestedAttribute{
			Description: "Kind of check to run for this probe. There can only be one configured at a time",
			Required:    true,
			Attributes: map[string]schema.Attribute{
				"tcp": schema.SingleNestedAttribute{
					Description: "Check that the given port accepting connection",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							Description: "The port number to try to connect to",
							Required:    true,
						},
						"host": schema.StringAttribute{
							Description: "Optional. If the host need to be different than localhost/pod ip",
							Optional:    true,
							Computed:    true,
						},
					},
				},
				"http": schema.SingleNestedAttribute{
					Description: "Check that the given port respond to HTTP call (should return a 2xx response code)",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							Description: "The port number to try to connect to",
							Required:    true,
						},
						"path": schema.StringAttribute{
							Description: "The path that the HTTP GET request. By default it is `/`",
							Optional:    true,
							Computed:    true,
						},
						"scheme": schema.StringAttribute{
							Description: "if the HTTP GET request should be done in HTTP or HTTPS. Default is HTTP",
							Optional:    true,
							Computed:    true,
						},
					},
				},
				"grpc": schema.SingleNestedAttribute{
					Description: "Check that the given port respond to GRPC call",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							Description: "The port number to try to connect to",
							Required:    true,
						},
						"service": schema.StringAttribute{
							Description: "The grpc service to connect to. It needs to implement grpc health protocol. https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe",
							Optional:    true,
						},
					},
				},
				"exec": schema.SingleNestedAttribute{
					Description: "Check that the given command return an exit 0. Binary should be present in the image",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"command": schema.ListAttribute{
							Description: "The command and its arguments to exec",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
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
		Command: ToStringArray(p.Command),
	})
}

func (p *Probe) toProbeRequest() *qovery.NullableProbe {
	if p == nil {
		return nil
	}

	probe := qovery.Probe{
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
	return qovery.NewNullableProbe(&probe)
}

func (h HealthChecks) toHealthchecksRequest() qovery.Healthcheck {
	var readinessProbe = qovery.NewNullableProbe(nil)
	if h.ReadinessProbe != nil {
		readinessProbe = h.ReadinessProbe.toProbeRequest()
	}
	var livenessProbe = qovery.NewNullableProbe(nil)
	if h.LivenessProbe != nil {
		livenessProbe = h.LivenessProbe.toProbeRequest()
	}
	return qovery.Healthcheck{
		ReadinessProbe: *readinessProbe,
		LivenessProbe:  *livenessProbe,
	}
}

func convertProbeResponseToDomain(probe *qovery.NullableProbe) *Probe {
	var p = probe.Get()
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
			Command: FromStringArray(p.Type.Exec.Get().Command),
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
		ReadinessProbe: convertProbeResponseToDomain(&r.ReadinessProbe),
		LivenessProbe:  convertProbeResponseToDomain(&r.LivenessProbe),
	}
}
