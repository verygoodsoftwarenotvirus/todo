import axios, { AxiosError, AxiosResponse } from 'axios';

import { Logger } from '@/logger';

import { backendRoutes } from '@/constants/routes';
import { statusCodes } from '@/constants';

const logger = new Logger().withDebugValue('source', 'src/apiClient/admin.ts');

export function cycleCookieSecret(): Promise<AxiosResponse> {
  const uri = backendRoutes.CYCLE_COOKIE_SECRET;
  return axios
    .post(uri, {}, { validateStatus: () => true })
    .then((res: AxiosResponse) => {
      if (res.status === statusCodes.UNAUTHORIZED) {
        throw new Error('unauthorized for cycling cookie secrets');
      }
      return res;
    });
}
