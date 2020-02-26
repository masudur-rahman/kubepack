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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/artifacts/products"
	"kubepack.dev/kubepack/pkg/lib"

	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
	"sigs.k8s.io/yaml"
)

func main() {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())

	m.Get("/products", func(ctx *macaron.Context) {
		// /products

		phase := ctx.Params(":phase")
		var out v1alpha1.ProductList
		for _, filename := range products.AssetNames() {
			data, err := products.Asset(filename)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			var p v1alpha1.Product
			err = json.Unmarshal(data, &p)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			if phase == "" || p.Spec.Phase == v1alpha1.Phase(phase) {
				out.Items = append(out.Items, p)
			}
		}
		ctx.JSON(http.StatusOK, out)
	})

	m.Get("/products/:owner/:key", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		data, err := products.Asset(ctx.Params(":key") + ".json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.Write(data)
	})

	m.Get("/products/:owner/:key/plans", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		dir := "artifacts/products/kubedb-plans"
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		var out v1alpha1.PlanList
		for _, file := range files {
			data, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			var plaan v1alpha1.Plan
			err = json.Unmarshal(data, &plaan)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out.Items = append(out.Items, plaan)
		}
		ctx.JSON(http.StatusOK, out)
	})

	m.Get("/products/:owner/:key/plans/:plan", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		dir := "artifacts/products/" + ctx.Params(":key") + "-plans"

		data, err := ioutil.ReadFile(filepath.Join(dir, ctx.Params(":plan")+".json"))
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var plaan v1alpha1.Plan
		err = json.Unmarshal(data, &plaan)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, plaan)
	})

	m.Get("/products/:owner/:key/plans/:plan/bundle", func(ctx *macaron.Context) {
		// /products/appscode/kubedb
		// TODO: get product by (owner, key)

		// product
		data, err := products.Asset(ctx.Params(":key") + ".json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var p v1alpha1.Product
		err = json.Unmarshal(data, &p)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		// plan
		dir := "artifacts/products/" + ctx.Params(":key") + "-plans"
		data, err = ioutil.ReadFile(filepath.Join(dir, ctx.Params(":plan")+".json"))
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		var plaan v1alpha1.Plan
		err = json.Unmarshal(data, &plaan)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		bv, err := lib.CreateBundleViewForBundle(&v1alpha1.ChartRepoRef{
			URL:     plaan.Spec.Bundle.URL,
			Name:    plaan.Spec.Bundle.Name,
			Version: p.Spec.LatestVersion,
		})
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, bv)
	})

	m.Get("/product_id/:id", func(ctx *macaron.Context) {
		// TODO: get product by id

		data, err := products.Asset("kubedb.json")
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.Write(data)
	})

	m.Group("/deploy/orders", func() {
		m.Post("", binding.Json(v1alpha1.BundleView{}), func(ctx *macaron.Context, params v1alpha1.BundleView) {
			order, err := lib.CreateOrder(params)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			data, err := yaml.Marshal(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			err = lib.Upload(string(order.UID), "order.yaml", data)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			ctx.JSON(http.StatusOK, order)
		})
		m.Get("/:id/helm2", func(ctx *macaron.Context) {
			data, err := lib.Download(ctx.Params(":id"), "order.yaml")
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			script, err := lib.GenerateHelm2Script(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, map[string]string{
				"linux":  script,
				"darwin": script,
			})
		})
		m.Get("/:id/helm3", func(ctx *macaron.Context) {
			data, err := lib.Download(ctx.Params(":id"), "order.yaml")
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			script, err := lib.GenerateHelm3Script(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, map[string]string{
				"linux":  script,
				"darwin": script,
			})
		})
		m.Get("/:id/yaml", func(ctx *macaron.Context) {
			data, err := lib.Download(ctx.Params(":id"), "order.yaml")
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			var order v1alpha1.Order
			err = yaml.Unmarshal(data, &order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			script, err := lib.GenerateYAMLScript(order)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}

			ctx.JSON(http.StatusOK, map[string]string{
				"linux":  script,
				"darwin": script,
			})
		})
		m.Post("/:id/clusters/:cluster", func(ctx *macaron.Context) {
		})
		m.Delete("/:id/clusters/:cluster", func(ctx *macaron.Context) {

		})
	})
	m.Get("/bundleview", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
		// TODO: verify params

		bv, err := lib.CreateBundleViewForChart(&params)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, bv)
	})
	m.Get("/packageview", binding.Json(v1alpha1.ChartRepoRef{}), func(ctx *macaron.Context, params v1alpha1.ChartRepoRef) {
		// TODO: verify params

		chrt, err := lib.GetChart(params.URL, params.Name, params.Version)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		pv, err := lib.CreatePackageView(params.URL, chrt.Chart)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, pv)
	})
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Run()
}