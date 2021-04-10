import {GlobalState} from 'mattermost-redux/types/store';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';

import {connect} from 'react-redux';

import {spoilerMode} from 'selectors';

import SpoilerPostType from './spoiler_post_type';

function mapStateToProps(state: GlobalState) {
    return {
        spoilerMode: spoilerMode(state),
        theme: getTheme(state),
    };
}

export default connect(mapStateToProps)(SpoilerPostType);
