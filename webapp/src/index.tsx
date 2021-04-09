import {Store, Action} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';

import SpoilerPostType from './components/spoiler_post_type';
import {
    fetchPluginConfig,
    pluginConfigChange,
} from './actions';
import {spoilerMode} from './selectors';
import reducer from './reducer';
import manifest from './manifest';

// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerReducer(reducer);

        let spoilerPostTypeId: string;
        let currentSpoilerMode: string;

        // Watch for plugin configuration changes in the state
        store.subscribe(() => {
            const previousSpoilerMode = currentSpoilerMode;
            currentSpoilerMode = spoilerMode(store.getState());

            // Only define the custom post type display if the configuration requires it
            if (previousSpoilerMode !== currentSpoilerMode) {
                if (currentSpoilerMode === 'redacted') {
                    spoilerPostTypeId = registry.registerPostTypeComponent('custom_spoiler', SpoilerPostType);
                } else if (spoilerPostTypeId) {
                    registry.unregisterComponent(spoilerPostTypeId);
                }
            }
        });

        // Immediately fetch the current plugin config
        await fetchPluginConfig(store.dispatch, store.getState);

        // Be alerted if the plugin configuration change
        registry.registerWebSocketEventHandler(
            'custom_' + manifest.id + '_config_change',
            (message: any) => pluginConfigChange(message)(store.dispatch),
        );

        // Fetch the current config whenever we recover an internet connection.
        registry.registerReconnectHandler(() => fetchPluginConfig(store.dispatch, store.getState));
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
