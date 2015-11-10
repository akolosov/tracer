package tracer

import (
  "fmt"
  "runtime"
  "strconv"
  "strings"
)

type stackElement struct {
  LineNumber  int    `json:"lineNumber"`
  PackageName string `json:"className"`
  FileName    string `json:"fileName"`
  MethodName  string `json:"methodName"`
}

type Stack []stackElement

type stackTrace interface {
  AddEntry(lineNumber int, packageName string, fileName string, methodName string)
}

func (this *Stack) AddEntry(lineNumber int, packageName, fileName, methodName string) {
  *this = append(*this, stackElement{LineNumber: lineNumber, PackageName: packageName, FileName: fileName, MethodName: methodName})
}

func CurrentStack(offset int) Stack {
  stack := make(Stack, 0, 0)
  current(&stack)
  return stack[offset:]
}

func current(stack stackTrace) {
  rawStack := make([]byte, 1<<16)
  rawStack = rawStack[:runtime.Stack(rawStack, false)]
  parse(rawStack, stack)
}

func parse(trace []byte, stack stackTrace) {
  lines := strings.Split(string(trace), "\n")

  var lineNumber int
  var fileName, packageName, methodName string

  for index, line := range lines[1:] {
    if len(line) == 0 {
      continue
    }
    if index%2 == 0 {
      packageName, methodName = extractPackageName(line)
    } else {
      lineNumber, fileName = extractLineNumberAndFile(line)
      stack.AddEntry(lineNumber, packageName, fileName, methodName)
    }
  }
}

func extractPackageName(line string) (packageName, methodName string) {
  packagePath, packageNameAndFunction := splitAtLastSlash(line)
  parts := strings.Split(packageNameAndFunction, ".")
  packageName = parts[0]
  if len(packagePath) > 0 {
    packageName = fmt.Sprintf("%s/%s", packagePath, packageName)
  }
  methodName = strings.Join(parts[1:], ".")
  return
}

func extractLineNumberAndFile(line string) (lineNumber int, fileName string) {
  _, fileAndLine := splitAtLastSlash(line)
  fileAndLine = removeSpaceAndSuffix(fileAndLine)
  parts := strings.Split(fileAndLine, ":")

  numberAsString := parts[1]
  number, _ := strconv.ParseUint(numberAsString, 10, 32)
  lineNumber = int(number)

  fileName = parts[0]
  return lineNumber, fileName
}

func splitAtLastSlash(line string) (left, right string) {
  parts := strings.Split(line, "/")
  right = parts[len(parts)-1]
  left = strings.Join(parts[:len(parts)-1], "/")
  return
}

func removeSpaceAndSuffix(line string) string {
  parts := strings.Split(line, " ")
  return strings.Join(parts[:len(parts)-1], " ")
}
