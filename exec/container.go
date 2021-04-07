/*
 * Copyright 1999-2019 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package exec

import (
	"context"
	"fmt"
	"time"

	"github.com/chaosblade-io/chaosblade-spec-go/spec"
	"github.com/chaosblade-io/chaosblade-spec-go/util"
)

const (
	ForceFlag = "force"
)

type ContainerCommandModelSpec struct {
	spec.BaseExpModelCommandSpec
}

func NewContainerCommandSpec() spec.ExpModelCommandSpec {
	return &ContainerCommandModelSpec{
		spec.BaseExpModelCommandSpec{
			ExpActions: []spec.ExpActionCommandSpec{
				NewRemoveActionCommand(),
			},
			ExpFlags: []spec.ExpFlagSpec{},
		},
	}
}

func (cms *ContainerCommandModelSpec) Name() string {
	return "container"
}

func (cms *ContainerCommandModelSpec) ShortDesc() string {
	return `Execute a docker experiment`
}

func (cms *ContainerCommandModelSpec) LongDesc() string {
	return `Execute a docker experiment. The local host must be installed docker command.`
}

type removeActionCommand struct {
	spec.BaseExpActionCommandSpec
}

func NewRemoveActionCommand() spec.ExpActionCommandSpec {
	return &removeActionCommand{
		spec.BaseExpActionCommandSpec{
			ActionMatchers: []spec.ExpFlagSpec{},
			ActionFlags: []spec.ExpFlagSpec{
				&spec.ExpFlag{
					Name:   ForceFlag,
					Desc:   "force remove",
					NoArgs: true,
				},
			},
			ActionExecutor: &removeActionExecutor{},
			ActionExample: `# Delete the container id that is a76d53933d3f",
blade create docker container remove --container-id a76d53933d3f`,
		},
	}
}

func (*removeActionCommand) Name() string {
	return "remove"
}

func (*removeActionCommand) Aliases() []string {
	return []string{"rm"}
}

func (*removeActionCommand) ShortDesc() string {
	return "remove a container"
}

func (r *removeActionCommand) LongDesc() string {
	if r.ActionLongDesc != "" {
		return r.ActionLongDesc
	}
	return "remove a container"
}

type removeActionExecutor struct {
}

func (*removeActionExecutor) Name() string {
	return "remove"
}

func (e *removeActionExecutor) SetChannel(channel spec.Channel) {
}

func (e *removeActionExecutor) Exec(uid string, ctx context.Context, model *spec.ExpModel) *spec.Response {
	flags := model.ActionFlags
	client, err := GetClient(flags[EndpointFlag.Name])
	if err != nil {
		util.Errorf(uid, util.GetRunFuncName(), fmt.Sprintf(spec.ResponseErr[spec.DockerExecFailed].ErrInfo, "GetClient", err.Error()))
		return spec.ResponseFail(spec.DockerExecFailed, fmt.Sprintf(spec.ResponseErr[spec.DockerExecFailed].ErrInfo, "GetClient", err.Error()))
	}
	containerId := flags[ContainerIdFlag.Name]
	if containerId == "" {
		util.Errorf(uid, util.GetRunFuncName(), fmt.Sprintf(spec.ResponseErr[spec.ParameterLess].ErrInfo, ContainerIdFlag.Name))
		return spec.ResponseFailWaitResult(spec.ParameterLess, fmt.Sprintf(spec.ResponseErr[spec.ParameterLess].Err, ContainerIdFlag.Name),
			fmt.Sprintf(spec.ResponseErr[spec.ParameterLess].ErrInfo, ContainerIdFlag.Name))
	}
	if _, ok := spec.IsDestroy(ctx); ok {
		return spec.ReturnSuccess(uid)
	}
	if _, err, code := client.getContainerById(containerId); err != nil {
		util.Errorf(uid, util.GetRunFuncName(), err.Error())
		return spec.ResponseFail(code, err.Error())
	}

	forceFlag := flags[ForceFlag]
	if forceFlag != "" {
		timeout := time.Second
		err = client.stopAndRemoveContainer(containerId, &timeout)
	} else {
		err = client.forceRemoveContainer(containerId)
	}
	if err != nil {
		util.Errorf(uid, util.GetRunFuncName(), fmt.Sprintf(spec.ResponseErr[spec.DockerExecFailed].ErrInfo, "ContainerRemove", err.Error()))
		return spec.ResponseFail(spec.DockerExecFailed, fmt.Sprintf(spec.ResponseErr[spec.DockerExecFailed].ErrInfo, "ContainerRemove", err.Error()))
	}
	return spec.ReturnSuccess(uid)
}
