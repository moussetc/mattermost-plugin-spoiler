import React from 'react';
import {Theme} from 'mattermost-redux/types/preferences';
import {Post} from 'mattermost-redux/types/posts';

const {formatText, messageHtmlToComponent} = (window as any).PostUtils;

export type Props = {
    post: Post;
    spoilerMode: string;
    theme: Theme;
}

type State = {
    displaySpoilerContent: boolean;
}

export default class SpoilerPostType extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = {
            displaySpoilerContent: false,
        };
    }

    revealSpoiler = () => {
        if (!this.state.displaySpoilerContent) {
            this.setState({displaySpoilerContent: true});
        }
    }

    renderNormal = (spoiler: string, description: string) => {
        const formattedText = formatText(description + '\n' + spoiler, {atMentions: true});
        return (
            <div>{messageHtmlToComponent(formattedText)}</div>
        );
    }

    renderSpoiler = (message: string, description: string) => {
        const formattedDescription = formatText(description, {atMentions: true});

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
                {messageHtmlToComponent(formattedDescription)}
                {lines.map((line, index) => {
                    return <div key={index}><span {...divProps}>{messageHtmlToComponent(line)}<br/></span></div>;
                })}
            </div>
        );
    }

    render() {
        // Don't use post.message directly as it has a special formatting used by the native apps
        const {post} = this.props;
        const message = post.props.CustomSpoilerRawMessage || '';
        const description = post.props.CustomSpoilerDescription || '';
        if (this.state.displaySpoilerContent) {
            return this.renderNormal(message, description);
        }
        return this.renderSpoiler(message, description);
    }
}
