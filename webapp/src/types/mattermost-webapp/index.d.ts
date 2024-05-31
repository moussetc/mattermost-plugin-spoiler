/* eslint-disable @typescript-eslint/no-explicit-any */
import React from 'react';

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType);

    registerReducer(reducer: React.Reducer);

    unregisterComponent(componentId: string);

    registerWebSocketEventHandler(event: string, handler: (msg: any) => void)

    registerReconnectHandler(handler);

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
