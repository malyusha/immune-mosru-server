import logger, {Logger} from '@app/pkg/logger';
import {AxiosInstance, AxiosRequestConfig} from 'axios';
import createClient from "@app/pkg/axios/axios";
import axiosRetry from "axios-retry";
import {HTTPClientConfig} from "@app/types";

/**
 * Represents response from Youla API, containing information of authenticated user.
 */
type QRDataResponse = {
    name: string
    dateBirth: string
    expiringAt: string
}

export default class AuthClient {
    readonly axios: AxiosInstance;
    protected logger: Logger;

    constructor(cfg: HTTPClientConfig) {
        const config: AxiosRequestConfig = {
            baseURL: cfg.url,
            timeout: Number(cfg.timeoutMs),
            headers: {'User-Agent': 'immune-app'}
        };

        const client = createClient(config);
        axiosRetry(client, {
            retries: 3,
        });
        this.axios = client;

        this.logger = logger.child({package: AuthClient.name});
    }

    public async getQRDataByCode(code: string): Promise<QRDataResponse> {
        return await this.axios.get(`/certs/${code}`).then((res) => {
            this.logger.info(`requesting certificate for code ${code}`);
            return res.data as QRDataResponse;
        }).catch((err) => {
            if (err.response) {
                const {status, statusText} = err.response;
                throw new Error(`response error: '${status} ${statusText}'`);
            } else if (err.request) {
                throw new Error(`failed to send request: ${err.message}`);
            } else {
                throw err;
            }
        });
    }
}
