/*
Copyright 2016 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/tables"
)

type GetFederationOptions struct {
}

func init() {
	var options GetFederationOptions

	cmd := &cobra.Command{
		Use:     "federations",
		Aliases: []string{"federation"},
		Short:   "get federations",
		Long:    `List or get federations.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGetFederations(&rootCommand, os.Stdout, &options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	getCmd.cobraCommand.AddCommand(cmd)
}

func RunGetFederations(context Factory, out io.Writer, options *GetFederationOptions) error {
	client, err := context.Clientset()
	if err != nil {
		return err
	}

	list, err := client.Federations().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var federations []*api.Federation
	for i := range list.Items {
		federations = append(federations, &list.Items[i])
	}
	if len(federations) == 0 {
		fmt.Fprintf(out, "No federations found\n")
		return nil
	}
	switch getCmd.output {

	case OutputTable:

		t := &tables.Table{}
		t.AddColumn("NAME", func(f *api.Federation) string {
			return f.ObjectMeta.Name
		})
		t.AddColumn("CONTROLLERS", func(f *api.Federation) string {
			return strings.Join(f.Spec.Controllers, ",")
		})
		t.AddColumn("MEMBERS", func(f *api.Federation) string {
			return strings.Join(f.Spec.Members, ",")
		})
		return t.Render(federations, out, "NAME", "CONTROLLERS", "MEMBERS")

	case OutputYaml:
		for i, f := range federations {
			if i != 0 {
				_, err = out.Write([]byte("\n\n---\n\n"))
				if err != nil {
					return fmt.Errorf("error writing to stdout: %v", err)
				}
			}
			if err := marshalToWriter(f, marshalYaml, os.Stdout); err != nil {
				return err
			}
		}
	case OutputJSON:
		for _, f := range federations {
			if err := marshalToWriter(f, marshalJSON, os.Stdout); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unknown output format: %q", getCmd.output)
	}
	return nil
}
