import React from 'react';
import PropTypes from 'prop-types';

const {formatText, messageHtmlToComponent} = window.PostUtils;

export default class SpoilerPostType extends React.PureComponent {
    static propTypes = {
        post: PropTypes.object.isRequired,
        theme: PropTypes.object.isRequired,
        spoilerMode: PropTypes.string.isRequired,
    };

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

        this.renderNormal = (formattedSpoilerCaption, spoilerContent) => {
            const formattedSpoilerContent = messageHtmlToComponent(formatText(spoilerContent));
            return (
                <div>{formattedSpoilerCaption ? <div>{formattedSpoilerCaption}</div> : ''} {formattedSpoilerContent}</div>
            );
        };

        this.renderSpoiler = (formattedSpoilerCaption, spoilerContent) => {
            // don't display real text so emoji, url, image... are not visible
            const yaourt = Array.from(spoilerContent).
                map((c) => ((/\s/).test(c) ? c : '_')).join('');
            const lines = yaourt.split(/\r?\n/).map((line) => messageHtmlToComponent(line));
            const divProps = {
                onClick: this.revealSpoiler,
                style: {background: this.props.theme.centerChannelColor},
                title: 'Reveal spoiler',
            };
            return (
                <div>
                    {formattedSpoilerCaption ? <div>{formattedSpoilerCaption}</div> : ''}
                    {lines.map((line, index) => {
                        return <div key={index}><span {...divProps}>{line}<br/></span></div>;
                    })}
                </div>
            );
        };
    }

    render() {
        // Don't use post.message directly as it has a special formatting used by the native apps
        const post = {...this.props.post};
        const spoilerCaption = messageHtmlToComponent(formatText(post.message || ''));
        const spoilerContent = post.props.CustomSpoilerRawMessage || '';
        if (this.state.displaySpoilerContent) {
            return this.renderNormal(spoilerCaption, spoilerContent);
        }
        return this.renderSpoiler(spoilerCaption, spoilerContent);
    }
}
