import {connect} from 'react-redux';

import {spoilerMode} from 'selectors';

import SpoilerPostType from './spoiler_post_type';

const mapStateToProps = (state) => {
    return {spoilerMode: spoilerMode(state)};
};

export default connect(mapStateToProps)(SpoilerPostType);
