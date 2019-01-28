// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: process_info.proto

/*
Package process is a generated protocol buffer package.

It is generated from these files:
	process_info.proto

It has these top-level messages:
	ProcessInfo
	ProcessInfoMap
*/
package process

import fmt "fmt"
import go_proto_validators "github.com/mwitkow/go-proto-validators"
import proto "github.com/gogo/protobuf/proto"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import _ "github.com/mwitkow/go-proto-validators"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *ProcessInfo) Validate() error {
	if !(this.StartTime > 0) {
		return go_proto_validators.FieldError("StartTime", fmt.Errorf(`value '%v' must be greater than '0'`, this.StartTime))
	}
	if !(this.PID >= 0) {
		return go_proto_validators.FieldError("PID", fmt.Errorf(`value '%v' must be greater than '0'`, this.PID))
	}
	if this.Program == "" {
		return go_proto_validators.FieldError("Program", fmt.Errorf(`value '%v' must not be an empty string`, this.Program))
	}
	return nil
}
func (this *ProcessInfoMap) Validate() error {
	if !(this.Version > 0) {
		return go_proto_validators.FieldError("Version", fmt.Errorf(`value '%v' must be greater than '0'`, this.Version))
	}
	for _, item := range this.InfoMap {
		if err := go_proto_validators.CallValidatorIfExists(&(item)); err != nil {
			return go_proto_validators.FieldError("InfoMap", err)
		}
	}
	return nil
}
