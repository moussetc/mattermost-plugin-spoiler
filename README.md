# Spoiler Plugin

This plugin creates a slash command to display spoiler messages (blurred on the webapp, tagged native apps).

## Usage
The `/spoiler This is a spoiler` command will make a post which will be appear
-  As *blurred text* on the webapp client, which can be unblurred with a click. 
- Surrounded by a *<SPOILER> tag* on the native apps (Android, Apple).

## Compatibility
- This plugin is only compatible with **Mattermost versions 5.2 and higher.**
- **Android and Apple apps will display spoilers in a degraded mode** (see above) as the plugin system does not yet support customization of the native apps.

## Installation
1. Download the [release package](https://github.com/moussetc/mattermost-plugin-spoiler/releases).
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the package
3. **Activate the plugin** in the `System Console > Plugins Management > Management` page

## Manual configuration
If you need to enable & configure this plugin directly in the Mattermost configuration file `config.json`, for example if you are doing a [High Availability setup](https://docs.mattermost.com/deployment/cluster.html), you can use the following lines:
```json
 "PluginSettings": {
        // [...]
        "PluginStates": {
            // [...]
            "com.github.moussetc.mattermost.plugin.spoiler": {
                "Enable": true
            },
        }
    }
```

## Development
Build the plugin with the following command:
```
make
```
This will produce a single plugin package (with support for multiple architectures) in `dist/com.github.moussetc.mattermost.plugin.spoiler-X.X.X.tar.gz`

To automate deploying and enabling the plugin to your server, add the following lines at the beginning of the Makefile (it requires [http](https://httpie.org/) to be installed) and configure your admin login&password:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
```
and use this command:
```
make deploy
```