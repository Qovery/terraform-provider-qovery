package qovery

import "github.com/hashicorp/terraform-plugin-framework/types"

//
// Convert Terraform types to Go types
//

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

//
// Convert Go types to Terraform types
//

func fromString(v string) types.String {
	return types.String{Value: v}
}

func fromStringPointer(v *string) types.String {
	if v == nil {
		return types.String{Null: true}
	}
	return fromString(*v)
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
