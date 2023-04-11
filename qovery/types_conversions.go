package qovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	return FromString(string(v))
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

func StringAsPointer(v string) *string {
	return &v
}

//
// Convert Terraform types to Go types
//

func ToNullableNullableBuildPackLanguageEnum(v types.String) qovery.NullableBuildPackLanguageEnum {
	enum, err := qovery.NewBuildPackLanguageEnumFromValue(v.Value)
	if err != nil || v.Null || v.Unknown {
		s := qovery.NewNullableBuildPackLanguageEnum(nil)
		return *s
	}
	s := qovery.NewNullableBuildPackLanguageEnum(enum)
	return *s
}

func ToNullableString(v types.String) qovery.NullableString {
	if v.Null || v.Unknown {
		s := qovery.NewNullableString(nil)
		return *s
	}
	s := qovery.NewNullableString(&v.Value)
	return *s
}

func ToString(v types.String) string {
	return v.Value
}

func ToStringPointer(v types.String) *string {
	if v.Null || v.Unknown {
		return nil
	}
	return &v.Value
}

func ToBool(v types.Bool) bool {
	return v.Value
}

func ToBoolPointer(v types.Bool) *bool {
	if v.Null || v.Unknown {
		return nil
	}
	return &v.Value
}

func ToInt32(v types.Int64) int32 {
	return int32(v.Value)
}

func ToInt32Pointer(v types.Int64) *int32 {
	if v.Null || v.Unknown {
		return nil
	}
	i := int32(v.Value)
	return &i
}

func ToUInt32Pointer(v types.Int64) *uint32 {
	if v.Null || v.Unknown {
		return nil
	}
	i := uint32(v.Value)
	return &i
}

func ToInt64(v types.Int64) int32 {
	return int32(v.Value)
}

func ToInt64Pointer(v types.Int64) *int32 {
	if v.Null || v.Unknown {
		return nil
	}
	i := int32(v.Value)
	return &i
}

func ToMapStringString(obj types.Object) (map[string]interface{}, error) {
	ret := make(map[string]interface{}, len(obj.Attrs))
	for k, v := range obj.Attrs {
		value, err := FromTfValueToGoValue(v)
		if err != nil {
			return nil, err
		}
		ret[k] = value
	}
	return ret, nil
}

func ToStringArray(set types.List) []string {
	if set.Null || set.Unknown {
		return []string{}
	}

	array := make([]string, 0, len(set.Elems))
	for _, elem := range set.Elems {
		array = append(array, ToString(elem.(types.String)))
	}

	return array
}

//
// Convert Go types to Terraform types
//

func FromNullableNullableBuildPackLanguageEnum(v qovery.NullableBuildPackLanguageEnum) types.String {
	if v.Get() == nil {
		return FromStringPointer(nil)
	}
	return FromString(string(*v.Get()))
}

func FromString(v string) types.String {
	return types.String{Value: v}
}

func FromStringPointer(v *string) types.String {
	if v == nil {
		return types.String{Null: true}
	}
	return FromString(*v)
}

func FromNullableString(v qovery.NullableString) types.String {
	if v.Get() == nil {
		return types.String{Null: true}
	}
	return FromString(*v.Get())
}

func FromInt64(v int64) types.Int64 {
	return types.Int64{Value: v}
}

func FromUInt64(v uint64) types.Int64 {
	return types.Int64{Value: int64(v)}
}

func FromInt32(v int32) types.Int64 {
	return FromInt64(int64(v))
}

func FromUInt32(v uint32) types.Int64 {
	return FromUInt64(uint64(v))
}

func FromInt32Pointer(v *int32) types.Int64 {
	if v == nil {
		return types.Int64{Null: true}
	}
	return FromInt32(*v)
}

func FromBool(v bool) types.Bool {
	return types.Bool{Value: v}
}

func FromBoolPointer(v *bool) types.Bool {
	if v == nil {
		return types.Bool{Null: true}
	}
	return FromBool(*v)
}

func FromStringArray(array []string) types.List {
	set := types.List{
		ElemType: types.StringType,
	}

	if array == nil {
		set.Null = true
		return set
	}

	set.Elems = make([]attr.Value, 0, len(array))
	for _, v := range array {
		set.Elems = append(set.Elems, FromString(v))
	}
	return set
}

func FromGoValueToTfValue(value interface{}, _type attr.Type) (attr.Value, error) {
	switch _type {
	case types.StringType:
		return types.String{Value: value.(string)}, nil
	case types.BoolType:
		return types.Bool{Value: value.(bool)}, nil
	case types.Int64Type:
		return types.Int64{Value: int64(value.(float64))}, nil
	case types.SetType{ElemType: types.StringType}:
		var elems []attr.Value
		for _, v := range value.([]interface{}) {
			elems = append(elems, types.String{Value: strings.TrimSpace(v.(string))})
		}
		return types.Set{ElemType: types.StringType, Elems: elems}, nil
	case types.MapType{ElemType: types.StringType}:
		elems := make(map[string]attr.Value)
		for k, v := range value.(map[string]interface{}) {
			elems[k] = types.String{Value: v.(string)}
		}
		return types.Map{ElemType: types.StringType, Elems: elems}, nil
	}

	return types.Object{Null: true}, fmt.Errorf("unable to parse %s as %s", value, _type.String())
}

func FromTfValueToGoValue(v attr.Value) (interface{}, error) {
	switch v.Type(context.Background()) {
	case types.StringType:
		value := strings.Trim(v.String(), "\"")
		return value, nil
	case types.Int64Type:
		value, err := strconv.ParseInt(v.String(), 10, 64)
		return value, err
	case types.BoolType:
		value, err := strconv.ParseBool(v.String())
		return value, err
	case types.SetType{ElemType: types.StringType}:
		var elems []string
		jsonErr := json.Unmarshal([]byte(v.String()), &elems)
		return elems, jsonErr
	case types.MapType{ElemType: types.StringType}:
		elems := make(map[string]string)
		jsonErr := json.Unmarshal([]byte(v.String()), &elems)
		return elems, jsonErr
	}

	return nil, fmt.Errorf("unable to parse %s as Go value", v.String())
}

func FromStringMap(value *map[string]interface{}) types.Object {
	if value == nil || len(*value) == 0 {
		return types.Object{Null: true}
	}

	attrs := make(map[string]attr.Value)
	attrTypes := make(map[string]attr.Type)
	for k, f := range advancedSettingsDefault {
		attrTypes[k] = f._type
	}

	for k, f := range *value {
		attribute, err := FromGoValueToTfValue(f, attrTypes[k])

		if err != nil {
			tflog.Warn(context.Background(), "Unable to parse attribute, using default value.", map[string]interface{}{"error": err.Error()})
			attribute = advancedSettingsDefault[k].defaultValue
		}

		attrs[k] = attribute
	}

	return types.Object{
		Attrs:     attrs,
		AttrTypes: attrTypes,
	}
}
