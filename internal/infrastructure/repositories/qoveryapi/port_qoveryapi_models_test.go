package qoveryapi

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
)

func TestNewDomainPortsFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Ports         []qovery.ServicePort
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Ports: []qovery.ServicePort{
				{
					Id:                 gofakeit.UUID(),
					InternalPort:       5000,
					Protocol:           qovery.PORTPROTOCOLENUM_HTTP,
					PubliclyAccessible: gofakeit.Bool(),
				},
				{
					Id:                 gofakeit.UUID(),
					InternalPort:       5000,
					Protocol:           qovery.PORTPROTOCOLENUM_HTTP,
					PubliclyAccessible: gofakeit.Bool(),
					ExternalPort:       pointer.ToInt32(5001),
					Name:               new(gofakeit.Name()),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			ss, err := newDomainPortsFromQovery(tc.Ports)
			assert.NoError(t, err)
			assert.Len(t, tc.Ports, len(ss))

			for idx, s := range ss {
				assert.True(t, s.IsValid())
				assert.Equal(t, tc.Ports[idx].Name, s.Name)
				assert.Equal(t, tc.Ports[idx].ExternalPort, s.ExternalPort)
				assert.Equal(t, tc.Ports[idx].InternalPort, s.InternalPort)
				assert.Equal(t, tc.Ports[idx].PubliclyAccessible, s.PubliclyAccessible)
				assert.Equal(t, string(tc.Ports[idx].Protocol), s.Protocol.String())
			}
		})
	}
}

func TestNewQoveryPortsRequestFromDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Ports         []port.UpsertRequest
		ExpectedError error
	}{
		{
			TestName: "success_with_nil_container",
		},
		{
			TestName: "success",
			Ports: []port.UpsertRequest{
				{
					PubliclyAccessible: gofakeit.Bool(),
					InternalPort:       5000,
				},
				{
					Name:               new(gofakeit.Name()),
					Protocol:           new(port.ProtocolHTTP.String()),
					PubliclyAccessible: gofakeit.Bool(),
					InternalPort:       5000,
					ExternalPort:       pointer.ToInt32(5001),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			ss, err := newQoveryPortsRequestFromDomain(tc.Ports)
			assert.NoError(t, err)
			assert.Len(t, tc.Ports, len(ss))

			for idx, s := range ss {
				assert.Equal(t, tc.Ports[idx].Name, s.Name)
				assert.Equal(t, tc.Ports[idx].ExternalPort, s.ExternalPort)
				assert.Equal(t, tc.Ports[idx].InternalPort, s.InternalPort)
				assert.Equal(t, tc.Ports[idx].PubliclyAccessible, s.PubliclyAccessible)
				if tc.Ports[idx].Protocol != nil {
					require.NotNil(t, s.Protocol)
					assert.Equal(t, *tc.Ports[idx].Protocol, string(*s.Protocol))
				}
			}
		})
	}
}
