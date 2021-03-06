/*
Copyright 2017 The Kubernetes Authors.

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

// Package app does all of the work necessary to create a Kubernetes
// APIServer by binding together the API, master and APIServer infrastructure.
// It can be configured and called directly or via the hyperkube framework.
package app

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	apiextensionsapiserver "k8s.io/kube-apiextensions-server/pkg/apiserver"
	apiextensionscmd "k8s.io/kube-apiextensions-server/pkg/cmd/server"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
)

func createAPIExtensionsConfig(kubeAPIServerConfig genericapiserver.Config, commandOptions *options.ServerRunOptions) (*apiextensionsapiserver.Config, error) {
	// make a shallow copy to let us twiddle a few things
	// most of the config actually remains the same.  We only need to mess with a couple items related to the particulars of the apiextensions
	genericConfig := kubeAPIServerConfig

	// the apiextensions doesn't wire these up.  It just delegates them to the kubeapiserver
	genericConfig.EnableSwaggerUI = false

	// TODO these need to be sorted out.  There's an issue open
	genericConfig.OpenAPIConfig = nil
	genericConfig.SwaggerConfig = nil

	// copy the loopbackclientconfig.  We're going to change the contenttype back to json until we get protobuf serializations for it
	t := *kubeAPIServerConfig.LoopbackClientConfig
	genericConfig.LoopbackClientConfig = &t
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = ""

	// copy the etcd options so we don't mutate originals.
	etcdOptions := *commandOptions.Etcd
	etcdOptions.StorageConfig.Codec = apiextensionsapiserver.Codecs.LegacyCodec(schema.GroupVersion{Group: "apiextensions.k8s.io", Version: "v1beta1"})
	etcdOptions.StorageConfig.Copier = apiextensionsapiserver.Scheme
	genericConfig.RESTOptionsGetter = &genericoptions.SimpleRestOptionsFactory{Options: etcdOptions}

	apiextensionsConfig := &apiextensionsapiserver.Config{
		GenericConfig:        &genericConfig,
		CRDRESTOptionsGetter: apiextensionscmd.NewCRDRESTOptionsGetter(etcdOptions),
	}

	return apiextensionsConfig, nil

}

func createAPIExtensionsServer(apiextensionsConfig *apiextensionsapiserver.Config, delegateAPIServer genericapiserver.DelegationTarget) (*apiextensionsapiserver.CustomResourceDefinitions, error) {
	apiextensionsServer, err := apiextensionsConfig.Complete().New(delegateAPIServer)
	if err != nil {
		return nil, err
	}

	return apiextensionsServer, nil
}
