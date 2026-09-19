package network

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGoPackageNamespace(t *testing.T) {
	want := reflect.TypeOf(Msg{}).PkgPath() + ";network"
	for _, file := range []protoreflect.FileDescriptor{
		File_msg_proto, File_code_proto, File_connect_proto,
		File_forward_proto, File_shell_proto, File_vnc_proto,
	} {
		got := file.Options().(*descriptorpb.FileOptions).GetGoPackage()
		if got != want {
			t.Errorf("%s: go_package = %q, want %q", file.Path(), got, want)
		}
	}
}
