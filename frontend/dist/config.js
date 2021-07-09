"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const fs_1 = __importDefault(require("fs"));
const lodash_1 = require("lodash");
const defaultConfiguration = {
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
function loadConfiguration(filepath) {
    let config;
    if (filepath) {
        try {
            config = JSON.parse(fs_1.default.readFileSync(filepath, 'utf8'));
        }
        catch (e) {
            console.warn(`Failed to parse config file: ${e.message}`);
        }
    }
    const res = lodash_1.defaultsDeep(config, Object.assign({}, defaultConfiguration));
    return envSubstitute(res);
}
exports.default = loadConfiguration;
function envSubstitute(config) {
    const envMap = {
        'VERSION': 'version',
        'APP_NAME': 'appName',
        'APP_HTTP_PORT': 'http.port',
        'LOG_LEVEL': 'logLevel',
        'CERTS_CLIENT_URL': 'clients.certs.url',
        'CERTS_CLIENT_TIMEOUT_MS': 'clients.certs.timeoutMs'
    };
    const envVars = Object.keys(envMap);
    envVars.forEach((envVar) => {
        if (!(envVar in process.env) || process.env[envVar] === '') {
            return true;
        }
        let envValue = process.env[envVar];
        try {
            envValue = JSON.parse(envValue);
        }
        catch (e) {
            // do nothing, it's a string
        }
        lodash_1.set(config, envMap[envVar], envValue);
    });
    return config;
}
