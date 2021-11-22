package buildtest

import (
	"testing"

	"github.com/commonjava/indy-tests/pkg/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlterUploadPath(t *testing.T) {
	rawPath := "/org/apache/kafka/connect-api/2.7.0.redhat-00012/connect-api-2.7.0.redhat-00012-javadoc.jar"
	altered := common.AlterUploadPath(rawPath, "999999")
	expected := "/org/apache/kafka/connect-api/2.7.0.redhat-999999/connect-api-2.7.0.redhat-999999-javadoc.jar"
	Convey("Replacing should work", t, func() {
		So(altered, ShouldEqual, expected)
	})
}

func Test_getRepoNameByOriginUrl(t *testing.T) {
	type args struct {
		originUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "aTest", args: args{originUrl: "https://nodejs.org/dist/v8.11.4/node-v8.11.4-linux-x64.tar.gz"}, want: "nodejs-org"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRepoNameByOriginUrl(tt.args.originUrl); got != tt.want {
				t.Errorf("getRepoNameByOriginUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
