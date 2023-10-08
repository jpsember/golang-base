package webserv

import (
	"errors"
	. "github.com/jpsember/golang-base/base"
	"hash/fnv"
	"os/exec"
	"strings"
	"sync"
)

func MakeSysCall(c ...string) (string, error) {
	Todo("!have a basic function to convert 'any' to string")
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
	cacheResults  JSMap
	cachePath     Path
	lock          sync.RWMutex
	modified      bool
}

type HTMLValidator = *HTMLValidatorStruct

var sharedHTMLValidator HTMLValidator

func SharedHTMLValidator() HTMLValidator {
	if sharedHTMLValidator == nil {
		sharedHTMLValidator = &HTMLValidatorStruct{
			validatorPath: ProjectDirM().JoinM("validator/vnu.jar").EnsureExists(),
			cachePath:     ProjectDirM().JoinM("validator/cached_results.json"),
		}
		SharedBackgroundTaskManager().Add("html_validator", JSec*1, sharedHTMLValidator.flushResults)
	}
	return sharedHTMLValidator
}

func HashOfString(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32() & 0xffffffff)
}

func (h HTMLValidator) ValidateWithoutCache(content string) (JSMap, error) {
	return h.auxValidate(content, false)
}

func (h HTMLValidator) Validate(content string) (JSMap, error) {
	return h.auxValidate(content, true)
}
func (h HTMLValidator) auxValidate(content string, useCache bool) (JSMap, error) {

	pr := PrIf("HTMLValidator.Validate", false)

	var cache JSMap
	key := ""
	if useCache {
		key = IntToString(HashOfString(content))
		cache = h.cachedResults()
		h.lock.RLock()
		res := cache.OptMap(key)
		h.lock.RUnlock()

		if res != nil {
			var err error
			errstr := res.OptString("error", "")
			if errstr != "" {
				err = errors.New(errstr)
			}
			return res, err
		}
	}

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
	if err != nil {
		results.Put("error", err.Error())
		pr("err:", err)
	}
	pr("output:", INDENT, results)

	if useCache {
		h.lock.Lock()
		cache.Put(key, results)
		pr("storing results in cache:", key)
		h.modified = true
		h.lock.Unlock()
	}

	return results, err
}

func (h HTMLValidator) cachedResults() JSMap {
	if h.cacheResults == nil {
		h.lock.Lock()
		h.cacheResults = JSMapFromFileIfExistsM(h.cachePath)
		h.lock.Unlock()
	}
	return h.cacheResults
}

func (h HTMLValidator) flushResults() {
	pr := PrIf("HTMLValidator.flushResults", false)
	if !h.modified {
		return
	}
	pr("modified:", h.modified)
	h.lock.Lock()
	defer h.lock.Unlock()
	h.cachePath.WriteStringM(h.cacheResults.CompactString())
	pr("...wrote results", INDENT, h.cacheResults)
	h.modified = false
}
