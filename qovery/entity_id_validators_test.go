//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/qovery/terraform-provider-qovery/qovery/validators"
)

const (
	awsAccessKeyID = "AWS access key id"
	awsSubnetIDs   = "AWS subnet ids"
	gcpProjectID   = "GCP project id"
	scalewayID     = "Scaleway project or organization id"
)

// nonUUIDIDAttributes lists the configurable attributes named *_id or *_ids that do not reference a
// Qovery entity, so they carry no UUID validator.
var nonUUIDIDAttributes = map[string]string{
	"qovery_aws_credentials.access_key_id":                                                             awsAccessKeyID,
	"qovery_eks_anywhere_vsphere_credentials.access_key_id":                                            awsAccessKeyID,
	"qovery_container_registry.config.access_key_id":                                                   awsAccessKeyID,
	"qovery_helm_repository.config.access_key_id":                                                      awsAccessKeyID,
	"qovery_cluster_dns_provider.route53.credentials.aws_access_key_id":                                awsAccessKeyID,
	"qovery_cluster_dns_provider.route53.hosted_zone_id":                                               "Route 53 hosted zone id",
	"qovery_cluster.features.existing_vpc.aws_vpc_eks_id":                                              "AWS VPC id",
	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_a_ids":                               awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_b_ids":                               awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.documentdb_subnets_zone_c_ids":                               awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_karpenter_fargate_subnets_zone_a_ids":                    awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_karpenter_fargate_subnets_zone_b_ids":                    awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_karpenter_fargate_subnets_zone_c_ids":                    awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_subnets_zone_a_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_subnets_zone_b_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.eks_subnets_zone_c_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_a_ids":                              awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_b_ids":                              awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.elasticache_subnets_zone_c_ids":                              awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_a_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_b_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.existing_vpc.rds_subnets_zone_c_ids":                                      awsSubnetIDs,
	"qovery_cluster.features.gcp_existing_vpc.vpc_project_id":                                          gcpProjectID,
	"qovery_cluster.secret_manager_accesses.endpoint.project_id":                                       gcpProjectID,
	"qovery_container_registry.config.project_id":                                                      gcpProjectID,
	"qovery_cluster.infrastructure_charts_parameters.eks_anywhere_parameters.git_repository.commit_id": "git commit SHA",
	"qovery_container_registry.config.scaleway_project_id":                                             scalewayID,
	"qovery_helm_repository.config.scaleway_project_id":                                                scalewayID,
	"qovery_scaleway_credentials.scaleway_project_id":                                                  scalewayID,
	"qovery_scaleway_credentials.scaleway_organization_id":                                             scalewayID,
}

// uuidAttributesWithoutIDSuffix lists the configurable attributes that reference a Qovery entity
// although their name does not end in _id.
var uuidAttributesWithoutIDSuffix = map[string]bool{
	"qovery_deployment_stage.is_after":  true,
	"qovery_deployment_stage.is_before": true,
}

// Asserts that every configurable attribute named *_id or *_ids either validates a UUID at plan
// time or is listed in nonUUIDIDAttributes, so a new id attribute cannot go unclassified.
func TestEntityIDAttributes_HaveUUIDValidator(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	var missing, unexpected []string
	for _, tc := range allResourceCases(t) {
		walkAttributes("", schemaFor(t, tc.resource).Attributes, func(path string, attr schema.Attribute) {
			if !attr.IsRequired() && !attr.IsOptional() {
				return
			}
			key := tc.typeName + "." + path
			name := path[strings.LastIndex(path, ".")+1:]
			isIDName := strings.HasSuffix(name, "_id") || strings.HasSuffix(name, "_ids")
			if !isIDName && !uuidAttributesWithoutIDSuffix[key] {
				return
			}
			seen[key] = true

			_, excluded := nonUUIDIDAttributes[key]
			validated := hasUUIDValidator(attr)
			switch {
			case excluded && validated:
				unexpected = append(unexpected, key)
			case !excluded && !validated:
				missing = append(missing, key)
			}
		})
	}

	for key := range nonUUIDIDAttributes {
		if !seen[key] {
			unexpected = append(unexpected, key+" (stale: no such configurable attribute)")
		}
	}
	for key := range uuidAttributesWithoutIDSuffix {
		if !seen[key] {
			unexpected = append(unexpected, key+" (stale: no such configurable attribute)")
		}
	}

	sort.Strings(missing)
	sort.Strings(unexpected)
	if len(missing) > 0 {
		t.Errorf("%d attribute(s) reference a Qovery entity but have no validators.NewUUIDValidator(); add it, "+
			"or list the attribute in nonUUIDIDAttributes with the reason it is not a UUID:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
	if len(unexpected) > 0 {
		t.Errorf("nonUUIDIDAttributes or uuidAttributesWithoutIDSuffix is out of date:\n  %s", strings.Join(unexpected, "\n  "))
	}
}

func hasUUIDValidator(attr schema.Attribute) bool {
	switch a := attr.(type) {
	case schema.StringAttribute:
		for _, v := range a.Validators {
			if _, ok := v.(validators.UUIDValidator); ok {
				return true
			}
		}
	case schema.SetAttribute:
		for _, v := range a.Validators {
			if _, ok := v.(validators.UUIDValidator); ok {
				return true
			}
		}
	}
	return false
}
