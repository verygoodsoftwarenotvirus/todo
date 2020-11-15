import axios, { AxiosResponse } from 'axios';

import { Logger } from '@/logger';

import { backendRoutes } from '@/constants/routes';
import {
  defaultAPIRequestConfig,
  requestLogFunction,
} from '@/requests/defaults';

const logger = new Logger().withDebugValue('source', 'src/requests/admin.ts');

export function cycleCookieSecret(): Promise<AxiosResponse> {
  const uri = backendRoutes.CYCLE_COOKIE_SECRET;
  return axios
    .post(uri, defaultAPIRequestConfig)
    .then(requestLogFunction(logger, uri));
}
