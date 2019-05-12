module.exports = {
    presets: [
        ['@babel/env', {
            targets: {
                chrome: 66,
                firefox: 60,
                edge: 42,
                ie: 11,
                safari: 12,
            },
            useBuiltIns: 'usage',
            shippedProposals: true,
        }],
        '@babel/preset-react',
    ],
    plugins: [
        '@babel/proposal-class-properties',
        '@babel/proposal-object-rest-spread',
    ],
};
