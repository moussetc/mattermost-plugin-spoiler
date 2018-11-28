import {CONFIG_CHANGE} from './action_types';

const config = (state = '', action) => {
    switch (action.type) {
    case CONFIG_CHANGE:
        return action.data;

    default:
        return state;
    }
};

export default config;
