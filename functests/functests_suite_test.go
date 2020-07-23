package functests

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	BeforeTestSuiteSetup()
})

var _ = AfterSuite(func() {
	AfterTestSuiteCleanup()
})
