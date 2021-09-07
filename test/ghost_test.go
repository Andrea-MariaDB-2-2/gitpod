// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
	wsmanapi "github.com/gitpod-io/gitpod/ws-manager/api"
)

func TestGhostWorkspace(t *testing.T) {
	ghostWorkspace := features.New("ghost").
		WithLabel("component", "ws-manager").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can start a ghost workspace", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			// there's nothing specific about ghost that we want to test beyond that they start properly
			ws, err := integration.LaunchWorkspaceDirectly(ctx, api, integration.WithRequestModifier(func(req *wsmanapi.StartWorkspaceRequest) error {
				req.Type = wsmanapi.WorkspaceType_GHOST
				req.Spec.Envvars = append(req.Spec.Envvars, &wsmanapi.EnvironmentVariable{
					Name:  "GITPOD_TASKS",
					Value: `[{ "init": "echo \"some output\" > someFile; sleep 20; exit 0;" }]`,
				})
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}

			ctx = context.WithValue(ctx, workspaceKey(workspaceIDKey), ws.Req.Id)

			_, err = integration.WaitForWorkspaceStart(ctx, ws.Req.Id, api)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			wsID := ctx.Value(workspaceKey(workspaceIDKey))
			if wsID == nil {
				return ctx
			}

			err := integration.DeleteWorkspace(ctx, api, wsID.(string))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, ghostWorkspace)
}
