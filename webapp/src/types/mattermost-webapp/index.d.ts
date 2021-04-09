import React from 'react';

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType);

    registerReducer(reducer: React.Reducer);

    unregisterComponent(componentId: string);

    registerWebSocketEventHandler(event, handler);

    registerReconnectHandler(handler);

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
