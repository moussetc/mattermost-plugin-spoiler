import React from 'react';
import PropTypes from 'prop-types';

const PostUtils = window.PostUtils; // must be accessed through `window`

export default class SpoilerPostType extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
        spoilerMode: PropTypes.string.isRequired,
    }

    constructor(props) {
        super(props);
        this.state = {
            displaySpoilerContent: false,
        };

        this.revealSpoiler = () => {
            if (!this.state.displaySpoilerContent) {
                this.setState({displaySpoilerContent: true});
            }
        };

        this.renderNormal = (spoiler, description) => {
            const formattedText = PostUtils.formatText(description + '\n' + spoiler, {atMentions: true});
            return (
                <div>{PostUtils.messageHtmlToComponent(formattedText)}</div>
            );
        };

        this.renderSpoiler = (message, description) => {
            const formattedDescription = PostUtils.formatText(description, {atMentions: true});

            // don't display real text so emoji, url, image... are not visible
            const yaourt = Array.from(message).
                map((c) => ((/\s/).test(c) ? c : '_')).join('');
            const lines = yaourt.split(/\r?\n/);
            const divProps = {
                onClick: this.revealSpoiler,
                style: {background: this.props.theme.centerChannelColor},
                title: 'Reveal spoiler',
            };
            return (
                <div>
                    {PostUtils.messageHtmlToComponent(formattedDescription)}
                    {lines.map((line, index) => {
                        return <div key={index}><span {...divProps}>{PostUtils.messageHtmlToComponent(line)}<br/></span></div>;
                    })}
                </div>
            );
        };
    }

    render() {
        // Don't use post.message directly as it has a special formatting used by the native apps
        const post = {...this.props.post};
        const message = post.props.CustomSpoilerRawMessage || '';
        const description = post.props.CustomSpoilerDescription || '';
        if (this.state.displaySpoilerContent) {
            return this.renderNormal(message, description);
        }
        return this.renderSpoiler(message, description);
    }
}
