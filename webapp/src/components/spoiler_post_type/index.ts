import {connect} from 'react-redux';

import {GlobalState} from 'mattermost-redux/types/store';

import {spoilerMode} from 'selectors';

import SpoilerPostType from './spoiler_post_type';

const mapStateToProps = (state: GlobalState) => {
    return {spoilerMode: spoilerMode(state)};
};

export default connect(mapStateToProps)(SpoilerPostType);
