/*
Copyright 2018 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ctc_lib

import (
	"errors"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/constants"

	"github.com/GoogleCloudPlatform/runtimes-common/ctc_lib/util"
	"github.com/spf13/cobra"
)

type ContainerToolListCommand struct {
	*ContainerToolCommandBase
	OutputList      []interface{}
	SummaryObject   interface{}
	SummaryTemplate string
	// RunO Executes cobra.Command.Run and returns an List[Output]
	RunO func(command *cobra.Command, args []string) ([]interface{}, error)
	// When defined, StreamO Executes cobra.Command.Run and streams each item in the List as its added.
	// This will ignore the RunO function.
	StreamO func(command *cobra.Command, args []string)
	// This function will execute execute over the output and return a Summary Object which can be printed.
	TotalO func(list []interface{}) (interface{}, error)
	Stream chan interface{}
}

func (commandList ContainerToolListCommand) ReadFromStream() ([]interface{}, error) {
	results := make([]interface{}, 0)
	for obj := range commandList.Stream {
		results = append(results, obj)
		if _, ok := obj.(string); ok {
			// Display any Arbitary strings written to the channel as is.
			// These could be headers or any text.
			// TODO: Provide a callback for users to overwrite this default behavior.
			util.ExecuteTemplate(constants.EmptyTemplate,
				obj, commandList.TemplateFuncMap, commandList.OutOrStdout())
			continue
		}
		err := util.ExecuteTemplate(commandList.ReadTemplateFromFlagOrCmdDefault(),
			obj, commandList.TemplateFuncMap, commandList.OutOrStdout())
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (commandList ContainerToolListCommand) isRunODefined() bool {
	return commandList.RunO != nil || commandList.StreamO != nil
}

func (ctc *ContainerToolListCommand) ValidateCommand() error {
	if (ctc.Run != nil || ctc.RunE != nil) && ctc.isRunODefined() {
		return errors.New(`Cannot provide both Command.Run and RunO implementation.
Either implement Command.Run implementation or RunO implemetation`)
	}
	return nil
}

func (ctc *ContainerToolListCommand) printO(c *cobra.Command, args []string) error {
	var err error
	if ctc.StreamO != nil {
		// Stream Objects
		ctc.StreamO(c, args)
		ctc.OutputList, err = ctc.ReadFromStream()
	} else {
		// Run RunO function.
		ctc.OutputList, err = ctc.RunO(c, args)
		if err != nil {
			return err
		}
		err = util.ExecuteTemplate(ctc.ReadTemplateFromFlagOrCmdDefault(),
			ctc.OutputList, ctc.TemplateFuncMap, ctc.OutOrStdout())
	}
	if err != nil {
		return err
	}
	// If TotalO function defined and Summary Template provided, print the summary.
	if ctc.TotalO != nil && ctc.SummaryTemplate != "" {
		total, err := ctc.TotalO(ctc.OutputList)
		if err != nil {
			return err
		}
		ctc.SummaryObject = total
		return util.ExecuteTemplate(ctc.SummaryTemplate, ctc.SummaryObject,
			ctc.TemplateFuncMap, ctc.OutOrStdout())
	}
	return nil
}
