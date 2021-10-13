package buildtest

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// func TestDecorate(t *testing.T) {
// 	Convey("Decorate should work", t, func() {
// 		Convey("No duplication", func() {
// 			testDownloads := []string{"http://testdownload/maven/for/test/org/jboss/eap/wildfly-clustering-web-extension/7.3.8.GA-redhat-00001/wildfly-clustering-web-extension-7.3.8.GA-redhat-00001.jar",
// 				"http://testdownload/maven/for/test/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom"}
// 			finalDowns := decorateChecksums(testDownloads)
// 			So(len(finalDowns), ShouldEqual, 6)
// 		})
// 		Convey("Should avoid duplication", func() {
// 			testDownloads := []string{"http://testdownload/maven/for/test/org/jboss/eap/wildfly-clustering-web-extension/7.3.8.GA-redhat-00001/wildfly-clustering-web-extension-7.3.8.GA-redhat-00001.jar",
// 				"http://testdownload/maven/for/test/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom",
// 				"http://testdownload/maven/for/test/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom.md5",
// 				"http://testdownload/maven/for/test/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom.sha1"}
// 			finalDowns := decorateChecksums(testDownloads)
// 			So(len(finalDowns), ShouldEqual, 6)
// 		})

// 	})
// }

func TestAlterUploadPath(t *testing.T) {
	rawPath := "/org/apache/kafka/connect-api/2.7.0.redhat-00012/connect-api-2.7.0.redhat-00012-javadoc.jar"
	altered := alterUploadPath(rawPath, "999999")
	expected := "/org/apache/kafka/connect-api/2.7.0.redhat-999999/connect-api-2.7.0.redhat-999999-javadoc.jar"
	Convey("Replacing should work", t, func() {
		So(altered, ShouldEqual, expected)
	})
}

// func TestReplaceTarget(t *testing.T) {
// 	Convey("Replacing should work", t, func() {
// 		testDownloads := []string{"http://testdownload/api/folo/track/build-sdekjf-galj/maven/group/build-sdekjf-galj/org/jboss/eap/wildfly-clustering-web-extension/7.3.8.GA-redhat-00001/wildfly-clustering-web-extension-7.3.8.GA-redhat-00001.jar",
// 			"http://testdownload/api/folo/track/build-sdekjf-galj/maven/group/build-sdekjf-galj/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom"}
// 		replaced := replaceTargets(testDownloads, "", "replacedtestdownload", "build-test-12345")
// 		So(replaced[0], ShouldEqual, "http://replacedtestdownload/api/folo/track/build-test-12345/maven/group/build-test-12345/org/jboss/eap/wildfly-clustering-web-extension/7.3.8.GA-redhat-00001/wildfly-clustering-web-extension-7.3.8.GA-redhat-00001.jar")
// 		So(replaced[1], ShouldEqual, "http://replacedtestdownload/api/folo/track/build-test-12345/maven/group/build-test-12345/org/jboss/eap/wildfly-configadmin/7.3.8.GA-redhat-00001/wildfly-configadmin-7.3.8.GA-redhat-00001.pom")
// 	})
// }
