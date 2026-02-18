package qovery

import (
	"fmt"
	"time"

	"github.com/qovery/terraform-provider-qovery/internal/domain/helm"
	"github.com/qovery/terraform-provider-qovery/internal/domain/helmRepository"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
	"github.com/qovery/terraform-provider-qovery/internal/domain/gittoken"
	"github.com/qovery/terraform-provider-qovery/internal/domain/organization"
	"github.com/qovery/terraform-provider-qovery/internal/domain/port"
	"github.com/qovery/terraform-provider-qovery/internal/domain/registry"
	"github.com/qovery/terraform-provider-qovery/internal/domain/status"
	"github.com/qovery/terraform-provider-qovery/internal/domain/storage"
)

//
// Convert client enums to []string
//

type ClientEnum interface {
	environment.Mode |
		organization.Plan |
		port.Protocol |
		qovery.CloudProviderEnum |
		qovery.CloudVendorEnum |
		qovery.CustomDomainStatusEnum |
		qovery.DatabaseAccessibilityEnum |
		qovery.DatabaseModeEnum |
		qovery.DatabaseTypeEnum |
		qovery.KubernetesEnum |
		qovery.PlanEnum |
		qovery.PortProtocolEnum |
		qovery.StateEnum |
		qovery.StorageTypeEnum |
		registry.Kind |
		status.State |
		storage.Type |
		qovery.BuildModeEnum |
		qovery.ClusterStateEnum |
		gittoken.GitTokenType |
		helmRepository.Kind |
		helm.Protocol |
		qovery.OrganizationAnnotationsGroupScopeEnum |
		qovery.JobLifecycleTypeEnum
}

func clientEnumToStringArray[T ClientEnum](enum []T) []string {
	arr := make([]string, len(enum))
	for idx, e := range enum {
		arr[idx] = fmt.Sprintf("%s", e)
	}
	return arr
}

func fromClientEnum[T ClientEnum](v T) types.String {
	return FromString(string(v))
}

func fromClientEnumPointer[T ClientEnum](v *T) types.String {
	if v == nil {
		return basetypes.NewStringNull()
	}
	return fromClientEnum(*v)
}

//
// Convert to pointer
//

//go:fix inline
func StringAsPointer(v string) *string {
	return new(v)
}

func ToNullableString(v types.String) qovery.NullableString {
	if v.IsNull() || v.IsUnknown() {
		s := qovery.NewNullableString(nil)
		return *s
	}
	s := qovery.NewNullableString(v.ValueStringPointer())
	return *s
}

func ToString(v types.String) string {
	return v.ValueString()
}

func ToStringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}

func ToBool(v types.Bool) bool {
	return v.ValueBool()
}

func ToBoolPointer(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueBoolPointer()
}

func ToInt32(v types.Int64) int32 {
	return int32(v.ValueInt64())
}

func ToInt32Pointer(v types.Int64) *int32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int32(v.ValueInt64())
	return &i
}

func ToInt64Pointer(v types.Int64) *int32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int32(v.ValueInt64())
	return &i
}

func ToStringArray(v types.List) []string {
	if v.IsNull() || v.IsUnknown() {
		return []string{}
	}

	array := make([]string, 0, len(v.Elements()))
	for _, elem := range v.Elements() {
		array = append(array, ToString(elem.(types.String)))
	}

	return array
}

func ToStringArrayFromSet(v types.Set) []string {
	if v.IsNull() || v.IsUnknown() {
		return []string{}
	}

	array := make([]string, 0, len(v.Elements()))
	for _, elem := range v.Elements() {
		array = append(array, ToString(elem.(types.String)))
	}

	return array
}

func FromString(v string) types.String {
	return basetypes.NewStringValue(v)
}

func FromStringPointer(v *string) types.String {
	if v == nil {
		return basetypes.NewStringNull()
	}
	return FromString(*v)
}

func FromNullableString(v qovery.NullableString) types.String {
	if v.Get() == nil {
		return basetypes.NewStringNull()
	}
	return FromString(*v.Get())
}

func FromInt64(v int64) types.Int64 {
	return basetypes.NewInt64Value(v)
}

func FromUInt64(v uint64) types.Int64 {
	return basetypes.NewInt64Value(int64(v))
}

func FromInt32(v int32) types.Int64 {
	return FromInt64(int64(v))
}

func FromUInt32(v uint32) types.Int64 {
	return FromUInt64(uint64(v))
}

func FromInt32Pointer(v *int32) types.Int64 {
	if v == nil {
		return basetypes.NewInt64Null()
	}
	return FromInt32(*v)
}

func FromBool(v bool) types.Bool {
	return basetypes.NewBoolValue(v)
}

func FromBoolPointer(v *bool) types.Bool {
	if v == nil {
		return basetypes.NewBoolNull()
	}
	return FromBool(*v)
}

func FromStringArray(array []string) types.List {
	if array == nil {
		return basetypes.NewListNull(types.StringType)
	}

	elements := make([]attr.Value, 0, len(array))
	for _, v := range array {
		elements = append(elements, FromString(v))
	}
	value, _ := basetypes.NewListValue(types.StringType, elements)
	return value
}

func FromStringSet(array []string) types.Set {
	if array == nil {
		return basetypes.NewSetNull(types.StringType)
	}

	elements := make([]attr.Value, 0, len(array))
	for _, v := range array {
		elements = append(elements, FromString(v))
	}
	value, _ := basetypes.NewSetValue(types.StringType, elements)
	return value
}

func FromTime(t time.Time) types.String {
	return FromString(t.Format(time.RFC3339))
}

func FromTimePointer(t *time.Time) types.String {
	if t == nil {
		return basetypes.NewStringNull()
	}
	return FromTime(*t)
}
