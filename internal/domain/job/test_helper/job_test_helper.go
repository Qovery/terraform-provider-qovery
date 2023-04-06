package test_helper

import (
	"github.com/google/uuid"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	port_helper "github.com/qovery/terraform-provider-qovery/internal/domain/port/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	secrets_helper "github.com/qovery/terraform-provider-qovery/internal/domain/secret/test_helper"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
	variables_helper "github.com/qovery/terraform-provider-qovery/internal/domain/variable/test_helper"
)

var (
	DefaultJobID                                     = uuid.New()
	DefaultJobEnvironmentID                          = uuid.New()
	DefaultJobName                                   = "MyJobName-" + uuid.New().String()
	DefaultInvalidJobName                            = ""
	DefaultJobCPU                         int32      = 100
	DefaultInvalidJobCPU                  int32      = 8
	DefaultJobMemory                      int32      = 250
	DefaultInvalidJobMemory               int32      = 0
	DefaultJobMaxNbRestart                uint32     = 0
	DefaultJobMaxDurationSeconds          uint32     = 300
	DefaultJobAutoPreview                            = false
	DefaultJobPort                        *port.Port = &port_helper.DefaultValidPort
	DefaultInvalidJobPort                 *port.Port = &port_helper.DefaultInvalidPort
	DefaultJobEnvironmentVariables                   = []variable.Variable{variables_helper.DefaultValidVariable}
	DefaultInvalidJobEnvironmentVariables            = []variable.Variable{variables_helper.DefaultInvalidVariable}
	DefaultJobEnvironmentSecrets                     = []secret.Secret{secrets_helper.DefaultValidSecret}
	DefaultInvalidJobEnvironmentSecrets              = []secret.Secret{secrets_helper.DefaultInvalidSecret}
	DefaultJobDeploymentStageID                      = uuid.New()
)
