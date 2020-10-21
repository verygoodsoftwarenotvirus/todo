import type { AxiosRequestConfig, AxiosResponse } from 'axios';

import type { Logger } from "@/logger";

export const defaultAPIRequestConfig: AxiosRequestConfig = {
    withCredentials: true
}

export function requestLogFunction(logger: Logger, uri: string) {
    return (response: AxiosResponse) => {
        logger.withDebugValue("responseStatus", response.status.toString()).debug(`request made to ${uri}`);
        return response;
    }
}