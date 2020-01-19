/*
Copyright The Kubepack Authors.

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
	"io/ioutil"
	"net/http"
	"strings"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"sigs.k8s.io/yaml"
)

func main() {
	var repos = map[string]string{}

	resp, err := http.Get("https://raw.githubusercontent.com/helm/hub/master/repos.yaml")
	if err == nil {
		defer resp.Body.Close()
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			var hub v1alpha1.Hub
			err = yaml.Unmarshal(data, &hub)
			if err == nil {
				for _, repo := range hub.Repositories {
					repos[strings.TrimSuffix(repo.URL, "/")] = repo.Name
				}
			}
		}
	}

	fmt.Println(repos)
}
