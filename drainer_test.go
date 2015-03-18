package gate_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/cloudfoundry-incubator/garden"
	gfakes "github.com/cloudfoundry-incubator/garden/fakes"

	. "github.com/concourse/gate"
)

var _ = Describe("Drainer", func() {
	var (
		drainer *Drainer
		logger  *lagertest.TestLogger

		fakeGardenClient *gfakes.FakeClient
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("drainer")
		fakeGardenClient = new(gfakes.FakeClient)
	})

	JustBeforeEach(func() {
		drainer = NewDrainer(logger, fakeGardenClient)
	})

	Context("when there are containers to clean up", func() {
		var (
			containerA *gfakes.FakeContainer
			containerB *gfakes.FakeContainer
		)

		BeforeEach(func() {
			containerA = new(gfakes.FakeContainer)
			containerA.HandleReturns("container-a")

			containerB = new(gfakes.FakeContainer)
			containerB.HandleReturns("container-b")

			fakeGardenClient.ContainersReturns([]garden.Container{containerA, containerB}, nil)
		})

		It("cleans up ephemeral containers", func() {
			err := drainer.Drain()
			Ω(err).ShouldNot(HaveOccurred())

			By("querying for ephemeral containers", func() {
				Ω(fakeGardenClient.ContainersCallCount()).Should(Equal(1))
				queriedProperties := fakeGardenClient.ContainersArgsForCall(0)

				Ω(queriedProperties).Should(HaveKeyWithValue("concourse:ephemeral", "true"))
			})

			By("taking each of the returned containers and destroying it", func() {
				Ω(fakeGardenClient.DestroyCallCount()).Should(Equal(2))

				destroyedHandle := fakeGardenClient.DestroyArgsForCall(0)
				Ω(destroyedHandle).Should(Equal("container-a"))

				destroyedHandle = fakeGardenClient.DestroyArgsForCall(1)
				Ω(destroyedHandle).Should(Equal("container-b"))
			})
		})

		Context("when one of the containers fails to delete", func() {
			BeforeEach(func() {
				fakeGardenClient.DestroyStub = func(handle string) error {
					if handle == containerA.Handle() {
						return errors.New("container a cannot be destroyed at this time")
					}

					return nil
				}
			})

			It("keeps on goin' but returns a composite error", func() {
				err := drainer.Drain()
				Ω(err).Should(HaveOccurred())

				Ω(fakeGardenClient.DestroyCallCount()).Should(Equal(2))
			})
		})
	})

	Context("when garden returns an error", func() {
		disaster := errors.New("oh no")

		BeforeEach(func() {
			fakeGardenClient.ContainersReturns(nil, disaster)
		})

		It("re-returns that error", func() {
			err := drainer.Drain()
			Ω(err).Should(MatchError(disaster))
		})
	})
})
