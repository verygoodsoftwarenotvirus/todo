import axios, { AxiosResponse } from "axios";
import format from "string-format";

import type {
    QueryFilter,
    User,
} from "@/types";

import {defaultAPIRequestConfig, requestLogFunction} from "@/requests/defaults";
import { Logger } from "@/logger";
import {backendRoutes} from "@/constants/routes";

const logger = new Logger().withDebugValue("source", "src/requests/users.ts");

export function fetchListOfUsers(qf: QueryFilter, adminMode: boolean = false): Promise<AxiosResponse> {
    const outboundURLParams = qf.toURLSearchParams();

    if (adminMode) {
        outboundURLParams.set("admin", "true");
    }

    const uri = `/api/v1/users?${outboundURLParams.toString()}`;

    return axios.get(uri, { withCredentials: true })
        .then(requestLogFunction(logger, uri));
}

export function fetchUser(userID: number): Promise<AxiosResponse> {
    const uri = format(backendRoutes.INDIVIDUAL_USER, userID.toString())
    return axios.get(uri, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function saveUser(u: User): Promise<AxiosResponse> {
    const uri = format(backendRoutes.INDIVIDUAL_USER, u.id.toString())
    return axios.put(uri, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}

export function deleteUser(id: number): Promise<AxiosResponse> {
    const uri = format(backendRoutes.INDIVIDUAL_USER, id.toString())
    return axios.delete(uri, defaultAPIRequestConfig)
        .then(requestLogFunction(logger, uri));
}
