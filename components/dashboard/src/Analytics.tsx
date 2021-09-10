/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { getGitpodService } from "./service/service";
import { log } from "@gitpod/gitpod-protocol/lib/util/logging";

export type Event = "invite_url_requested" | "organisation_authorised";

export type TrackingMsg = {
    dnt?: boolean,
    path: string,
    button_type: string,
    label?: string,
    destination?: string
  }

//call this to track all events outside of button and anchor clicks
export const trackEvent = (event: Event, properties: any) => {
    getGitpodService().server.trackEvent({
        event: event,
        properties: properties
    })
}

export const trackButtonOrAnchor = (target: HTMLAnchorElement | HTMLButtonElement) => {
     //read manually passed analytics props from 'data-analytics' attribute of event target
     let passedProps;
     if (target.dataset.analytics) {
       try{
         passedProps = JSON.parse(target.dataset.analytics);
         if (passedProps.dnt) {
           return;
         }
       } catch(error) {
           log.debug(error);
       }

     }

     let trackingMsg: TrackingMsg = {
       path: window.location.pathname,
       button_type: "primary" //primary button is the default if secondary is not specified
     };

     if (target instanceof HTMLButtonElement) {
       //parse button data
       const button = target as HTMLButtonElement;
       trackingMsg.label = button.textContent || undefined;
       if (button.classList.contains("secondary")) {
         trackingMsg.button_type = "secondary";
       }
       //retrieve href if parent is an anchor element
       if (button.parentElement instanceof HTMLAnchorElement) {
         const anchor = button.parentElement as HTMLAnchorElement;
         trackingMsg.destination = anchor.href;
       }
     }

     if (target instanceof HTMLAnchorElement) {
       const anchor = target as HTMLAnchorElement;
       trackingMsg.label = anchor.textContent || undefined
       trackingMsg.destination = anchor.href;
     }

     if (passedProps) {
       trackingMsg.button_type = passedProps.button_type || trackingMsg.button_type;
       trackingMsg.destination = passedProps.destination || trackingMsg.destination;
       trackingMsg.label = passedProps.label || trackingMsg.label;
       trackingMsg.path = passedProps.path || trackingMsg.path;
     }

     getGitpodService().server.trackEvent({
         event: "dashboard_clicked",
         properties: trackingMsg
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