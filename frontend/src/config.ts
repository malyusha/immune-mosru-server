import fs from 'fs';
import {defaultsDeep, set} from 'lodash';
import {Config} from './types';

const defaultConfiguration: Config = {
    http: {
        port: 8080
    },
    logLevel: 'info',
    appName: 'immune',
    clients: {
        certs: {
            url: '',
            timeoutMs: 3000
        },
    },
};

export default function loadConfiguration(filepath?: string): Config {
    let config: Config;

    if (filepath) {
        try {
            config = JSON.parse(fs.readFileSync(filepath, 'utf8')) as Config;
        } catch (e) {
            console.warn(`Failed to parse config file: ${e.message}`);
        }
    }

    const res = defaultsDeep(config, {...defaultConfiguration}) as Config;

    return envSubstitute(res);
}

function envSubstitute(config: Config): Config {
    const envMap = {
        'VERSION': 'version',
        'APP_NAME': 'appName',
        'APP_HTTP_PORT': 'http.port',
        'LOG_LEVEL': 'logLevel',
        'CERTS_CLIENT_URL': 'clients.certs.url',
        'CERTS_CLIENT_TIMEOUT_MS': 'clients.certs.timeoutMs'
    };

    const envVars = Object.keys(envMap);
    envVars.forEach((envVar: string) => {
        if (!(envVar in process.env) || process.env[envVar] === '') {
            return true;
        }

        let envValue = process.env[envVar];

        try {
            envValue = JSON.parse(envValue);
        } catch (e) {
            // do nothing, it's a string
        }

        set(config, envMap[envVar], envValue);
    });

    return config;
}
