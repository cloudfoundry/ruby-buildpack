package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("app using system yaml library", func() {
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	BeforeEach(func() {
		app = cutlass.New(Fixtures("sinatra"))
	})

	It("displays metasyntactic variables as yaml", func() {
		PushAppAndConfirm(app)
		Expect(app.GetBody("/yaml")).To(ContainSubstring(`---
foo:
- bar
- baz
- quux
`))
	})
})
