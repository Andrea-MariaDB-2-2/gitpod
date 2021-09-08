// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
	var (
		namespace = flag.String("namespace", "default", `namespace to execute the test against. Defaults to the one configured in "kubeconfig".`)
		username  = flag.String("username", "", "username to execute the tests with. Chooses one automatically if left blank.")
	)

	klog.InitFlags(nil)
	flag.Parse()

	restConfig, err := conf.New(conf.ResolveKubeConfigFile())
	if err != nil {
		klog.Fatalf("unexpected error: %v", err)
	}

	// change defaults to avoid limiting connections
	restConfig.QPS = 20
	restConfig.Burst = 50

	client, err := klient.New(restConfig)
	if err != nil {
		klog.Fatalf("unexpected error: %v", err)
	}

	conf := envconf.New()
	conf.WithClient(client)
	conf.WithNamespace(*namespace)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "username", *username)

	testenv, err = env.NewWithContext(ctx, conf)
	if err != nil {
		klog.Fatalf("unexpected error: %v", err)
	}
	testenv.Setup(
		checkGitpodIsRunning(),
	)
	testenv.Finish(
		finish(),
	)

	os.Exit(testenv.Run(m))
}

func checkGitpodIsRunning() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}

func finish() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return ctx, nil
	}
}
