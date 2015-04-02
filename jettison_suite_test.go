package jettison_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestJettison(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jettison Suite")
}
