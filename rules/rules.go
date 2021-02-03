
// This file is generated. Do not edit this file.

package rules

//go:generate go run embed/embed.go

var Rules = map[string]string{
    "../RULES/cc/cc.go": `package cc

import (
	"fmt"
	"strings"

	"dbt/RULES/core"
)

// Toolchain represents a C++ toolchain.
type Toolchain struct {
	Ar      core.GlobalFile
	As      core.GlobalFile
	Cc      core.GlobalFile
	Cpp     core.GlobalFile
	Cxx     core.GlobalFile
	Objcopy core.GlobalFile

	Includes core.Files
}

var defaultToolchain = Toolchain{
	Ar:      core.NewGlobalFile("ar"),
	As:      core.NewGlobalFile("as"),
	Cc:      core.NewGlobalFile("gcc"),
	Cpp:     core.NewGlobalFile("g++"),
	Cxx:     core.NewGlobalFile("gcc"),
	Objcopy: core.NewGlobalFile("objcopy"),
}

// ObjectFile compiles a single C++ source file.
type ObjectFile struct {
	Src       core.File
	Includes  core.Files
	CFlags    []string
	Toolchain *Toolchain
}

// Out provides the name of the output created by ObjectFile.
func (obj ObjectFile) Out() core.OutFile {
	return obj.Src.WithExt("o")
}

// BuildSteps provides the steps to produce an ObjectFile.
func (obj ObjectFile) BuildSteps() []core.BuildStep {
	if obj.Toolchain == nil {
		obj.Toolchain = &defaultToolchain
	}

	includes := strings.Builder{}
	for _, include := range obj.Includes {
		includes.WriteString(fmt.Sprintf("-I%s ", include))
	}
	for _, include := range obj.Toolchain.Includes {
		includes.WriteString(fmt.Sprintf("-isystem %s ", include))
	}
	depfile := obj.Src.WithExt("d")
	flags := strings.Join(obj.CFlags, " ")
	cmd := fmt.Sprintf("%s -c -MD -MF %s %s %s -o %s %s", obj.Toolchain.Cc, depfile, flags, includes.String(), obj.Out(), obj.Src)
	return []core.BuildStep{{
		Out:     obj.Out(),
		Depfile: &depfile,
		In:      obj.Src,
		Cmd:     cmd,
		Descr:   fmt.Sprintf("CC %s", obj.Out().RelPath()),
		Alias:   obj.Out().RelPath(),
	}}
}

// Library builds and links a C++ library.
type Library struct {
	Out       core.OutFile
	Srcs      core.Files
	Includes  core.Files
	CFlags    []string
	Toolchain *Toolchain
}

// BuildSteps provides the steps to build a Library.
func (lib Library) BuildSteps() []core.BuildStep {
	if lib.Toolchain == nil {
		lib.Toolchain = &defaultToolchain
	}

	lib.Includes = append(lib.Includes, core.NewInFile("."))

	var steps = []core.BuildStep{}
	var objs = core.Files{}

	for _, src := range lib.Srcs {
		obj := ObjectFile{
			Src:       src,
			Includes:  lib.Includes,
			CFlags:    lib.CFlags,
			Toolchain: lib.Toolchain,
		}
		objs = append(objs, obj.Out())
		steps = append(steps, obj.BuildSteps()...)
	}

	cmd := fmt.Sprintf("%s rv %s %s > /dev/null 2> /dev/null", lib.Toolchain.Ar, lib.Out, objs)
	linkStep := core.BuildStep{
		Out:   lib.Out,
		Ins:   objs,
		Cmd:   cmd,
		Descr: fmt.Sprintf("AR %s", lib.Out.RelPath()),
		Alias: lib.Out.RelPath(),
	}

	return append(steps, linkStep)
}
`,

    "../RULES/core/file.go": `package core

import (
	"fmt"
	"path"
	"strings"
)

// File represents an on-disk file that is either an input to or an output from a BuildStep (or both).
type File interface {
	Path() string
	RelPath() string
	WithExt(ext string) OutFile
	WithSuffix(suffix string) OutFile
}

// Files represents a group of Files.
type Files []File

func (fs Files) String() string {
	paths := []string{}
	for _, f := range fs {
		paths = append(paths, fmt.Sprint(f))
	}
	return strings.Join(paths, " ")
}

// inFile represents a file relative to the workspace source directory.
type inFile struct {
	relPath string
}

// Path returns the file's absolute path.
func (f inFile) Path() string {
	return path.Join(SourceDir(), f.relPath)
}

// RelPath returns the file's path relative to the source directory.
func (f inFile) RelPath() string {
	return f.relPath
}

// WithExt creates an OutFile with the same relative path and the given file extension.
func (f inFile) WithExt(ext string) OutFile {
	return OutFile{f.relPath}.WithExt(ext)
}

// WithSuffix creates an OutFile with the same relative path and the given suffix.
func (f inFile) WithSuffix(suffix string) OutFile {
	return OutFile{f.relPath}.WithSuffix(suffix)
}

func (f inFile) String() string {
	return fmt.Sprintf("\"%s\"", f.Path())
}

// OutFile represents a file relative to the workspace build directory.
type OutFile struct {
	relPath string
}

// Path returns the file's absolute path.
func (f OutFile) Path() string {
	return path.Join(BuildDir(), f.relPath)
}

// RelPath returns the file's path relative to the build directory.
func (f OutFile) RelPath() string {
	return f.relPath
}

// WithExt creates an OutFile with the same relative path and the given file extension.
func (f OutFile) WithExt(ext string) OutFile {
	oldExt := path.Ext(f.relPath)
	relPath := fmt.Sprintf("%s.%s", strings.TrimSuffix(f.relPath, oldExt), ext)
	return OutFile{relPath}
}

// WithSuffix creates an OutFile with the same relative path and the given suffix.
func (f OutFile) WithSuffix(suffix string) OutFile {
	return OutFile{f.relPath + suffix}
}

func (f OutFile) String() string {
	return fmt.Sprintf("\"%s\"", f.Path())
}

// GlobalFile represents a global file.
type GlobalFile interface {
	Path() string
}

type globalFile struct {
	absPath string
}

// Path returns the file's absolute path.
func (f globalFile) Path() string {
	return f.absPath
}

// NewInFile creates an inFile for a file relativ to the source directory.
func NewInFile(p string) File {
	return inFile{p}
}

// NewOutFile creates an OutFile for a file relativ to the build directory.
func NewOutFile(p string) OutFile {
	return OutFile{p}
}

// NewGlobalFile creates a globalFile.
func NewGlobalFile(p string) GlobalFile {
	return globalFile{p}
}
`,

    "../RULES/core/step.go": `package core

import (
	"fmt"
	"strings"
)

// BuildStep represents one build step (i.e., one build command).
// Each BuildStep produces "Out" from "Ins" and "In" by running "Cmd".
type BuildStep struct {
	Out     OutFile
	In      File
	Ins     Files
	Depfile *OutFile
	Cmd     string
	Descr   string
	Alias   string
}

var nextRuleId = 1

func ninjaEscape(s string) string {
	return strings.ReplaceAll(s, " ", "$ ")
}

// Print outputs a Ninja build rule for the BuildStep.
func (step BuildStep) Print() {
	ins := []string{}
	for _, in := range step.Ins {
		ins = append(ins, ninjaEscape(in.Path()))
	}
	if step.In != nil {
		ins = append(ins, ninjaEscape(step.In.Path()))
	}

	alias := ninjaEscape(step.Alias)
	out := ninjaEscape(step.Out.Path())

	fmt.Printf("rule r%d\n", nextRuleId)
	if step.Depfile != nil {
		depfile := ninjaEscape(step.Depfile.Path())
		fmt.Printf("  depfile = %s\n", depfile)
	}
	fmt.Printf("  command = %s\n", step.Cmd)
	if step.Descr != "" {
		fmt.Printf("  description = %s\n", step.Descr)
	}
	fmt.Print("\n")
	fmt.Printf("build %s: r%d %s\n", out, nextRuleId, strings.Join(ins, " "))
	if alias != "" {
		fmt.Print("\n")
		fmt.Printf("build %s: phony %s\n", alias, out)
	}
	fmt.Print("\n\n")

	nextRuleId++
}
`,

    "../RULES/core/util.go": `package core

import (
	"fmt"
	"os"
	"strings"
)

// CurrentTarget holds the current target relative to the workspace directory.
var CurrentTarget string

func SourceDir() string {
	return os.Args[1]
}

func BuildDir() string {
	return os.Args[2]
}

// Flag provides values of build flags.
func Flag(name string) string {
	prefix := fmt.Sprintf("--%s=", name)
	for _, arg := range os.Args[3:] {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

// Fatal can be used in build rules to abort buildfile generation with an error message unconditionally.
func Fatal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "A fatal error occured while processing target '%s': %s", CurrentTarget, msg)
	os.Exit(1)
}

// Assert can be used in build rules to abort buildfile generation with an error message.
func Assert(cond bool, format string, a ...interface{}) {
	if !cond {
		msg := fmt.Sprintf(format, a...)
		fmt.Fprintf(os.Stderr, "Assertion failed while processing target '%s': %s", CurrentTarget, msg)
		os.Exit(1)
	}
}
`,

}
