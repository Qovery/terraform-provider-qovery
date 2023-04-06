package validators

import "fmt"

type ClusterSettingsValidator struct {
	AdvSettings map[string]interface{}
}

func (c ClusterSettingsValidator) Validate() error {

	switch (c.AdvSettings["aws.cloudwatch.eks_logs_retention_days"]).(int64) {
	case 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 2192, 2557, 2922, 3288, 3653:
		break
	default:
		return fmt.Errorf("aws.cloudwatch.eks_logs_retention_days possible values: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 2192, 2557, 2922, 3288, 3653")
	}

	return nil
}
