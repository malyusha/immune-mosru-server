"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const logger_1 = __importDefault(require("@app/pkg/logger"));
const axios_1 = __importDefault(require("@app/pkg/axios/axios"));
const axios_retry_1 = __importDefault(require("axios-retry"));
class AuthClient {
    constructor(cfg) {
        const config = {
            baseURL: cfg.url,
            timeout: cfg.timeoutMs,
            headers: { 'User-Agent': 'immune-app' }
        };
        const client = axios_1.default(config);
        axios_retry_1.default(client, {
            retries: 3,
        });
        this.axios = client;
        this.logger = logger_1.default.child({ package: AuthClient.name });
    }
    async getQRDataByCode(code) {
        return await this.axios.get(`/certs/${code}`).then((res) => {
            this.logger.info(`requesting certificate for code ${code}`);
            return res.data;
        }).catch((err) => {
            if (err.response) {
                const { status, statusText } = err.response;
                throw new Error(`response error: '${status} ${statusText}'`);
            }
            else if (err.request) {
                throw new Error(`failed to send request: ${err.message}`);
            }
            else {
                throw err;
            }
        });
    }
}
exports.default = AuthClient;
