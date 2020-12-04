import { statusCodes } from '@/constants';
import { backendRoutes } from '@/constants/routes';
import { Logger } from '@/logger';
import axios, { AxiosResponse } from 'axios';

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
