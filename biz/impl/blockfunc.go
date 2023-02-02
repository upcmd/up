package impl

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/upcmd/up/model/core"
	u "github.com/upcmd/up/utils"
)

type BlockFuncAction struct {
	Do        interface{}
	Vars      *core.Cache
	Tasknames []string
	Steps     *Steps
}

func (f *BlockFuncAction) Adapt() {
	var flowname string
	var flow Steps

	switch f.Do.(type) {
	case string:
		//a flow name + refdir to load the flow
		raw := f.Do.(string)
		flowname = Render(raw, f.Vars)
		u.P(flowname)

	case []interface{}:
		err := ms.Decode(f.Do, &flow)
		u.LogErrorAndPanic("load steps", err, "steps has configuration problem, please fix it")

	default:
		u.LogWarn("Block func", "Not implemented or void for no action!")
	}

	f.Steps = &flow
}

func (f *BlockFuncAction) Exec() {
	BlockFlowRun(f.Steps, f.Vars)
}

func BlockFlowRun(flow *Steps, execVars *core.Cache) {
	rtContext := BlockRuntimeContext{
		BlockBaseVars: execVars,
	}
	BlockStack().Push(&rtContext)

	flow.Exec(true)
	BlockStack().Pop()
}
