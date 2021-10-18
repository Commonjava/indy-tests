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
