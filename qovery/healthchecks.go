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
		MarkdownDescription: "Readiness and liveness probes of the service. `healthchecks = {}` sets none; production workloads should set at least one.",
		Required:            required,
		Optional:            !required,
		Attributes: map[string]schema.Attribute{
			"readiness_probe": schema.SingleNestedAttribute{
				MarkdownDescription: "Probe that decides when the service receives traffic: while it fails, the service is out of the load balancer.",
				Optional:            true,
				Attributes:          probeSchemaAttributes(),
			},
			"liveness_probe": schema.SingleNestedAttribute{
				MarkdownDescription: "Probe that decides whether the service works: when it fails, the container restarts.",
				Optional:            true,
				Attributes:          probeSchemaAttributes(),
			},
		},
	}
}

func probeSchemaAttributes() map[string]schema.Attribute {
	const probePortDescription = "Port to check."
	return map[string]schema.Attribute{
		"initial_delay_seconds": schema.Int64Attribute{
			MarkdownDescription: "Seconds to wait after the container starts before the first probe.",
			Required:            true,
		},
		"period_seconds": schema.Int64Attribute{
			MarkdownDescription: "Seconds between two probes.",
			Required:            true,
		},
		"timeout_seconds": schema.Int64Attribute{
			MarkdownDescription: "Seconds after which a probe that has not answered fails.",
			Required:            true,
		},
		"success_threshold": schema.Int64Attribute{
			MarkdownDescription: "Consecutive successes after a failure for the probe to pass.",
			Required:            true,
		},
		"failure_threshold": schema.Int64Attribute{
			MarkdownDescription: "Consecutive failures for the probe to fail.",
			Required:            true,
		},
		"type": schema.SingleNestedAttribute{
			MarkdownDescription: "Check the probe runs: set exactly one of `tcp`, `http`, `grpc` and `exec`.",
			Required:            true,
			Attributes: map[string]schema.Attribute{
				"tcp": schema.SingleNestedAttribute{
					MarkdownDescription: "Opens a TCP connection to `port`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							MarkdownDescription: probePortDescription,
							Required:            true,
						},
						"host": schema.StringAttribute{
							MarkdownDescription: "Host to connect to. Defaults to the pod IP.",
							Optional:            true,
						},
					},
				},
				"http": schema.SingleNestedAttribute{
					MarkdownDescription: "Sends an HTTP GET request to `port` and passes on a status code from 200 to 399.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							MarkdownDescription: probePortDescription,
							Required:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path of the request, for example `/health`. Defaults to `/`.",
							Optional:            true,
						},
						"scheme": schema.StringAttribute{
							MarkdownDescription: "Scheme of the request: `HTTP` or `HTTPS`.",
							Required:            true,
						},
					},
				},
				"grpc": schema.SingleNestedAttribute{
					MarkdownDescription: "Calls the [gRPC health checking protocol](https://kubernetes.io/blog/2018/10/01/health-checking-grpc-servers-on-kubernetes/#introducing-grpc-health-probe) on `port`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"port": schema.Int64Attribute{
							MarkdownDescription: probePortDescription,
							Required:            true,
						},
						"service": schema.StringAttribute{
							MarkdownDescription: "gRPC service to check. Defaults to the health of the whole server.",
							Optional:            true,
						},
					},
				},
				"exec": schema.SingleNestedAttribute{
					MarkdownDescription: "Runs a command in the container, and passes when it exits with `0`.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"command": schema.ListAttribute{
							MarkdownDescription: "Command and its arguments, for example `[\"cat\", \"/tmp/healthy\"]`.",
							Required:            true,
							ElementType:         types.StringType,
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
	readinessProbe := qovery.NewNullableProbe(nil)
	if h.ReadinessProbe != nil {
		readinessProbe = h.ReadinessProbe.toProbeRequest()
	}
	livenessProbe := qovery.NewNullableProbe(nil)
	if h.LivenessProbe != nil {
		livenessProbe = h.LivenessProbe.toProbeRequest()
	}
	return qovery.Healthcheck{
		ReadinessProbe: *readinessProbe,
		LivenessProbe:  *livenessProbe,
	}
}

func convertProbeResponseToDomain(probe *qovery.NullableProbe) *Probe {
	p := probe.Get()
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
