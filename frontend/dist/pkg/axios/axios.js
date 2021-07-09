"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const axios_1 = __importDefault(require("axios"));
const http_1 = require("http");
const https_1 = require("https");
const defaultHeaders = {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
};
/**
 * @param {AxiosRequestConfig} cfg Base schemaRegistryClient configuration.
 * @returns {AxiosInstance}
 */
function createClient(cfg) {
    // merge default headers and provided
    cfg.headers = Object.assign(Object.assign({}, defaultHeaders), cfg.headers);
    cfg.httpAgent = new http_1.Agent({ keepAlive: true });
    cfg.httpsAgent = new https_1.Agent({ keepAlive: true });
    const instance = axios_1.default.create(cfg);
    return instance;
}
exports.default = createClient;
