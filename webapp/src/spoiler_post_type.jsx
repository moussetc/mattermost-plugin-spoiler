import React from 'react';
import PropTypes from 'prop-types';

const {formatText, messageHtmlToComponent} = window.PostUtils;

export default class SpoilerPostType extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
    }

    constructor(props) {
	super(props);
	this.state = {
            displaySpoiler: false
	};
	
	this.revealSpoiler = () => {
	    if (!this.state.displaySpoiler) {
                this.setState( { displaySpoiler: true });	
	    }
	}
    }

    render() {
        const style = {};
	if (!this.state.displaySpoiler) {
            style.filter = 'blur(4px)'
	}
        const post = {...this.props.post};
	// Don't use post.message directly as it has a special formatting used by the native apps
        const message = post.props.CustomSpoilerRawMessage || '';
        const formattedText = messageHtmlToComponent(formatText(message));
	const props = {
		onClick: this.revealSpoiler,
		style: style,
		title: this.state.displaySpoiler ? '' : 'Reveal spoiler',
	};
        return (
            <div {...props}>{formattedText}</div>
        );
    }
}
