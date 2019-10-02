// +build tools

package pkg

import (
	_ "github.com/emicklei/go-restful"
	_ "github.com/onsi/ginkgo"
	_ "github.com/onsi/gomega"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/code-generator/cmd/client-gen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "sigs.k8s.io/controller-runtime/pkg/client/config"
	_ "sigs.k8s.io/controller-runtime/pkg/controller"
	_ "sigs.k8s.io/controller-runtime/pkg/handler"
	_ "sigs.k8s.io/controller-runtime/pkg/manager"
	_ "sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	_ "sigs.k8s.io/controller-runtime/pkg/source"
	_ "sigs.k8s.io/testing_frameworks/integration"
	_ "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	_ "github.com/vektra/mockery"
)
