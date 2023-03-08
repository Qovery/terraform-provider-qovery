package qovery

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qovery/qovery-client-go"

	"github.com/qovery/terraform-provider-qovery/internal/domain/environment"
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
		qovery.BuildPackLanguageEnum |
		qovery.CloudProviderEnum |
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
		qovery.BuildModeEnum
}

func clientEnumToStringArray[T ClientEnum](enum []T) []string {
	arr := make([]string, len(enum))
	for idx, e := range enum {
		arr[idx] = fmt.Sprintf("%s", e)
	}
	return arr
}

func fromClientEnum[T ClientEnum](v T) types.String {
	return fromString(string(v))
}

func fromClientEnumPointer[T ClientEnum](v *T) types.String {
	if v == nil {
		return types.String{Null: true}
	}
	return fromClientEnum(*v)
}

//
// Convert to pointer
//

func stringAsPointer(v string) *string {
	return &v
}

//
// Convert Terraform types to Go types
//

func toNullableNullableBuildPackLanguageEnum(v types.String) qovery.NullableBuildPackLanguageEnum {
	enum, err := qovery.NewBuildPackLanguageEnumFromValue(v.Value)
	if err != nil || v.Null || v.Unknown {
		s := qovery.NewNullableBuildPackLanguageEnum(nil)
		return *s
	}
	s := qovery.NewNullableBuildPackLanguageEnum(enum)
	return *s
}

func toNullableString(v types.String) qovery.NullableString {
	if v.Null || v.Unknown {
		s := qovery.NewNullableString(nil)
		return *s
	}
	s := qovery.NewNullableString(&v.Value)
	return *s
}

func toString(v types.String) string {
	return v.Value
}

func toStringPointer(v types.String) *string {
	if v.Null || v.Unknown {
		return nil
	}
	return &v.Value
}

func toBool(v types.Bool) bool {
	return v.Value
}

func toBoolPointer(v types.Bool) *bool {
	if v.Null || v.Unknown {
		return nil
	}
	return &v.Value
}

func toInt32(v types.Int64) int32 {
	return int32(v.Value)
}

func toInt32Pointer(v types.Int64) *int32 {
	if v.Null || v.Unknown {
		return nil
	}
	i := int32(v.Value)
	return &i
}

func toInt64(v types.Int64) int32 {
	return int32(v.Value)
}

func toInt64Pointer(v types.Int64) *int32 {
	if v.Null || v.Unknown {
		return nil
	}
	i := int32(v.Value)
	return &i
}

func toMapStringString(v types.Map) map[string]string {
	ret := make(map[string]string, len(v.Elems))
	for k, v := range v.Elems {
		ret[k] = v.String()
	}
	return ret
}

func toStringArray(set types.List) []string {
	if set.Null || set.Unknown {
		return []string{}
	}

	array := make([]string, 0, len(set.Elems))
	for _, elem := range set.Elems {
		array = append(array, toString(elem.(types.String)))
	}

	return array
}

//
// Convert Go types to Terraform types
//

func fromNullableNullableBuildPackLanguageEnum(v qovery.NullableBuildPackLanguageEnum) types.String {
	if v.Get() == nil {
		return fromStringPointer(nil)
	}
	return fromString(string(*v.Get()))
}

func fromString(v string) types.String {
	return types.String{Value: v}
}

func fromStringPointer(v *string) types.String {
	if v == nil {
		return types.String{Null: true}
	}
	return fromString(*v)
}

func fromNullableString(v qovery.NullableString) types.String {
	if v.Get() == nil {
		return types.String{Null: true}
	}
	return fromString(*v.Get())
}

func fromInt64(v int64) types.Int64 {
	return types.Int64{Value: v}
}

func fromInt32(v int32) types.Int64 {
	return fromInt64(int64(v))
}

func fromInt32Pointer(v *int32) types.Int64 {
	if v == nil {
		return types.Int64{Null: true}
	}
	return fromInt32(*v)
}

func fromBool(v bool) types.Bool {
	return types.Bool{Value: v}
}

func fromBoolPointer(v *bool) types.Bool {
	if v == nil {
		return types.Bool{Null: true}
	}
	return fromBool(*v)
}

func fromStringArray(array []string) types.List {
	set := types.List{
		ElemType: types.StringType,
	}

	if array == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(array))
	for _, v := range array {
		set.Elems = append(set.Elems, fromString(v))
	}
	return set
}
