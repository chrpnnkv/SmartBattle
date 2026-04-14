import type { IApiService } from './IApiService';
import { mockApiService } from './mock/mockApiService';
import { realApiService } from './real/realApiService';

const USE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false';

export const api: IApiService = USE_MOCK ? mockApiService : realApiService;
