package gate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gate Suite")
}
