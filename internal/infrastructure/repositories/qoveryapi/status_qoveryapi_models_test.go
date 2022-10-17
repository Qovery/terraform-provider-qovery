package qoveryapi

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
)

func TestNewDomainStatusFromQovery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		Status        *qovery.Status
		ExpectedError error
	}{
		{
			TestName:      "fail_with_nil_container",
			ExpectedError: status.ErrNilStatus,
		},
		{
			TestName: "success",
			Status: &qovery.Status{
				Id:                      gofakeit.UUID(),
				ServiceDeploymentStatus: qovery.SERVICEDEPLOYMENTSTATUSENUM_UP_TO_DATE,
				State:                   qovery.STATEENUM_RUNNING,
				LastDeploymentDate:      pointer.ToTime(gofakeit.Date()),
				Message:                 pointer.ToString(gofakeit.Word()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			st, err := newDomainStatusFromQovery(tc.Status)
			if tc.ExpectedError != nil {
				assert.ErrorContains(t, err, tc.ExpectedError.Error())
				assert.Nil(t, st)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, st)
			assert.Equal(t, tc.Status.Id, st.ID.String())
			assert.Equal(t, string(tc.Status.ServiceDeploymentStatus), st.ServiceDeploymentStatus.String())
			assert.Equal(t, string(tc.Status.State), st.State.String())
			assert.Equal(t, tc.Status.LastDeploymentDate, st.LastDeploymentDate)
			assert.Equal(t, tc.Status.Message, st.Message)
		})
	}
}
