//go:build unit || !integration

package qovery

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

func strPtr(s string) *string { return &s }

func int32Ptr(i int32) *int32 { return &i }

func boolPtr(b bool) *bool { return &b }

func TestConvertResponseToApplicationPorts(t *testing.T) {
	t.Parallel()

	apiPort := qovery.ServicePort{
		Id:                 "id-api",
		Name:               strPtr("api"),
		InternalPort:       4000,
		ExternalPort:       int32Ptr(443),
		Protocol:           qovery.PORTPROTOCOLENUM_HTTP,
		PubliclyAccessible: true,
		IsDefault:          boolPtr(true),
	}
	sqlPort := qovery.ServicePort{
		Id:                 "id-sql",
		Name:               strPtr("sql"),
		InternalPort:       15432,
		ExternalPort:       int32Ptr(15432),
		Protocol:           qovery.PORTPROTOCOLENUM_TCP,
		PubliclyAccessible: false,
		IsDefault:          boolPtr(false),
	}

	testCases := []struct {
		TestName     string
		InitialState []ApplicationPort
		APIPorts     []qovery.ServicePort
		ExpectedIDs  []string
	}{
		{
			TestName:     "empty_state_sorts_by_internal_port",
			InitialState: nil,
			APIPorts:     []qovery.ServicePort{sqlPort, apiPort},
			ExpectedIDs:  []string{"id-api", "id-sql"}, // 4000 < 15432
		},
		{
			TestName: "state_with_ids_preserves_order",
			InitialState: []ApplicationPort{
				{Id: FromString("id-sql"), Name: FromString("sql")},
				{Id: FromString("id-api"), Name: FromString("api")},
			},
			APIPorts:    []qovery.ServicePort{apiPort, sqlPort},
			ExpectedIDs: []string{"id-sql", "id-api"},
		},
		{
			TestName: "port_renamed_matches_by_id",
			InitialState: []ApplicationPort{
				{Id: FromString("id-api"), Name: FromString("old-name")},
				{Id: FromString("id-sql"), Name: FromString("sql")},
			},
			APIPorts:    []qovery.ServicePort{apiPort, sqlPort},
			ExpectedIDs: []string{"id-api", "id-sql"},
		},
		{
			TestName: "new_port_appended_after_matched",
			InitialState: []ApplicationPort{
				{Id: FromString("id-sql"), Name: FromString("sql")},
			},
			APIPorts:    []qovery.ServicePort{apiPort, sqlPort},
			ExpectedIDs: []string{"id-sql", "id-api"},
		},
		{
			TestName: "deleted_port_not_in_result",
			InitialState: []ApplicationPort{
				{Id: FromString("id-sql"), Name: FromString("sql")},
				{Id: FromString("id-gone"), Name: FromString("gone")},
			},
			APIPorts:    []qovery.ServicePort{sqlPort},
			ExpectedIDs: []string{"id-sql"},
		},
		{
			TestName:     "nil_ports_nil_state_returns_nil",
			InitialState: nil,
			APIPorts:     []qovery.ServicePort{},
			ExpectedIDs:  nil,
		},
		{
			TestName: "name_fallback_when_id_not_matched",
			InitialState: []ApplicationPort{
				{Id: FromString(""), Name: FromString("api")},
				{Id: FromString(""), Name: FromString("sql")},
			},
			APIPorts:    []qovery.ServicePort{sqlPort, apiPort},
			ExpectedIDs: []string{"id-api", "id-sql"},
		},
		{
			TestName: "console_changed_first_port_preserves_index",
			InitialState: []ApplicationPort{
				{Id: FromString("id-gone"), Name: FromString("api")},
				{Id: FromString("id-sql"), Name: FromString("sql")},
			},
			APIPorts: []qovery.ServicePort{
				sqlPort,
				{
					Id:                 "id-new",
					Name:               strPtr("p4200"),
					InternalPort:       4200,
					ExternalPort:       int32Ptr(443),
					Protocol:           qovery.PORTPROTOCOLENUM_HTTP,
					PubliclyAccessible: true,
					IsDefault:          boolPtr(true),
				},
			},
			// p4200 fills gap at index 0 (where api was), sql stays at index 1
			ExpectedIDs: []string{"id-new", "id-sql"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			result := convertResponseToApplicationPorts(tc.InitialState, tc.APIPorts)
			if tc.ExpectedIDs == nil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, len(tc.ExpectedIDs), len(result))
			for i, expectedID := range tc.ExpectedIDs {
				assert.Equal(t, expectedID, result[i].Id.ValueString(), "port at index %d", i)
			}
		})
	}
}

func TestConvertDomainPortsToPortList(t *testing.T) {
	t.Parallel()

	httpProtocol := port.ProtocolHTTP
	tcpProtocol := port.ProtocolTCP

	apiPort := port.Port{
		ID:                 uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
		Name:               strPtr("api"),
		InternalPort:       4000,
		ExternalPort:       int32Ptr(443),
		Protocol:           &httpProtocol,
		PubliclyAccessible: true,
		IsDefault:          true,
	}
	sqlPort := port.Port{
		ID:                 uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002"),
		Name:               strPtr("sql"),
		InternalPort:       15432,
		ExternalPort:       int32Ptr(15432),
		Protocol:           &tcpProtocol,
		PubliclyAccessible: false,
		IsDefault:          false,
	}

	apiID := apiPort.ID.String()
	sqlID := sqlPort.ID.String()

	portObjectType := types.ObjectType{AttrTypes: portAttrTypes}

	makeStateList := func(ports []Port) types.List {
		elements := make([]attr.Value, 0, len(ports))
		for _, p := range ports {
			elements = append(elements, p.toTerraformObject())
		}
		list, _ := types.ListValueFrom(context.Background(), portObjectType, elements)
		return list
	}

	testCases := []struct {
		TestName     string
		InitialState types.List
		DomainPorts  port.Ports
		ExpectedIDs  []string
	}{
		{
			TestName:     "null_state_sorts_by_internal_port",
			InitialState: types.ListNull(portObjectType),
			DomainPorts:  port.Ports{sqlPort, apiPort},
			ExpectedIDs:  []string{apiID, sqlID}, // 4000 < 15432
		},
		{
			TestName: "state_with_ids_preserves_order",
			InitialState: makeStateList([]Port{
				{Id: FromString(sqlID), Name: FromString("sql")},
				{Id: FromString(apiID), Name: FromString("api")},
			}),
			DomainPorts: port.Ports{apiPort, sqlPort},
			ExpectedIDs: []string{sqlID, apiID},
		},
		{
			TestName: "port_renamed_matches_by_id",
			InitialState: makeStateList([]Port{
				{Id: FromString(apiID), Name: FromString("old-name")},
				{Id: FromString(sqlID), Name: FromString("sql")},
			}),
			DomainPorts: port.Ports{apiPort, sqlPort},
			ExpectedIDs: []string{apiID, sqlID},
		},
		{
			TestName:     "null_state_empty_ports_returns_nil",
			InitialState: types.ListNull(portObjectType),
			DomainPorts:  port.Ports{},
			ExpectedIDs:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			result := convertDomainPortsToPortList(ctx, tc.InitialState, tc.DomainPorts)
			if tc.ExpectedIDs == nil {
				assert.Nil(t, PortList(result))
				return
			}
			assert.Equal(t, len(tc.ExpectedIDs), len(result))
			for i, expectedID := range tc.ExpectedIDs {
				assert.Equal(t, expectedID, result[i].Id.ValueString(), "port at index %d", i)
			}
		})
	}
}
