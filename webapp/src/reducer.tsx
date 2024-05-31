import {PLUGIN_CONFIG_CHANGE} from './action_types';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const config = (state = '', action: any) => {
    switch (action.type) {
    case PLUGIN_CONFIG_CHANGE:
        return action.data;

    default:
        return state;
    }
};

export default config;
