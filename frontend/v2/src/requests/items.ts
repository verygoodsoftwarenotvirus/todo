import type { QueryFilter } from "@/models";
import type { AxiosResponse } from "axios";
import axios from "axios";

export function fetchListOfItems(qf: QueryFilter, adminMode: boolean = false): Promise<AxiosResponse> {
    const outboundURLParams = qf.toURLSearchParams();

    if (adminMode) {
        outboundURLParams.set("admin", "true");
    }

    const uri = `/api/v1/items?${outboundURLParams.toString()}`;

    return axios.get(uri, { withCredentials: true })
}