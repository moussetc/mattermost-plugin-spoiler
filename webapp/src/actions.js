import {getConfig as getServerConfig} from 'mattermost-redux/selectors/entities/general';

import {id as pluginId} from './manifest';
import {CONFIG_CHANGE} from './action_types';

export const getConfig = () => async (dispatch, getState) => {
    fetch(getPluginServerRoute(getState()) + '/config').then((r) => r.json()).then((r) => {
        dispatch({
            type: CONFIG_CHANGE,
            data: r,
        });
    });
};

export const getPluginServerRoute = (state) => {
    const config = getServerConfig(state);

    let basePath = '/';
    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath + '/plugins/' + pluginId;
};

export const websocketConfigChange = (message) => (dispatch) => dispatch({
    type: CONFIG_CHANGE,
    data: message.data,
});
