package vncnetwork

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGoPackageNamespace(t *testing.T) {
	want := reflect.TypeOf(VncMsg{}).PkgPath() + ";vncnetwork"
	got := File_vncmsg_proto.Options().(*descriptorpb.FileOptions).GetGoPackage()
	if got != want {
		t.Fatalf("go_package = %q, want %q", got, want)
	}
}
