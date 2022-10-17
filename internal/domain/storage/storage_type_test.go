package storage_test

import (
	"testing"

	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

// TestNewTypeFromString validate that the kinds qovery.StorageTypeEnum defined in Qovery's API Client are valid.
// This is useful to make sure the storage.Type stays up to date.
func TestNewTypeFromString(t *testing.T) {
	t.Parallel()

	assert.Len(t, storage.AllowedTypeValues, len(qovery.AllowedStorageTypeEnumEnumValues))
	for _, storageType := range qovery.AllowedStorageTypeEnumEnumValues {
		storageTypeStr := string(storageType)
		t.Run(storageTypeStr, func(t *testing.T) {
			st, err := storage.NewTypeFromString(storageTypeStr)
			assert.NoError(t, err)
			assert.Equal(t, st.String(), storageTypeStr)
		})
	}
}
