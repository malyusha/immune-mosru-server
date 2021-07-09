import axios, {AxiosInstance, AxiosRequestConfig} from 'axios';
import {Agent as HTTPAgent} from 'http';
import {Agent as HTTPSAgent} from 'https';

const defaultHeaders = {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
};

/**
 * @param {AxiosRequestConfig} cfg Base schemaRegistryClient configuration.
 * @returns {AxiosInstance}
 */
export default function createClient(cfg: AxiosRequestConfig): AxiosInstance {
    // merge default headers and provided
    cfg.headers = {...defaultHeaders, ...cfg.headers};
    cfg.httpAgent = new HTTPAgent({keepAlive: true});
    cfg.httpsAgent = new HTTPSAgent({keepAlive: true});

    const instance = axios.create(cfg);

    return instance;
}
