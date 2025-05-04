import React from 'react';
import PropTypes from 'prop-types';
import './spoiler_post_type.scss';

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

        this.revealSpoiler = (e) => {
            e.preventDefault();
            if (!this.state.displaySpoilerContent) {
                this.setState({displaySpoilerContent: true});
            }
        };
    }

    render() {
        const post = {...this.props.post};
        const spoilerCaption = post.message || '';
        const spoilerContent = post.props.CustomSpoilerRawMessage || '';
        const showSpoiler = this.state.displaySpoilerContent;

        const formattedCaption = messageHtmlToComponent(formatText(spoilerCaption));
        const formattedContent = messageHtmlToComponent(formatText(spoilerContent));

        const divProps = {
            onClick: this.revealSpoiler,
        };
        return (
            <div>
                {formattedCaption ? <div>{formattedCaption}</div> : ''}
                <span
                    className={`spoiler-content ${showSpoiler ? 'reveal-spoiler' : ''}`}
                    aria-label='Show spoiler'
                    aria-expanded={showSpoiler}
                    tabIndex='0'
                    role='button'
                    {...divProps}
                >
                    <span
                        aria-hidden={!showSpoiler}
                        className='spoilered-text-content'
                    >
                        {formattedContent}
                    </span>
                </span>
            </div>
        );
    }
}
