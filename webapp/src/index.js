import {id as pluginId} from './manifest';
import SpoilerPostType from './spoiler_post_type';

export default class Plugin {
    // eslint-disable-next-line no-unused-vars
    initialize(registry, store) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerPostTypeComponent('custom_spoiler', SpoilerPostType);
    }
}

window.registerPlugin(pluginId, new Plugin());
