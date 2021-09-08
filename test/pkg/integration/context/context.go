// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package context

import (
	"context"

	"github.com/gitpod-io/gitpod/test/pkg/integration"
)

const (
	componentAPI = "component-api"
)

type contextKey string

func GetComponentAPI(ctx context.Context) *integration.ComponentAPI {
	return ctx.Value(contextKey(componentAPI)).(*integration.ComponentAPI)
}

func SetComponentAPI(ctx context.Context, api *integration.ComponentAPI) context.Context {
	return context.WithValue(ctx, contextKey(componentAPI), api)
}
