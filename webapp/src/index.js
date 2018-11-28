import {id as pluginId} from './manifest';
import SpoilerPostType from './components/spoiler_post_type';
import {
    getConfig,
    websocketConfigChange,
} from './actions';
import {spoilerMode} from './selectors';
import reducer from './reducer';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerReducer(reducer);

        let spoilerPostTypeId;
        let currentSpoilerMode;

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
        store.dispatch(getConfig());

        // Be alerted if the plugin configuration change
        registry.registerWebSocketEventHandler(
            'custom_' + pluginId + '_config_change',
            (message) => store.dispatch(websocketConfigChange(message))
        );

        // Fetch the current config whenever we recover an internet connection.
        registry.registerReconnectHandler(() => store.dispatch(getConfig()));
    }
}

window.registerPlugin(pluginId, new Plugin());
