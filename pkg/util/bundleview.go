package util

import (
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"

	"github.com/gobuffalo/flect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateBundleView(url, name, version string) (*v1alpha1.BundleView, error) {
	view, err := toBundleOptionView(&v1alpha1.BundleOption{
		BundleRef: v1alpha1.BundleRef{
			URL:  url,
			Name: name,
		},
		Version: version,
	}, 0)
	if err != nil {
		return nil, err
	}
	bv := v1alpha1.BundleView{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "BundleView",
		},
		BundleOptionView: *view,
	}
	return &bv, nil
}

func toBundleOptionView(in *v1alpha1.BundleOption, level int) (*v1alpha1.BundleOptionView, error) {
	chrt, bundle := GetBundle(in)

	bv := v1alpha1.BundleOptionView{
		PackageMeta: v1alpha1.PackageMeta{
			Name:              chrt.Name(),
			URL:               in.URL,
			Version:           chrt.Metadata.Version,
			PackageDescriptor: GetPackageDescriptor(chrt),
		},
		DisplayName: XorY(bundle.Spec.DisplayName, flect.Titleize(flect.Humanize(bundle.Name))),
	}

	for _, pkg := range bundle.Spec.Packages {
		if pkg.Chart != nil {
			var chartVersion string
			for _, v := range pkg.Chart.Versions {
				if v.Selected {
					chartVersion = v.Version
					break
				}
			}
			if chartVersion == "" {
				chartVersion = pkg.Chart.Versions[0].Version
			}
			pkgChart, err := GetChart(pkg.Chart.URL, pkg.Chart.Name, chartVersion)
			if err != nil {
				return nil, err
			}

			required := pkg.Chart.Required
			if level > 0 {
				// Only direct charts can be required
				required = false
			}

			card := v1alpha1.PackageCard{
				Chart: &v1alpha1.ChartCard{
					ChartRef: v1alpha1.ChartRef{
						Name: pkg.Chart.Name,
						URL:  pkg.Chart.URL,
					},
					PackageDescriptor: GetPackageDescriptor(pkgChart.Chart),
					Features:          pkg.Chart.Features,
					MultiSelect:       pkg.Chart.MultiSelect,
					Namespace:         XorY(pkg.Chart.Namespace, bundle.Spec.Namespace),
					Required:          required,
					Selected:          pkg.Chart.Required,
				},
			}
			for _, v := range pkg.Chart.Versions {
				card.Chart.Versions = append(card.Chart.Versions, v.VersionOption)
			}
			if len(card.Chart.Versions) == 1 {
				card.Chart.Versions[0].Selected = true
			}
			bv.Packages = append(bv.Packages, card)
		} else if pkg.Bundle != nil {
			view, err := toBundleOptionView(pkg.Bundle, level+1)
			if err != nil {
				return nil, err
			}
			bv.Packages = append(bv.Packages, v1alpha1.PackageCard{
				Bundle: view,
			})
		} else if pkg.OneOf != nil {
			bovs := make([]*v1alpha1.BundleOptionView, 0, len(pkg.OneOf.Bundles))
			for _, bo := range pkg.OneOf.Bundles {
				view, err := toBundleOptionView(bo, level+1)
				if err != nil {
					return nil, err
				}
				bovs = append(bovs, view)
			}
			bv.Packages = append(bv.Packages, v1alpha1.PackageCard{
				OneOf: &v1alpha1.OneOfBundleOptionView{
					Description: pkg.OneOf.Description,
					Bundles:     bovs,
				},
			})
		}
	}

	return &bv, nil
}
