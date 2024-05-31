import {GlobalState} from 'mattermost-redux/types/store';

import manifest from './manifest';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getPluginState = (state: any) => {
    return state['plugins-' + manifest.id] || {};
};

export const spoilerMode = (state: GlobalState): string => {
    return getPluginState(state).spoilerMode;
};
