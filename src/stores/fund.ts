import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

// import { useSettingsStore } from './setting.ts';
// import { useUserStore } from './user.ts';

import {
    Fund,
    FundMember,
    type FundMemberCreateRequest,
    type FundMemberDeleteRequest,
    type FundMemberLinkRequest
} from '@/models/fund.ts';

// import { isObject } from '@/lib/common.ts';
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

export const useFundsStore = defineStore('funds', () => {
    // const settingsStore = useSettingsStore();
    // const userStore = useUserStore();

    const allFunds = ref<Fund[]>([]);
    const allFundsMap = ref<Record<string, Fund>>({});
    const currentFundId = ref<string | null>(null);
    const fundListStateInvalid = ref<boolean>(true);
    const fundListLoading = ref<boolean>(false);
    let fundListLoadingPromise: Promise<Fund[]> | null = null;
    let ensureFundLoadedPromise: Promise<string | null> | null = null;

    const currentFund = computed<Fund | null>(() => {
        if (!currentFundId.value || !allFundsMap.value[currentFundId.value]) {
            return null;
        }
        return allFundsMap.value[currentFundId.value] || null;
    });

    const personalFund = computed<Fund | null>(() => {
        // Find the personal fund (usually the first one created or named "Personal")
        const personalFunds = allFunds.value.filter((fund: Fund) =>
            fund.name === 'Personal' || fund.name.toLowerCase().includes('personal')
        );

        if (personalFunds.length > 0) {
            return personalFunds[0] || null;
        }

        // If no personal fund found, return the first fund
        return allFunds.value.length > 0 ? (allFunds.value[0] || null) : null;
    });

    const currentCurrency = computed<string>(() => {
        const fund = currentFund.value;
        if (fund && fund.defaultCurrency) {
            return fund.defaultCurrency;
        }
        // Fallback to USD if no fund is selected
        return 'USD';
    });

    function loadFundList(funds: Fund[]): void {
        allFunds.value = funds;
        allFundsMap.value = {};

        for (const fund of funds) {
            allFundsMap.value[fund.id] = fund;
        }

        // Set current fund if not set
        if (!currentFundId.value && funds.length > 0) {
            const personal = personalFund.value;
            currentFundId.value = personal ? personal.id : (funds[0]?.id || null);
            if (currentFundId.value) {
                localStorage.setItem('currentFundId', currentFundId.value);
                console.log('Fund store - Set current fund to:', currentFundId.value, 'from', funds.length, 'available funds');
            }
        } else if (currentFundId.value) {
            console.log('Fund store - Current fund already set to:', currentFundId.value);
        } else {
            console.log('Fund store - No funds available, working in legacy mode');
        }
    }

    function addFundToFundList(fund: Fund): void {
        allFunds.value.push(fund);
        allFundsMap.value[fund.id] = fund;
    }

    function updateFundInFundList(fund: Fund): void {
        const index = allFunds.value.findIndex((f: Fund) => f.id === fund.id);
        if (index >= 0) {
            allFunds.value.splice(index, 1, fund);
        }
        allFundsMap.value[fund.id] = fund;
    }

    function removeFundFromFundList(fundId: string): void {
        const index = allFunds.value.findIndex((f: Fund) => f.id === fundId);
        if (index >= 0) {
            allFunds.value.splice(index, 1);
        }
        delete allFundsMap.value[fundId];

        // If current fund was deleted, switch to another fund
        if (currentFundId.value === fundId) {
            currentFundId.value = allFunds.value.length > 0 ? (allFunds.value[0]?.id || null) : null;
        }
    }

    function setCurrentFund(fundId: string): boolean {
        if (allFundsMap.value[fundId]) {
            const previousFundId = currentFundId.value;
            currentFundId.value = fundId;
            // Store in localStorage for persistence
            localStorage.setItem('currentFundId', fundId);

            // Return whether the fund actually changed
            return previousFundId !== fundId;
        }
        return false;
    }

    function updateFundListInvalidState(invalidState: boolean): void {
        fundListStateInvalid.value = invalidState;
    }

    function resetFunds(): void {
        allFunds.value = [];
        allFundsMap.value = {};
        currentFundId.value = null;
        fundListStateInvalid.value = true;
        fundListLoading.value = false;
        fundListLoadingPromise = null;
        ensureFundLoadedPromise = null;
        localStorage.removeItem('currentFundId');
    }

    function loadAllFunds({ force }: { force: boolean }): Promise<Fund[]> {
        if (!force && !fundListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allFunds.value);
            });
        }

        // If already loading, return the existing promise
        if (fundListLoading.value && fundListLoadingPromise) {
            console.log('Fund store - Already loading funds, returning existing promise');
            return fundListLoadingPromise;
        }

        fundListLoading.value = true;
        const promise = new Promise<Fund[]>((resolve, reject) => {
            services.getAllFunds().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve fund list' });
                    reject({ message: 'Unable to retrieve fund list' });
                    return;
                }

                if (fundListStateInvalid.value) {
                    updateFundListInvalidState(false);
                }

                const funds = Fund.ofMulti(data.result);
                
                // Restore current fund from localStorage BEFORE calling loadFundList
                const savedFundId = localStorage.getItem('currentFundId');
                console.log('Fund store - Restoring fund from localStorage:', savedFundId);
                console.log('Fund store - Available funds:', funds.map(f => f.id));
                
                // Create funds map first to check if saved fund exists
                const fundsMap: Record<string, Fund> = {};
                for (const fund of funds) {
                    fundsMap[fund.id] = fund;
                }
                
                if (savedFundId && fundsMap[savedFundId]) {
                    currentFundId.value = savedFundId;
                    console.log('Fund store - Successfully restored fund:', savedFundId);
                } else if (funds.length > 0) {
                    // If no saved fund or saved fund not found, set the first fund as current
                    const personal = funds.find(f => f.name === 'Personal' || f.name.toLowerCase().includes('personal'));
                    const defaultFund = personal || funds[0];
                    if (defaultFund) {
                        currentFundId.value = defaultFund.id;
                        localStorage.setItem('currentFundId', defaultFund.id);
                        console.log('Fund store - Set default fund:', defaultFund.id);
                    }
                } else if (savedFundId && !fundsMap[savedFundId]) {
                    console.warn('Fund store - Saved fund not found:', savedFundId, 'Available funds:', Object.keys(fundsMap));
                }
                
                // Now call loadFundList which won't override currentFundId since it's already set
                loadFundList(funds);

                fundListLoading.value = false;
                fundListLoadingPromise = null;
                resolve(funds);
            }).catch(error => {
                fundListLoading.value = false;
                fundListLoadingPromise = null;
                logger.error('failed to load fund list', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve fund list' });
                } else {
                    reject(error);
                }
            });
        });

        fundListLoadingPromise = promise;
        return promise;
    }

    function ensureFundLoaded(): Promise<string | null> {
        // If already have a current fund, return immediately
        if (currentFundId.value) {
            return Promise.resolve(currentFundId.value);
        }

        // If funds are loaded but no current fund is set
        if (allFunds.value.length > 0) {
            const personal = personalFund.value;
            const defaultFund = personal || allFunds.value[0];
            if (defaultFund) {
                currentFundId.value = defaultFund.id;
                localStorage.setItem('currentFundId', defaultFund.id);
                return Promise.resolve(defaultFund.id);
            } else {
                return Promise.resolve(null);
            }
        }

        // If already ensuring fund is loaded, return the existing promise
        if (ensureFundLoadedPromise) {
            console.log('Fund store - Already ensuring fund loaded, returning existing promise');
            return ensureFundLoadedPromise;
        }

        // Create new promise to load funds
        ensureFundLoadedPromise = loadAllFunds({ force: false }).then(() => {
            ensureFundLoadedPromise = null; // Clear the promise
            return currentFundId.value;
        }).catch(() => {
            ensureFundLoadedPromise = null; // Clear the promise
            return null;
        });

        return ensureFundLoadedPromise;
    }

    function getFund({ fundId }: { fundId: string }): Promise<Fund> {
        return new Promise((resolve, reject) => {
            services.getFund({ id: fundId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve fund' });
                    return;
                }

                const fund = Fund.of(data.result);
                resolve(fund);
            }).catch(error => {
                logger.error('failed to load fund info', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve fund' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveFund({ fund, isEdit }: { fund: Fund, isEdit: boolean }): Promise<Fund> {
        return new Promise((resolve, reject) => {
            let promise = null;

            if (!isEdit) {
                promise = services.addFund(fund.toCreateRequest());
            } else {
                promise = services.modifyFund(fund.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add fund' });
                    } else {
                        reject({ message: 'Unable to save fund' });
                    }
                    return;
                }

                const newFund = Fund.of(data.result);

                if (!isEdit) {
                    addFundToFundList(newFund);
                } else {
                    updateFundInFundList(newFund);
                }

                resolve(newFund);
            }).catch(error => {
                logger.error('failed to save fund', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!isEdit) {
                        reject({ message: 'Unable to add fund' });
                    } else {
                        reject({ message: 'Unable to save fund' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteFund({ fundId }: { fundId: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteFund({ id: fundId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this fund' });
                    return;
                }

                removeFundFromFundList(fundId);
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete fund', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this fund' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getFundMembers({ fundId }: { fundId: string }): Promise<FundMember[]> {
        return new Promise((resolve, reject) => {
            services.getFundMembers({ fundId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve fund members' });
                    return;
                }

                const members = FundMember.ofMulti(data.result);
                resolve(members);
            }).catch(error => {
                logger.error('failed to load fund members', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve fund members' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function addFundMember(req: FundMemberCreateRequest): Promise<FundMember> {
        return new Promise((resolve, reject) => {
            services.addFundMember(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to add fund member' });
                    return;
                }

                const member = FundMember.of(data.result);
                resolve(member);
            }).catch(error => {
                logger.error('failed to add fund member', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to add fund member' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function removeFundMember(req: FundMemberDeleteRequest): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.removeFundMember(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to remove fund member' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to remove fund member', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to remove fund member' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function linkFundMember(req: FundMemberLinkRequest): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.linkFundMember(req).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to link fund member' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to link fund member', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to link fund member' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        // states
        allFunds,
        allFundsMap,
        currentFundId,
        fundListStateInvalid,
        // computed states
        currentFund,
        personalFund,
        currentCurrency,
        // functions
        setCurrentFund,
        updateFundListInvalidState,
        resetFunds,
        loadAllFunds,
        ensureFundLoaded,
        getFund,
        saveFund,
        deleteFund,
        getFundMembers,
        addFundMember,
        removeFundMember,
        linkFundMember
    };
});