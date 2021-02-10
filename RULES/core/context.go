package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Context interface {
	AddTarget(name string, target interface{})
	AddBuildStep(BuildStep)
}

type NinjaContext struct {
	nextRuleID int
}

func ninjaEscape(s string) string {
	return strings.ReplaceAll(s, " ", "$ ")
}

type buildable interface {
	Build(ctx Context) OutPath
}

func (ctx *NinjaContext) AddTarget(name string, target interface{}) {
	iface, ok := target.(buildable)
	if !ok {
		return
	}

	currentTarget = name
	out := iface.Build(ctx)

	fmt.Printf("rule r%d\n", ctx.nextRuleID)
	relativePath, _ := filepath.Rel(WorkingDir(), out.Absolute())
	fmt.Printf("  command = echo \"%s\"\n", relativePath)
	fmt.Printf("  description = Created %s:", name)
	fmt.Printf("\n")
	fmt.Printf("build %s: r%d %s\n", name, ctx.nextRuleID, ninjaEscape(out.Absolute()))
	fmt.Printf("\n")
	fmt.Printf("\n")

	ctx.nextRuleID++
}

func (ctx *NinjaContext) AddBuildStep(step BuildStep) {
	ins := []string{}
	for _, in := range step.Ins {
		ins = append(ins, ninjaEscape(in.Absolute()))
	}
	if step.In != nil {
		ins = append(ins, ninjaEscape(step.In.Absolute()))
	}

	out := ninjaEscape(step.Out.Absolute())

	fmt.Printf("rule r%d\n", ctx.nextRuleID)
	if step.Depfile != nil {
		depfile := ninjaEscape(step.Depfile.Absolute())
		fmt.Printf("  depfile = %s\n", depfile)
	}
	fmt.Printf("  command = %s\n", step.Cmd)
	if step.Descr != "" {
		fmt.Printf("  description = %s\n", step.Descr)
	}
	fmt.Print("\n")
	fmt.Printf("build %s: r%d %s\n", out, ctx.nextRuleID, strings.Join(ins, " "))
	fmt.Print("\n\n")

	ctx.nextRuleID++
}

type ListTargetsContext struct{}

func (ctx *ListTargetsContext) AddTarget(name string, target interface{}) {
	_, ok := target.(buildable)
	if ok {
		fmt.Println(name)
	}
}

func (ctx *ListTargetsContext) AddBuildStep(step BuildStep) {}