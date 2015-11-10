package tracer

import (
  "testing"
  "os"

  . "github.com/akolosov/injector"
  . "github.com/akolosov/logger"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

func TestTracer(t *testing.T) {
  RegisterFailHandler(Fail)
  RunSpecs(t, "Tracer Suite")
}

var ST, ST_X *Tracer
var OptionalFunc func(Stack)

var _ = BeforeSuite(func() {
  ST = NewTracer(NewInjector()).(*Tracer)
  ST.Register("verbose_output", testing.Verbose())
  ST.Register("logger", NewLogger(NewStdLogger(os.Stderr, "TEST: "), false))
  ST.Inject(ST)

  OptionalFunc = func(stack Stack) {
    for _, trace := range stack  {
      ST.Printf("OPTIONAL => %s/%s - %s [%d]", trace.PackageName, trace.FileName, trace.MethodName, trace.LineNumber)
    }
  }

  ST.SetOptionalFunc(OptionalFunc)
})

var _ = Describe("Testing with Ginkgo", func() {
  Describe("#Tracer", func() {


    It("Tracer is initialized", func() {
      Expect(ST).ShouldNot(BeNil())
    })

    It("Tracer is singleton", func() {
      Expect(ST).Should(Equal(NewTracer(NewInjector())))
    })

    It("Current trace", func() {
      stack := CurrentStack(0)
      Expect(stack).ShouldNot(BeNil())
      Expect(len(stack)).ShouldNot(Equal(0))
    })

    It("Handle Panic error", func() {
      defer ST.HandleError(true)
      ST.SetOffset(6)
      ST.Panic("HandleError Testing")
    })

    It("Handle Panicf error", func() {
      defer ST.HandleError(true)
      ST.SetOffset(6)
      ST.Panicf("Message: %v", ST)
    })

    It("Handle panic error", func() {
      defer ST.HandleError(true)
      ST.SetOffset(3)
      ST.ResetOptionalFunc()
      panic("HandleError Testing")
    })

    It("Handle runtime error", func() {
      defer ST.HandleErrorWithAction(true, OptionalFunc)
      ST.SetOffset(0)
      ST.Printf("%v", ST_X.Verbose)
    })
  })
})

