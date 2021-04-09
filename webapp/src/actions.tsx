import {Dispatch} from 'redux';

import {getConfig as getServerConfig} from 'mattermost-redux/selectors/entities/general';
import {GetStateFunc} from 'mattermost-redux/types/actions';
import {GlobalState} from 'mattermost-redux/types/store';

import manifest from './manifest';
import {PLUGIN_CONFIG_CHANGE} from './action_types';

export const pluginConfigChange = (message: any) => (dispatch: Dispatch) => dispatch({
    type: PLUGIN_CONFIG_CHANGE,
    data: message.data,
});

export const fetchPluginConfig = async (dispatch: Dispatch, getState: GetStateFunc) => {
    fetch(getPluginServerRoute(getState()) + '/config').then((r) => r.json()).then((r) => pluginConfigChange({data: r})(dispatch));
};

export const getPluginServerRoute = (state: GlobalState) => {
    const config = getServerConfig(state);

    let basePath = '/';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath + '/plugins/' + manifest.id;
};
