import axios, { type AxiosRequestConfig, type AxiosRequestHeaders, type AxiosResponse } from 'axios';

import type { ApiResponse } from '@/core/api.ts';

import type {
    ApplicationCloudSetting
} from '@/core/setting.ts';
import type {
    VersionInfo
} from '@/core/version.ts';
import {
    TransactionType
} from '@/core/transaction.ts';

import {
    BASE_API_URL_PATH,
    BASE_QRCODE_PATH,
    BASE_PROXY_URL_PATH,
    BASE_AMAP_API_PROXY_URL_PATH,
    DEFAULT_API_TIMEOUT,
    DEFAULT_UPLOAD_API_TIMEOUT,
    DEFAULT_EXPORT_API_TIMEOUT,
    DEFAULT_IMPORT_API_TIMEOUT,
    DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT,
    DEFAULT_LLM_API_TIMEOUT,
    GOOGLE_MAP_JAVASCRIPT_URL,
    BAIDU_MAP_JAVASCRIPT_URL,
    AMAP_JAVASCRIPT_URL
} from '@/consts/api.ts';

import type {
    AccountCreateRequest,
    AccountModifyRequest,
    AccountInfoResponse,
    AccountHideRequest,
    AccountMoveRequest,
    AccountDeleteRequest
} from '@/models/account.ts';
import type {
    AuthResponse,
    RegisterResponse
} from '@/models/auth_response.ts';
import type {
    ExportTransactionDataRequest,
    ClearDataRequest,
    ClearAccountTransactionsRequest,
    DataStatisticsResponse
} from '@/models/data_management.ts';
import type {
    UserCustomExchangeRateUpdateRequest,
    UserCustomExchangeRateDeleteRequest,
    UserCustomExchangeRateUpdateResponse,
    LatestExchangeRateResponse
} from '@/models/exchange_rate.ts';
import type {
    ForgetPasswordRequest
} from '@/models/forget_password.ts';
import type {
    ImportTransactionResponsePageWrapper
} from '@/models/imported_transaction.ts';
import type {
    TransactionCreateRequest,
    TransactionModifyRequest,
    TransactionMoveBetweenAccountsRequest,
    TransactionDeleteRequest,
    TransactionImportRequest,
    TransactionListByMaxTimeRequest,
    TransactionListInMonthByPageRequest,
    TransactionInfoResponse,
    TransactionInfoPageWrapperResponse,
    TransactionInfoPageWrapperResponse2,
    TransactionReconciliationStatementRequest,
    TransactionReconciliationStatementResponse,
    TransactionStatisticRequest,
    TransactionStatisticResponse,
    TransactionStatisticTrendsRequest,
    TransactionStatisticTrendsResponseItem,
    TransactionAmountsRequestParams,
    TransactionAmountsResponse
} from '@/models/transaction.ts';
import {
    TransactionAmountsRequest
} from '@/models/transaction.ts';
import type {
    TransactionCategoryCreateRequest,
    TransactionCategoryCreateBatchRequest,
    TransactionCategoryModifyRequest,
    TransactionCategoryHideRequest,
    TransactionCategoryMoveRequest,
    TransactionCategoryDeleteRequest,
    TransactionCategoryInfoResponse
} from '@/models/transaction_category.ts';
import type {
    TransactionPictureUnusedDeleteRequest,
    TransactionPictureInfoBasicResponse
} from '@/models/transaction_picture_info.ts';
import type {
    TransactionTagCreateRequest,
    TransactionTagCreateBatchRequest,
    TransactionTagModifyRequest,
    TransactionTagHideRequest,
    TransactionTagMoveRequest,
    TransactionTagDeleteRequest,
    TransactionTagInfoResponse
} from '@/models/transaction_tag.ts';
import type {
    TransactionTemplateCreateRequest,
    TransactionTemplateModifyRequest,
    TransactionTemplateHideRequest,
    TransactionTemplateMoveRequest,
    TransactionTemplateDeleteRequest,
    TransactionTemplateInfoResponse
} from '@/models/transaction_template.ts';
import type {
    TokenGenerateMCPRequest,
    TokenRevokeRequest,
    TokenGenerateMCPResponse,
    TokenRefreshResponse,
    TokenInfoResponse
} from '@/models/token.ts';
import type {
    TwoFactorEnableConfirmRequest,
    TwoFactorEnableResponse,
    TwoFactorEnableConfirmResponse,
    TwoFactorDisableRequest,
    TwoFactorRegenerateRecoveryCodeRequest,
    TwoFactorStatusResponse
} from '@/models/two_factor.ts';
import type {
    UserLoginRequest,
    UserRegisterRequest,
    UserVerifyEmailResponse,
    UserResendVerifyEmailRequest,
    UserProfileResponse,
    UserProfileUpdateRequest,
    UserProfileUpdateResponse
} from '@/models/user.ts';
import type {
    UserExternalAuthUnlinkRequest,
    UserExternalAuthInfoResponse
} from '@/models/user_external_auth.ts';
import type {
    OAuth2CallbackLoginRequest
} from '@/models/oauth2.ts';
import type {
    UserApplicationCloudSettingsUpdateRequest
} from '@/models/user_app_cloud_setting.ts';
import type {
    RecognizedReceiptImageResponse
} from '@/models/large_language_model.ts';
import type {
    FundInfoResponse,
    FundMemberResponse,
    FundCreateRequest,
    FundModifyRequest,
    FundDeleteRequest,
    FundMemberCreateRequest,
    FundMemberDeleteRequest,
    FundMemberLinkRequest
} from '@/models/fund.ts';

import {
    getCurrentToken,
    clearCurrentTokenAndUserInfo
} from './userstate.ts';

import {
    isDefined,
    isBoolean
} from './common.ts';
import {
    getGoogleMapAPIKey,
    getBaiduMapAK,
    getAmapApplicationKey,
    getExchangeRatesRequestTimeout
} from './server_settings.ts';
import { getTimezoneOffsetMinutes } from './datetime.ts';
import { generateRandomUUID } from './misc.ts';
import { getBasePath } from './web.ts';
import logger from './logger.ts';

interface ApiRequestConfig extends AxiosRequestConfig {
    readonly headers: AxiosRequestHeaders;
    readonly noAuth?: boolean;
    readonly ignoreBlocked?: boolean;
    readonly ignoreError?: boolean;
    readonly timeout?: number;
    readonly cancelableUuid?: string;
}

export type ApiResponsePromise<T> = Promise<AxiosResponse<ApiResponse<T>>>;

let needBlockRequest = false;
const blockedRequests: ((token: string | undefined) => void)[] = [];
const cancelableRequests: Record<string, boolean> = {};

axios.defaults.baseURL = getBasePath() + BASE_API_URL_PATH;
axios.defaults.timeout = DEFAULT_API_TIMEOUT;
axios.interceptors.request.use((config: ApiRequestConfig) => {
    const token = getCurrentToken();

    if (token && !config.noAuth) {
        config.headers.Authorization = `Bearer ${token}`;
    }

    config.headers['X-Timezone-Offset'] = getTimezoneOffsetMinutes();

    if (needBlockRequest && !config.ignoreBlocked) {
        return new Promise(resolve => {
            blockedRequests.push(newToken => {
                if (newToken) {
                    config.headers.Authorization = `Bearer ${newToken}`;
                }

                resolve(config);
            });
        });
    }

    return config;
}, error => {
    return Promise.reject(error);
});

axios.interceptors.response.use(response => {
    if ('cancelableUuid' in response.config && response.config.cancelableUuid && cancelableRequests[response.config.cancelableUuid as string]) {
        logger.debug('Response canceled by user request, url: ' + response.config.url + ', cancelableUuid: ' + response.config.cancelableUuid);
        delete cancelableRequests[response.config.cancelableUuid as string];
        return Promise.reject({ canceled: true });
    }

    return response;
}, error => {
    if ('cancelableUuid' in error.response.config && error.response.config.cancelableUuid && cancelableRequests[error.response.config.cancelableUuid]) {
        logger.debug('Response canceled by user request, url: ' + error.response.config.url + ', cancelableUuid: ' + error.response.config.cancelableUuid);
        delete cancelableRequests[error.response.config.cancelableUuid];
        return Promise.reject({ canceled: true });
    }

    if (error.response && !error.response.config.ignoreError && error.response.data && error.response.data.errorCode) {
        const errorCode = error.response.data.errorCode;

        if (errorCode === 202001 // unauthorized access
            || errorCode === 202002 // current token is invalid
            || errorCode === 202003 // current token is expired
            || errorCode === 202004 // current token type is invalid
            || errorCode === 202005 // current token requires two-factor authorization
            || errorCode === 202006 // current token does not require two-factor authorization
            || errorCode === 202012 // token is empty
        ) {
            clearCurrentTokenAndUserInfo(false);
            location.reload();
            return Promise.reject({ processed: true });
        }
    }

    return Promise.reject(error);
});

export default {
    setLocale: (locale: string) => {
        axios.defaults.headers.common['Accept-Language'] = locale;
    },
    authorize: (data: UserLoginRequest): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('authorize.json', data);
    },
    authorize2FA: ({ passcode, token }: { passcode: string, token: string }): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('2fa/authorize.json', {
            passcode: passcode
        }, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        });
    },
    authorize2FAByBackupCode: ({ recoveryCode, token }: { recoveryCode: string, token: string }): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('2fa/recovery.json', {
            recoveryCode: recoveryCode
        }, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        });
    },
    authorizeOAuth2: ({ req, token }: { req: OAuth2CallbackLoginRequest, token: string }): ApiResponsePromise<AuthResponse> => {
        return axios.post<ApiResponse<AuthResponse>>('oauth2/authorize.json', req, {
            headers: {
                Authorization: `Bearer ${token}`
            }
        });
    },
    register: (req: UserRegisterRequest): ApiResponsePromise<RegisterResponse> => {
        return axios.post<ApiResponse<RegisterResponse>>('register.json', req);
    },
    verifyEmail: ({ token, requestNewToken }: { token: string, requestNewToken: boolean }): ApiResponsePromise<UserVerifyEmailResponse> => {
        return axios.post<ApiResponse<UserVerifyEmailResponse>>('verify_email/by_token.json?token=' + token, {
            requestNewToken: requestNewToken
        }, {
            noAuth: true,
            ignoreError: true
        } as ApiRequestConfig);
    },
    resendVerifyEmailByUnloginUser: (req: UserResendVerifyEmailRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('verify_email/resend.json', req);
    },
    requestResetPassword: (req: ForgetPasswordRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('forget_password/request.json', req);
    },
    resetPassword: ({ email, token, password }: { email: string, token: string, password: string }): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('forget_password/reset/by_token.json?token=' + token, {
            email: email,
            password: password
        }, {
            noAuth: true,
            ignoreError: true
        } as ApiRequestConfig);
    },
    logout: (): ApiResponsePromise<boolean> => {
        return axios.get<ApiResponse<boolean>>('logout.json');
    },
    refreshToken: (): ApiResponsePromise<TokenRefreshResponse> => {
        return new Promise((resolve) => {
            needBlockRequest = true;

            axios.post<ApiResponse<TokenRefreshResponse>>('v1/tokens/refresh.json', {}, {
                ignoreBlocked: true
            } as ApiRequestConfig).then(response => {
                const data = response.data;

                resolve(response);
                needBlockRequest = false;

                return data.result.newToken;
            }).then(newToken => {
                blockedRequests.forEach(func => func(newToken));
                blockedRequests.length = 0;
            });
        });
    },
    getExternalAuths: (): ApiResponsePromise<UserExternalAuthInfoResponse[]> => {
        return axios.get<ApiResponse<UserExternalAuthInfoResponse[]>>('v1/users/external_auth/list.json');
    },
    unlinkExternalAuth: (req: UserExternalAuthUnlinkRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/external_auth/unlink.json', req);
    },
    getTokens: (): ApiResponsePromise<TokenInfoResponse[]> => {
        return axios.get<ApiResponse<TokenInfoResponse[]>>('v1/tokens/list.json');
    },
    generateMCPToken: (req: TokenGenerateMCPRequest): ApiResponsePromise<TokenGenerateMCPResponse> => {
        return axios.post<ApiResponse<TokenGenerateMCPResponse>>('v1/tokens/generate/mcp.json', req);
    },
    revokeToken: ({ tokenId, ignoreError }: { tokenId: string, ignoreError?: boolean }): ApiResponsePromise<boolean> => {
        const req: TokenRevokeRequest = {
            tokenId: tokenId
        };

        return axios.post<ApiResponse<boolean>>('v1/tokens/revoke.json', req, {
            ignoreError: !!ignoreError
        } as ApiRequestConfig);
    },
    revokeAllTokens: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/tokens/revoke_all.json');
    },
    getProfile: (): ApiResponsePromise<UserProfileResponse> => {
        return axios.get<ApiResponse<UserProfileResponse>>('v1/users/profile/get.json');
    },
    updateProfile: (req: UserProfileUpdateRequest): ApiResponsePromise<UserProfileUpdateResponse> => {
        return axios.post<ApiResponse<UserProfileUpdateResponse>>('v1/users/profile/update.json', req);
    },
    updateAvatar: ({ avatarFile }: { avatarFile: File }): ApiResponsePromise<UserProfileResponse> => {
        return axios.postForm<ApiResponse<UserProfileResponse>>('v1/users/avatar/update.json', {
            avatar: avatarFile
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        });
    },
    removeAvatar: (): ApiResponsePromise<UserProfileResponse> => {
        return axios.post<ApiResponse<UserProfileResponse>>('v1/users/avatar/remove.json');
    },
    resendVerifyEmailByLoginedUser: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/verify_email/resend.json');
    },
    getUserApplicationCloudSettings: (): ApiResponsePromise<ApplicationCloudSetting[] | false> => {
        return axios.get<ApiResponse<ApplicationCloudSetting[] | false>>('v1/users/settings/cloud/get.json');
    },
    updateUserApplicationCloudSettings: (req: UserApplicationCloudSettingsUpdateRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/settings/cloud/update.json', req);
    },
    disableUserApplicationCloudSettings: (): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/settings/cloud/disable.json');
    },
    get2FAStatus: (): ApiResponsePromise<TwoFactorStatusResponse> => {
        return axios.get<ApiResponse<TwoFactorStatusResponse>>('v1/users/2fa/status.json');
    },
    enable2FA: (): ApiResponsePromise<TwoFactorEnableResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableResponse>>('v1/users/2fa/enable/request.json');
    },
    confirmEnable2FA: (req: TwoFactorEnableConfirmRequest): ApiResponsePromise<TwoFactorEnableConfirmResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableConfirmResponse>>('v1/users/2fa/enable/confirm.json', req);
    },
    disable2FA: (req: TwoFactorDisableRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/users/2fa/disable.json', req);
    },
    regenerate2FARecoveryCode: (req: TwoFactorRegenerateRecoveryCodeRequest): ApiResponsePromise<TwoFactorEnableConfirmResponse> => {
        return axios.post<ApiResponse<TwoFactorEnableConfirmResponse>>('v1/users/2fa/recovery/regenerate.json', req);
    },
    getUserDataStatistics: (): ApiResponsePromise<DataStatisticsResponse> => {
        return axios.get<ApiResponse<DataStatisticsResponse>>('v1/data/statistics.json');
    },
    getExportedUserData: (fileType: string, req?: ExportTransactionDataRequest): Promise<AxiosResponse<BlobPart>> => {
        let params = '';

        if (req) {
            const amountFilter = encodeURIComponent(req.amountFilter);
            const keyword = encodeURIComponent(req.keyword);
            params = `max_time=${req.maxTime}&min_time=${req.minTime}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_ids=${req.tagIds}&tag_filter_type=${req.tagFilterType}&amount_filter=${amountFilter}&keyword=${keyword}`;
        } else {
            params = 'max_time=0&min_time=0&type=0&category_ids=&account_ids=&tag_ids=&tag_filter_type=0&amount_filter=&keyword=';
        }

        if (fileType === 'csv') {
            return axios.get<BlobPart>('v1/data/export.csv?' + params, {
                timeout: DEFAULT_EXPORT_API_TIMEOUT
            } as ApiRequestConfig);
        } else if (fileType === 'tsv') {
            return axios.get<BlobPart>('v1/data/export.tsv?' + params, {
                timeout: DEFAULT_EXPORT_API_TIMEOUT
            } as ApiRequestConfig);
        } else {
            return Promise.reject('Parameter Invalid');
        }
    },
    clearAllData: (req: ClearDataRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/all.json', req, {
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    clearAllTransactions: (req: ClearDataRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/transactions.json', req, {
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    clearAllTransactionsOfAccount: (req: ClearAccountTransactionsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/data/clear/transactions/by_account.json', req, {
            timeout: DEFAULT_CLEAR_ALL_TRANSACTIONS_API_TIMEOUT
        } as ApiRequestConfig);
    },
    getAllAccounts: ({ visibleOnly, fundId }: { visibleOnly: boolean, fundId?: string }): ApiResponsePromise<AccountInfoResponse[]> => {
        const params = new URLSearchParams();
        params.append('visible_only', visibleOnly.toString());

        return axios.get<ApiResponse<AccountInfoResponse[]>>(`v1/funds/${fundId}/accounts/list.json?${params.toString()}`);
    },
    getAccount: ({ id, fundId }: { id: string, fundId?: string }): ApiResponsePromise<AccountInfoResponse> => {
        const params = new URLSearchParams();
        params.append('id', id);
        return axios.get<ApiResponse<AccountInfoResponse>>(`v1/funds/${fundId}/accounts/get.json?${params.toString()}`);
    },
    addAccount: (req: AccountCreateRequest, fundId?: string): ApiResponsePromise<AccountInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/accounts/add.json`
        return axios.post<ApiResponse<AccountInfoResponse>>(baseUrl, req);
    },
    modifyAccount: (req: AccountModifyRequest, fundId?: string): ApiResponsePromise<AccountInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/accounts/modify.json`
        return axios.post<ApiResponse<AccountInfoResponse>>(baseUrl, req);
    },
    hideAccount: (req: AccountHideRequest, fundId?: string): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${fundId}/accounts/hide.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    moveAccount: (req: AccountMoveRequest, fundId?: string): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${fundId}/accounts/move.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    deleteAccount: (req: AccountDeleteRequest, fundId?: string): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${fundId}/accounts/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    deleteSubAccount: (req: AccountDeleteRequest, fundId?: string): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${fundId}/accounts/sub_account/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    // Fund services
    getAllFunds: (): ApiResponsePromise<FundInfoResponse[]> => {
        return axios.get<ApiResponse<FundInfoResponse[]>>('v1/funds/list.json');
    },
    getFund: ({ id }: { id: string }): ApiResponsePromise<FundInfoResponse> => {
        return axios.get<ApiResponse<FundInfoResponse>>('v1/funds/get.json?id=' + id);
    },
    addFund: (req: FundCreateRequest): ApiResponsePromise<FundInfoResponse> => {
        return axios.post<ApiResponse<FundInfoResponse>>('v1/funds/add.json', req);
    },
    modifyFund: (req: FundModifyRequest): ApiResponsePromise<FundInfoResponse> => {
        return axios.post<ApiResponse<FundInfoResponse>>('v1/funds/modify.json', req);
    },
    deleteFund: (req: FundDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/funds/delete.json', req);
    },
    getFundMembers: ({ fundId }: { fundId: string }): ApiResponsePromise<FundMemberResponse[]> => {
        return axios.get<ApiResponse<FundMemberResponse[]>>('v1/funds/' + fundId + '/members/list.json');
    },
    addFundMember: (req: FundMemberCreateRequest): ApiResponsePromise<FundMemberResponse> => {
        return axios.post<ApiResponse<FundMemberResponse>>('v1/funds/' + req.fundId + '/members/add.json', req);
    },
    removeFundMember: (req: FundMemberDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/funds/' + req.fundId + '/members/' + req.memberId + '/delete.json', req);
    },
    linkFundMember: (req: FundMemberLinkRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/funds/members/' + req.memberId + '/link.json', req);
    },
    getTransactions: (req: TransactionListByMaxTimeRequest): ApiResponsePromise<TransactionInfoPageWrapperResponse> => {
        const amountFilter = encodeURIComponent(req.amountFilter);
        const keyword = encodeURIComponent(req.keyword);
        const baseUrl = `v1/funds/${req.fundId}/transactions/list.json`
        return axios.get<ApiResponse<TransactionInfoPageWrapperResponse>>(`${baseUrl}?max_time=${req.maxTime}&min_time=${req.minTime}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_ids=${req.tagIds}&tag_filter_type=${req.tagFilterType}&amount_filter=${amountFilter}&keyword=${keyword}&count=${req.count}&page=${req.page}&with_count=${req.withCount}&trim_account=true&trim_category=true&trim_tag=true`);
    },
    getAllTransactionsByMonth: (req: TransactionListInMonthByPageRequest): ApiResponsePromise<TransactionInfoPageWrapperResponse2> => {
        const amountFilter = encodeURIComponent(req.amountFilter);
        const keyword = encodeURIComponent(req.keyword);
        const baseUrl = `v1/funds/${req.fundId}/transactions/list/by_month.json`
        return axios.get<ApiResponse<TransactionInfoPageWrapperResponse2>>(`${baseUrl}?year=${req.year}&month=${req.month}&type=${req.type}&category_ids=${req.categoryIds}&account_ids=${req.accountIds}&tag_ids=${req.tagIds}&tag_filter_type=${req.tagFilterType}&amount_filter=${amountFilter}&keyword=${keyword}&trim_account=true&trim_category=true&trim_tag=true`);
    },
    getReconciliationStatements: (req: TransactionReconciliationStatementRequest): ApiResponsePromise<TransactionReconciliationStatementResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transactions/reconciliation_statements.json`
        return axios.get<ApiResponse<TransactionReconciliationStatementResponse>>(`${baseUrl}?account_id=${req.accountId}&start_time=${req.startTime}&end_time=${req.endTime}`);
    },
    getTransactionStatistics: (req: TransactionStatisticRequest): ApiResponsePromise<TransactionStatisticResponse> => {
        const queryParams = [];

        if (req.startTime) {
            queryParams.push(`start_time=${req.startTime}`);
        }

        if (req.endTime) {
            queryParams.push(`end_time=${req.endTime}`);
        }

        if (req.tagIds) {
            queryParams.push(`tag_ids=${req.tagIds}`);
        }

        if (req.tagFilterType) {
            queryParams.push(`tag_filter_type=${req.tagFilterType}`);
        }

        if (req.keyword) {
            queryParams.push(`keyword=${encodeURIComponent(req.keyword)}`);
        }

        const baseUrl = `v1/funds/${req.fundId}/transactions/statistics.json`
        return axios.get<ApiResponse<TransactionStatisticResponse>>(`${baseUrl}?use_transaction_timezone=${req.useTransactionTimezone}` + (queryParams.length ? '&' + queryParams.join('&') : ''));
    },
    getTransactionStatisticsTrends: (req: TransactionStatisticTrendsRequest): ApiResponsePromise<TransactionStatisticTrendsResponseItem[]> => {
        const queryParams = [];

        if (req.startYearMonth) {
            queryParams.push(`start_year_month=${req.startYearMonth}`);
        }

        if (req.endYearMonth) {
            queryParams.push(`end_year_month=${req.endYearMonth}`);
        }

        if (req.tagIds) {
            queryParams.push(`tag_ids=${req.tagIds}`);
        }

        if (req.tagFilterType) {
            queryParams.push(`tag_filter_type=${req.tagFilterType}`);
        }

        if (req.keyword) {
            queryParams.push(`keyword=${encodeURIComponent(req.keyword)}`);
        }

        const baseUrl = `v1/funds/${req.fundId}/transactions/statistics/trends.json`
        return axios.get<ApiResponse<TransactionStatisticTrendsResponseItem[]>>(`${baseUrl}?use_transaction_timezone=${req.useTransactionTimezone}` + (queryParams.length ? '&' + queryParams.join('&') : ''));
    },
    getTransactionAmounts: (params: TransactionAmountsRequestParams, excludeAccountIds: string[], excludeCategoryIds: string[], fundId?: string): ApiResponsePromise<TransactionAmountsResponse> => {
        const req = TransactionAmountsRequest.of(params);
        let queryParams = req.buildQuery();

        if (excludeAccountIds && excludeAccountIds.length) {
            queryParams = queryParams + `&exclude_account_ids=${excludeAccountIds.join(',')}`;
        }

        if (excludeCategoryIds && excludeCategoryIds.length) {
            queryParams = queryParams + `&exclude_category_ids=${excludeCategoryIds.join(',')}`;
        }

        const baseUrl = `v1/funds/${fundId}/transactions/amounts.json`
        return axios.get<ApiResponse<TransactionAmountsResponse>>(`${baseUrl}?${queryParams}`);
    },
    getTransaction: ({ id, withPictures, fundId }: { id: string, withPictures: boolean | undefined, fundId?: string }): ApiResponsePromise<TransactionInfoResponse> => {
        if (!isDefined(withPictures)) {
            withPictures = true;
        }

        const params = new URLSearchParams();
        params.append('id', id);
        params.append('with_pictures', withPictures.toString());
        params.append('trim_account', 'true');
        params.append('trim_category', 'true');
        params.append('trim_tag', 'true');

        const baseUrl = `v1/funds/${fundId}/transactions/get.json`
        return axios.get<ApiResponse<TransactionInfoResponse>>(`${baseUrl}?${params.toString()}`);
    },
    addTransaction: (req: TransactionCreateRequest, fundId?: string): ApiResponsePromise<TransactionInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/transactions/add.json`
        return axios.post<ApiResponse<TransactionInfoResponse>>(baseUrl, req);
    },
    modifyTransaction: (req: TransactionModifyRequest, fundId?: string): ApiResponsePromise<TransactionInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/transactions/modify.json`
        return axios.post<ApiResponse<TransactionInfoResponse>>(baseUrl, req);
    },
    moveAllTransactionsBetweenAccounts: (req: TransactionMoveBetweenAccountsRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transactions/move/all.json', req);
    },
    deleteTransaction: (req: TransactionDeleteRequest, fundId?: string): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${fundId}/transactions/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    parseImportDsvFile: ({ fileType, fileEncoding, importFile }: { fileType: string, fileEncoding?: string, importFile: File }): ApiResponsePromise<string[][]> => {
        return axios.postForm<ApiResponse<string[][]>>('v1/transactions/parse_dsv_file.json', {
            fileType: fileType,
            fileEncoding: fileEncoding,
            file: importFile
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    parseImportTransaction: ({ fileType, fileEncoding, importFile, columnMapping, transactionTypeMapping, hasHeaderLine, timeFormat, timezoneFormat, amountDecimalSeparator, amountDigitGroupingSymbol, geoSeparator, geoOrder, tagSeparator }: { fileType: string, fileEncoding?: string, importFile: File, columnMapping?: Record<number, number>, transactionTypeMapping?: Record<string, TransactionType>, hasHeaderLine?: boolean, timeFormat?: string, timezoneFormat?: string, amountDecimalSeparator?: string, amountDigitGroupingSymbol?: string, geoSeparator?: string, geoOrder?: string, tagSeparator?: string }): ApiResponsePromise<ImportTransactionResponsePageWrapper> => {
        let textualColumnMapping: string | undefined = undefined;
        let textualTransactionTypeMapping: string | undefined = undefined;
        let textualHasHeaderLine: string | undefined = undefined;

        if (columnMapping) {
            textualColumnMapping = JSON.stringify(columnMapping);
        }

        if (transactionTypeMapping) {
            textualTransactionTypeMapping = JSON.stringify(transactionTypeMapping);
        }

        if (hasHeaderLine) {
            textualHasHeaderLine = 'true';
        }

        return axios.postForm<ApiResponse<ImportTransactionResponsePageWrapper>>('v1/transactions/parse_import.json', {
            fileType: fileType,
            fileEncoding: fileEncoding,
            file: importFile,
            columnMapping: textualColumnMapping,
            transactionTypeMapping: textualTransactionTypeMapping,
            hasHeaderLine: textualHasHeaderLine,
            timeFormat: timeFormat,
            timezoneFormat: timezoneFormat,
            amountDecimalSeparator: amountDecimalSeparator,
            amountDigitGroupingSymbol: amountDigitGroupingSymbol,
            geoSeparator: geoSeparator,
            geoOrder: geoOrder,
            tagSeparator: tagSeparator
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    importTransactions: (req: TransactionImportRequest): ApiResponsePromise<number> => {
        return axios.post<ApiResponse<number>>('v1/transactions/import.json', req, {
            timeout: DEFAULT_IMPORT_API_TIMEOUT
        } as ApiRequestConfig);
    },
    getImportTransactionsProcess: (clientSessionId: string): ApiResponsePromise<number | null> => {
        return axios.get<ApiResponse<number | null>>('v1/transactions/import/process.json?client_session_id=' + clientSessionId, {
            ignoreError: true
        } as ApiRequestConfig);
    },
    uploadTransactionPicture: ({ pictureFile, clientSessionId }: { pictureFile: File, clientSessionId?: string }): ApiResponsePromise<TransactionPictureInfoBasicResponse> => {
        return axios.postForm<ApiResponse<TransactionPictureInfoBasicResponse>>('v1/transaction/pictures/upload.json', {
            picture: pictureFile,
            clientSessionId: clientSessionId
        }, {
            timeout: DEFAULT_UPLOAD_API_TIMEOUT
        } as ApiRequestConfig);
    },
    removeUnusedTransactionPicture: (req: TransactionPictureUnusedDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/transaction/pictures/remove_unused.json', req);
    },
    getAllTransactionCategories: ({ fundId }: { fundId?: string } = {}): ApiResponsePromise<Record<number, TransactionCategoryInfoResponse[]>> => {
        if (fundId) {
            return axios.get<ApiResponse<Record<number, TransactionCategoryInfoResponse[]>>>(`v1/funds/${fundId}/transaction/categories/list.json`);
        } else {
            return axios.get<ApiResponse<Record<number, TransactionCategoryInfoResponse[]>>>('v1/transaction/categories/list.json');
        }
    },
    getTransactionCategory: ({ id, fundId }: { id: string, fundId?: string }): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/transaction/categories/get.json`
        return axios.get<ApiResponse<TransactionCategoryInfoResponse>>(`${baseUrl}?id=${id}`);
    },
    addTransactionCategory: (req: TransactionCategoryCreateRequest & { fundId?: string }): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/add.json`
        return axios.post<ApiResponse<TransactionCategoryInfoResponse>>(baseUrl, req);
    },
    addTransactionCategoryBatch: (req: TransactionCategoryCreateBatchRequest & { fundId?: string }): ApiResponsePromise<Record<number, TransactionCategoryInfoResponse[]>> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/add_batch.json`
        return axios.post<ApiResponse<Record<number, TransactionCategoryInfoResponse[]>>>(baseUrl, req);
    },
    modifyTransactionCategory: (req: TransactionCategoryModifyRequest & { fundId?: string }): ApiResponsePromise<TransactionCategoryInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/modify.json`
        return axios.post<ApiResponse<TransactionCategoryInfoResponse>>(baseUrl, req);
    },
    hideTransactionCategory: (req: TransactionCategoryHideRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/hide.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    moveTransactionCategory: (req: TransactionCategoryMoveRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/move.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    deleteTransactionCategory: (req: TransactionCategoryDeleteRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/categories/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    getAllTransactionTags: ({ fundId }: { fundId?: string } = {}): ApiResponsePromise<TransactionTagInfoResponse[]> => {
        return axios.get<ApiResponse<TransactionTagInfoResponse[]>>(`v1/funds/${fundId}/transaction/tags/list.json`);
    },
    getTransactionTag: ({ id, fundId }: { id: string, fundId?: string }): ApiResponsePromise<TransactionTagInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/transaction/tags/get.json`
        return axios.get<ApiResponse<TransactionTagInfoResponse>>(`${baseUrl}?id=${id}`);
    },
    addTransactionTag: (req: TransactionTagCreateRequest & { fundId?: string }): ApiResponsePromise<TransactionTagInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/add.json`
        return axios.post<ApiResponse<TransactionTagInfoResponse>>(baseUrl, req);
    },
    addTransactionTagBatch: (req: TransactionTagCreateBatchRequest & { fundId?: string }): ApiResponsePromise<TransactionTagInfoResponse[]> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/add_batch.json`
        return axios.post<ApiResponse<TransactionTagInfoResponse[]>>(baseUrl, req);
    },
    modifyTransactionTag: (req: TransactionTagModifyRequest & { fundId?: string }): ApiResponsePromise<TransactionTagInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/modify.json`
        return axios.post<ApiResponse<TransactionTagInfoResponse>>(baseUrl, req);
    },
    hideTransactionTag: (req: TransactionTagHideRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/hide.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    moveTransactionTag: (req: TransactionTagMoveRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/move.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    deleteTransactionTag: (req: TransactionTagDeleteRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/tags/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    getAllTransactionTemplates: ({ templateType, fundId }: { templateType: number, fundId?: string }): ApiResponsePromise<TransactionTemplateInfoResponse[]> => {
        const params = new URLSearchParams();
        params.append('templateType', templateType.toString());
        return axios.get<ApiResponse<TransactionTemplateInfoResponse[]>>(`v1/funds/${fundId}/transaction/templates/list.json?${params.toString()}`);
    },
    getTransactionTemplate: ({ id, fundId }: { id: string, fundId?: string }): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        const baseUrl = `v1/funds/${fundId}/transaction/templates/get.json`
        return axios.get<ApiResponse<TransactionTemplateInfoResponse>>(`${baseUrl}?id=${id}`);
    },
    addTransactionTemplate: (req: TransactionTemplateCreateRequest & { fundId?: string }): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/templates/add.json`
        return axios.post<ApiResponse<TransactionTemplateInfoResponse>>(baseUrl, req);
    },
    modifyTransactionTemplate: (req: TransactionTemplateModifyRequest & { fundId?: string }): ApiResponsePromise<TransactionTemplateInfoResponse> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/templates/modify.json`
        return axios.post<ApiResponse<TransactionTemplateInfoResponse>>(baseUrl, req);
    },
    hideTransactionTemplate: (req: TransactionTemplateHideRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/templates/hide.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    moveTransactionTemplate: (req: TransactionTemplateMoveRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/templates/move.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    deleteTransactionTemplate: (req: TransactionTemplateDeleteRequest & { fundId?: string }): ApiResponsePromise<boolean> => {
        const baseUrl = `v1/funds/${req.fundId}/transaction/templates/delete.json`
        return axios.post<ApiResponse<boolean>>(baseUrl, req);
    },
    recognizeReceiptImage: ({ imageFile, cancelableUuid }: { imageFile: File, cancelableUuid?: string }): ApiResponsePromise<RecognizedReceiptImageResponse> => {
        return axios.postForm<ApiResponse<RecognizedReceiptImageResponse>>('v1/llm/transactions/recognize_receipt_image.json', {
            image: imageFile
        }, {
            timeout: DEFAULT_LLM_API_TIMEOUT,
            cancelableUuid: cancelableUuid
        } as ApiRequestConfig);
    },
    getLatestExchangeRates: (param: { ignoreError?: boolean }): ApiResponsePromise<LatestExchangeRateResponse> => {
        return axios.get<ApiResponse<LatestExchangeRateResponse>>('v1/exchange_rates/latest.json', {
            ignoreError: !!param.ignoreError,
            timeout: getExchangeRatesRequestTimeout() || DEFAULT_API_TIMEOUT
        } as ApiRequestConfig);
    },
    updateUserCustomExchangeRate: (req: UserCustomExchangeRateUpdateRequest): ApiResponsePromise<UserCustomExchangeRateUpdateResponse> => {
        return axios.post<ApiResponse<UserCustomExchangeRateUpdateResponse>>('v1/exchange_rates/user_custom/update.json', req);
    },
    deleteUserCustomExchangeRate: (req: UserCustomExchangeRateDeleteRequest): ApiResponsePromise<boolean> => {
        return axios.post<ApiResponse<boolean>>('v1/exchange_rates/user_custom/delete.json', req);
    },
    getServerVersion: (): ApiResponsePromise<VersionInfo> => {
        return axios.get<ApiResponse<VersionInfo>>('v1/systems/version.json');
    },
    cancelRequest: (cancelableUuid: string) => {
        cancelableRequests[cancelableUuid] = true;
    },
    generateOAuth2LoginUrl: (platform: 'mobile' | 'desktop', clientSessionId: string): string => {
        return `${getBasePath()}/oauth2/login?platform=${platform}&client_session_id=${clientSessionId}`;
    },
    generateQrCodeUrl: (qrCodeName: string): string => {
        return `${getBasePath()}${BASE_QRCODE_PATH}/${qrCodeName}.png`;
    },
    generateMapProxyTileImageUrl: (mapProvider: string, language: string): string => {
        const token = getCurrentToken();
        let url = `${getBasePath()}${BASE_PROXY_URL_PATH}/map/tile/{z}/{x}/{y}.png?provider=${mapProvider}&token=${token}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateMapProxyAnnotationImageUrl: (mapProvider: string, language: string): string => {
        const token = getCurrentToken();
        let url = `${getBasePath()}${BASE_PROXY_URL_PATH}/map/annotation/{z}/{x}/{y}.png?provider=${mapProvider}&token=${token}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateGoogleMapJavascriptUrl: (language: string | undefined, callbackFnName: string): string => {
        let url = `${GOOGLE_MAP_JAVASCRIPT_URL}?key=${getGoogleMapAPIKey()}&libraries=core,marker&callback=${callbackFnName}`;

        if (language) {
            url = url + `&language=${language}`;
        }

        return url;
    },
    generateBaiduMapJavascriptUrl: (callbackFnName: string): string => {
        return `${BAIDU_MAP_JAVASCRIPT_URL}&ak=${getBaiduMapAK()}&callback=${callbackFnName}`;
    },
    generateAmapJavascriptUrl: (callbackFnName: string): string => {
        return `${AMAP_JAVASCRIPT_URL}&key=${getAmapApplicationKey()}&plugin=AMap.ToolBar&callback=${callbackFnName}`;
    },
    generateAmapApiInternalProxyUrl: (): string => {
        return `${window.location.origin}${getBasePath()}${BASE_AMAP_API_PROXY_URL_PATH}`;
    },
    getInternalAvatarUrlWithToken(avatarUrl: string, disableBrowserCache?: boolean | string): string {
        if (!avatarUrl) {
            return avatarUrl;
        }

        const params = [];
        params.push('token=' + getCurrentToken());

        if (disableBrowserCache) {
            if (isBoolean(disableBrowserCache)) {
                params.push('_nocache=' + generateRandomUUID());
            } else {
                params.push('_nocache=' + disableBrowserCache);
            }
        }

        if (avatarUrl.indexOf('?') >= 0) {
            return avatarUrl + '&' + params.join('&');
        } else {
            return avatarUrl + '?' + params.join('&');
        }
    },
    getTransactionPictureUrlWithToken(pictureUrl: string, disableBrowserCache?: boolean | string): string {
        if (!pictureUrl) {
            return pictureUrl;
        }

        const params = [];
        params.push('token=' + getCurrentToken());

        if (disableBrowserCache) {
            if (isBoolean(disableBrowserCache)) {
                params.push('_nocache=' + generateRandomUUID());
            } else {
                params.push('_nocache=' + disableBrowserCache);
            }
        }

        if (pictureUrl.indexOf('?') >= 0) {
            return pictureUrl + '&' + params.join('&');
        } else {
            return pictureUrl + '?' + params.join('&');
        }
    }
};
