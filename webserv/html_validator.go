package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"os/exec"
	"strings"
)

func MakeSysCall(c ...string) (string, error) {
	pr := PrIf("MakeSysCall", false)
	var cmd = c[0]
	var args = c[1:]
	pr("MakeSysCall:", INDENT, cmd, args)
	syscall := exec.Command(cmd, args...)
	out, err := syscall.CombinedOutput()
	var strout = string(out)
	return strout, err
}

type HTMLValidatorStruct struct {
	validatorPath Path
}

type HTMLValidator = *HTMLValidatorStruct

var sharedHTMLValidator HTMLValidator

func SharedHTMLValidator() HTMLValidator {
	if sharedHTMLValidator == nil {
		sharedHTMLValidator = &HTMLValidatorStruct{
			validatorPath: ProjectDirM().JoinM("validator/vnu.jar").EnsureExists(),
		}
	}
	return sharedHTMLValidator
}

func (h HTMLValidator) Validate(content string) (JSMap, error) {
	pr := PrIf("HTMLValidator.Validate", true)
	if !strings.Contains(content, `<html>`) {
		content = `<!DOCTYPE html>
<html>
 <head>
   <title>Validate wrapper</title>
 </head>
 <body>
` + content + `
</body>
</html>
`
	}

	pr("validating content:", INDENT, content)

	var tempFile Path
	if false && Alert("using special temp file") {
		tempFile = "_SKIP_.txt"
		Pr(CurrentDirectory(), tempFile)
	} else {
		tempFile = TempFileM("htmlvalidatorinput")
		defer tempFile.DeleteFileM()
	}
	tempFile.WriteStringM(content)
	output, err := MakeSysCall("java", "-jar", h.validatorPath.String(), "--stdout", "--asciiquotes", "--no-langdetect", "--format", "json", tempFile.String())

	results := JSMapFromStringM(output)
	pr("err:", err)
	pr("output:", INDENT, results)
	return results, err
}
