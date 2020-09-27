import {id as pluginId} from './manifest';

const getPluginState = (state) => {
    return state['plugins-' + pluginId] || {};
};

export const spoilerMode = (state) => {
    return getPluginState(state).spoilerMode;
};
