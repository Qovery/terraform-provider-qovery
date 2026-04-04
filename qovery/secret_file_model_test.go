//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSecretFileList_diffRequest(t *testing.T) {
	t.Parallel()

	t.Run("create_new_file", func(t *testing.T) {
		t.Parallel()
		newList := SecretFileList{
			{
				Key:         types.StringValue("CONFIG"),
				Value:       types.StringValue("content"),
				MountPath:   types.StringValue("/etc/config.yaml"),
				Description: types.StringValue("config file"),
			},
		}
		oldList := SecretFileList{}

		diff := newList.diffRequest(oldList)
		assert.Len(t, diff.Create, 1)
		assert.Len(t, diff.Update, 0)
		assert.Len(t, diff.Delete, 0)
		assert.Equal(t, "CONFIG", diff.Create[0].Key)
		assert.Equal(t, "/etc/config.yaml", diff.Create[0].MountPath)
	})

	t.Run("delete_file", func(t *testing.T) {
		t.Parallel()
		newList := SecretFileList{}
		oldList := SecretFileList{
			{
				Id:        types.StringValue("some-id"),
				Key:       types.StringValue("CONFIG"),
				Value:     types.StringValue("content"),
				MountPath: types.StringValue("/etc/config.yaml"),
			},
		}

		diff := newList.diffRequest(oldList)
		assert.Len(t, diff.Create, 0)
		assert.Len(t, diff.Update, 0)
		assert.Len(t, diff.Delete, 1)
		assert.Equal(t, "some-id", diff.Delete[0].SecretID)
	})

	t.Run("update_value_only", func(t *testing.T) {
		t.Parallel()
		newList := SecretFileList{
			{
				Key:       types.StringValue("CONFIG"),
				Value:     types.StringValue("new-content"),
				MountPath: types.StringValue("/etc/config.yaml"),
			},
		}
		oldList := SecretFileList{
			{
				Id:        types.StringValue("some-id"),
				Key:       types.StringValue("CONFIG"),
				Value:     types.StringValue("old-content"),
				MountPath: types.StringValue("/etc/config.yaml"),
			},
		}

		diff := newList.diffRequest(oldList)
		assert.Len(t, diff.Create, 0)
		assert.Len(t, diff.Update, 1)
		assert.Len(t, diff.Delete, 0)
		assert.Equal(t, "some-id", diff.Update[0].SecretID)
		assert.Equal(t, "new-content", diff.Update[0].Value)
	})

	t.Run("mount_path_change_triggers_delete_and_create", func(t *testing.T) {
		t.Parallel()
		newList := SecretFileList{
			{
				Key:       types.StringValue("CONFIG"),
				Value:     types.StringValue("content"),
				MountPath: types.StringValue("/new/path/config.yaml"),
			},
		}
		oldList := SecretFileList{
			{
				Id:        types.StringValue("some-id"),
				Key:       types.StringValue("CONFIG"),
				Value:     types.StringValue("content"),
				MountPath: types.StringValue("/old/path/config.yaml"),
			},
		}

		diff := newList.diffRequest(oldList)
		assert.Len(t, diff.Create, 1)
		assert.Len(t, diff.Update, 0)
		assert.Len(t, diff.Delete, 1)
		assert.Equal(t, "some-id", diff.Delete[0].SecretID)
		assert.Equal(t, "/new/path/config.yaml", diff.Create[0].MountPath)
	})

	t.Run("no_changes", func(t *testing.T) {
		t.Parallel()
		entry := SecretFile{
			Id:        types.StringValue("some-id"),
			Key:       types.StringValue("CONFIG"),
			Value:     types.StringValue("content"),
			MountPath: types.StringValue("/etc/config.yaml"),
		}
		newList := SecretFileList{entry}
		oldList := SecretFileList{entry}

		diff := newList.diffRequest(oldList)
		assert.Len(t, diff.Create, 0)
		assert.Len(t, diff.Update, 0)
		assert.Len(t, diff.Delete, 0)
	})
}
