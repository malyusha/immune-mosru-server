export interface Runnable {
    /**
     * Begins execution of something execution.
     *
     * @throws Error
     */
    run(): void
}

export interface HTTPConfig {
    port: number
}

/**
 * Represents configuration for federation gateway service.
 */
export interface Config {
    http: HTTPConfig,
    logLevel: string,
    appName: string,
    clients: {
        certs: HTTPClientConfig
    }
}

export type HTTPClientConfig = {
    url: string
    timeoutMs: number
}
