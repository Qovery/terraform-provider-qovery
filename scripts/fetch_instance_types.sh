#!/bin/bash
set -o pipefail
curl -f -s -H "Accept: application/json" -H "Authorization: Token $QOVERY_API_TOKEN" https://api.qovery.com/aws/instanceType | jq "[.results[] | .type] | sort" | jq -rc tostring | tr -d '\n' > qovery/data/cluster_instance_types/aws.json
curl -f -s -H "Accept: application/json" -H "Authorization: Token $QOVERY_API_TOKEN" https://api.qovery.com/digitalOcean/instanceType | jq "[.results[] | .type] | sort" | jq -rc tostring | tr -d '\n' > qovery/data/cluster_instance_types/digital_ocean.json
curl -f -s -H "Accept: application/json" -H "Authorization: Token $QOVERY_API_TOKEN" https://api.qovery.com/scaleway/instanceType | jq "[.results[] | .type] | sort" | jq -rc tostring | tr -d '\n' > qovery/data/cluster_instance_types/scaleway.json