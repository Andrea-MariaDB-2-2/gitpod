/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { getGitpodService } from "./service/service";

//contexts from which calls are made in dashboard
export type dashboardContexts = "menu" | "/<team_name>/<project_name>/configure" | "/new" | "/<team_name>/<project_name>/prebuilds" | "/<team_name>/<project_name>" | "/projects" | "/<team_name>/members" | "/teams/new" | "/workspaces" | "/<team_name>/projects";
//buttons that are tracked in dashboard
export type buttons = "new_team" | "run_prebuild" | "add_organisation" | "select_git_provider" | "select_project" | "select_team" | "continue_with_github" | "continue_with_gitlab" | "create_team" | "trigger_prebuild" | "new_workspace" | "rerun_prebuild" | "new_project" | "invite_members" | "remove_project" | "leave_team";
//position of tracked button in page
export type buttonContexts = "dropdown" | "primary_button" | "secondary_button" | "kebab_menu" | "card";
//events are than generic button clicks that are tracked in dashboard
export type events = "invite_url_requested" | "workspace_new_clicked" | "workspace_button_clicked" | "organisation_authorised";
//actions that can be performed on workspaces in dashboard
export type workspaceActions = "open" | "stop" | "download" | "share" | "pin" | "delete";

//call this when a button in the dashboard is clicked
export const trackButton = (dashboard_context: dashboardContexts, button: buttons, button_context: buttonContexts) => {
    getGitpodService().server.trackEvent({
        event: "dashboard_button_clicked",
        properties: {
            dashboard_context: dashboard_context,
            button: button,
            button_context: button_context
        }
    })
}

//call this when a button that performs a certain action on a workspace is clicked
export const trackWorkspaceButton = (workspaceId: string, workspace_action: workspaceActions, button_context: buttonContexts, state: string) => {
    getGitpodService().server.trackEvent({
        event: "workspace_button_clicked",
        properties: {
            workspaceId: workspaceId,
            workspace_action: workspace_action,
            button_context: button_context,
            state: state
        }
    })
}

//call this when anything that is not a button or a page call should be tracked
export const trackEvent = (event: events, properties: any) => {
    getGitpodService().server.trackEvent({
        event: event,
        properties: properties
    })
}

//call this when the path changes. Complete page call is unnecessary for SPA after initial call
export const trackPathChange = (path: string) => {
    getGitpodService().server.trackEvent({
        event: "path_changed",
        properties: {
            path: path
        }
    });
}

//call this to record a page call if the user is known or record the page info for a later call if the user is anonymous
export const trackLocation = async (userKnown: boolean) => {
    const w = window as any;
    if (!w._gp.trackLocation) {
        //set _gp.trackLocation on first visit
        w._gp.trackLocation = {
            locationTracked: false,
            properties: {
                referrer: document.referrer,
                path: window.location.pathname,
                host: window.location.hostname,
                url: window.location.href
            }
        };
    } else if (w._gp.trackLocation.locationTracked) {
        return; //page call was already recorded earlier
    }

    if (userKnown) {
        //if the user is known, make server call
        getGitpodService().server.trackLocation({
            properties: w._gp.trackLocation.properties
        });
        w._gp.locationTracked = true;
        delete w._gp.locationTracked.properties;
    }
}