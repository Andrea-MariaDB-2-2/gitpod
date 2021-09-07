// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
	"github.com/gitpod-io/gitpod/test/tests/workspace/common"
)

type ContextTest struct {
	Skip               bool
	Name               string
	ContextURL         string
	WorkspaceRoot      string
	ExpectedBranch     string
	ExpectedBranchFunc func(username string) string
}

func TestGitHubContexts(t *testing.T) {
	tests := []ContextTest{
		{
			Name:           "open repository",
			ContextURL:     "github.com/gitpod-io/gitpod",
			WorkspaceRoot:  "/workspace/gitpod",
			ExpectedBranch: "main",
		},
		{
			Name:           "open branch",
			ContextURL:     "github.com/gitpod-io/gitpod-test-repo/tree/integration-test-1",
			WorkspaceRoot:  "/workspace/gitpod-test-repo",
			ExpectedBranch: "integration-test-1",
		},
		{
			Name:          "open issue",
			ContextURL:    "github.com/gitpod-io/gitpod-test-repo/issues/88",
			WorkspaceRoot: "/workspace/gitpod-test-repo",
			ExpectedBranchFunc: func(username string) string {
				return fmt.Sprintf("%s/integration-tests-test-context-88", username)
			},
		},
		{
			Name:           "open tag",
			ContextURL:     "github.com/gitpod-io/gitpod-test-repo/tree/integration-test-context-tag",
			WorkspaceRoot:  "/workspace/gitpod-test-repo",
			ExpectedBranch: "HEAD",
		},
	}
	runContextTests(t, tests)
}

func TestGitLabContexts(t *testing.T) {
	tests := []ContextTest{
		{
			Name:           "open repository",
			ContextURL:     "gitlab.com/AlexTugarev/gp-test",
			WorkspaceRoot:  "/workspace/gp-test",
			ExpectedBranch: "master",
		},
		{
			Name:           "open branch",
			ContextURL:     "gitlab.com/AlexTugarev/gp-test/tree/wip",
			WorkspaceRoot:  "/workspace/gp-test",
			ExpectedBranch: "wip",
		},
		{
			Name:          "open issue",
			ContextURL:    "gitlab.com/AlexTugarev/gp-test/issues/1",
			WorkspaceRoot: "/workspace/gp-test",
			ExpectedBranchFunc: func(username string) string {
				return fmt.Sprintf("%s/write-a-readme-1", username)
			},
		},
		{
			Name:           "open tag",
			ContextURL:     "gitlab.com/AlexTugarev/gp-test/merge_requests/2",
			WorkspaceRoot:  "/workspace/gp-test",
			ExpectedBranch: "wip2",
		},
	}
	runContextTests(t, tests)
}

func runContextTests(t *testing.T, tests []ContextTest) {
	contextFeat := features.New("context").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("should run context tests", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			for _, test := range tests {
				t.Run(test.ContextURL, func(t *testing.T) {
					if test.Skip {
						t.SkipNow()
					}
					// TODO(geropl) Why does this not work? Logs hint to a race around bucket creation...?
					// t.Parallel()

					username := ctx.Value("username").(string)

					if username == "" && test.ExpectedBranchFunc != nil {
						t.Logf("skipping '%s' because there is not username configured", test.Name)
						t.SkipNow()
					}

					nfo, stopWS, err := integration.LaunchWorkspaceFromContextURL(ctx, test.ContextURL, api)
					if err != nil {
						t.Fatal(err)
					}
					defer stopWS(false) // we do not wait for stopped here as it does not matter for this test case and speeds things up

					wctx, wcancel := context.WithTimeout(ctx, 1*time.Minute)
					defer wcancel()

					_, err = integration.WaitForWorkspaceStart(wctx, nfo.LatestInstance.ID, api)
					if err != nil {
						t.Fatal(err)
					}

					rsa, closer, err := integration.Instrument(integration.ComponentWorkspace, "workspace", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.LatestInstance.ID))
					if err != nil {
						t.Fatal(err)
					}
					defer rsa.Close()
					defer dispose(t, closer)

					// get actual from workspace
					git := common.Git(rsa)
					actBranch, err := git.GetBranch(test.WorkspaceRoot)
					if err != nil {
						t.Fatal(err)
					}

					expectedBranch := test.ExpectedBranch
					if test.ExpectedBranchFunc != nil {
						expectedBranch = test.ExpectedBranchFunc(username)
					}
					if actBranch != expectedBranch {
						t.Fatalf("expected branch '%s', got '%s'!", test.ExpectedBranch, actBranch)
					}
				})
			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testenv.Test(t, contextFeat)
}
