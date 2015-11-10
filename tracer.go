package tracer

import (
  "errors"
  "strings"
  "sync"
  "os"

  . "github.com/akolosov/injector"
  . "github.com/akolosov/logger"
)

type TracerInterface interface {
  SetOffset(int)
  SetOptionalFunc(func(Stack))
  HandleError(bool)
  HandleErrorWithAction(bool, func(Stack))
  ResetOptionalFunc()
}

type Tracer struct {
  // Интерфейс к функциям DI/I
  InjectorInterface
  // Указатель на систему логирования по умолчанию
  LoggerInterface       `injection:"logger"`
  // Логгировать (true) дополнительную информацию
  Verbose bool          `injection:"verbose_output"`
  // Системный mutex
  mutex sync.Mutex
  // Смещение в стэке вызовов
  offset int
  // Функция дополнительной обработки стэка вызовов
  optionalFunc func(Stack)
}

// Singleton-объект
var masterTracer *Tracer

func NewTracer(injector InjectorInterface) TracerInterface {
  if masterTracer == nil {
    masterTracer = &Tracer{InjectorInterface: injector}
    masterTracer.offset = 0
  }

  masterTracer.Inject(masterTracer)

  masterTracer.Register("tracer", masterTracer)

  return masterTracer
}

func (this *Tracer) SetOffset(offset int) {
  this.mutex.Lock()
  defer this.mutex.Unlock()

  this.offset = offset
}

func (this *Tracer) SetOptionalFunc(optionalFunc func(Stack)) {
  this.mutex.Lock()
  defer this.mutex.Unlock()

  this.optionalFunc = optionalFunc
}

func (this *Tracer) ResetOptionalFunc() {
  this.mutex.Lock()
  defer this.mutex.Unlock()

  this.optionalFunc = nil
}

func (this *Tracer) HandleErrorWithAction(recovered bool, action func(Stack)) {
  this.mutex.Lock()
  defer this.mutex.Unlock()

  if e := recover(); e != nil {
    err, ok := e.(error)
    if !ok {
      err = errors.New(e.(string))
    }

    stack := this.traceError(err)

    if action != nil {
      action(stack)
    }
    if !recovered {
      os.Exit(254)
    }
  }
}

func (this *Tracer) HandleError(recovered bool) {
  this.mutex.Lock()
  defer this.mutex.Unlock()

  if e := recover(); e != nil {
    err, ok := e.(error)
    if !ok {
      err = errors.New(e.(string))
    }

    stack := this.traceError(err)

    if this.optionalFunc != nil {
      this.optionalFunc(stack)
    }
    if !recovered {
      os.Exit(254)
    }
  }
}

func (this *Tracer) traceError(err error) Stack {
  if err != nil && !strings.HasPrefix(err.Error(), "PANIC") {
    this.Print("PANIC: ", err.Error())
  }
  this.Print("----------------------[ Stack trace begin ]-----------------------")
  stack := CurrentStack(this.offset)
  for _, trace := range stack  {
    this.Printf("%s/%s - %s [%d]", trace.PackageName, trace.FileName, trace.MethodName, trace.LineNumber)
  }
  this.Print("-----------------------[ Stack trace end ]------------------------")

  return stack
}
