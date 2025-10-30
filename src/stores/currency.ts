import { computed } from 'vue';
import { defineStore } from 'pinia';
import { useFundsStore } from './fund.ts';

export const useCurrencyStore = defineStore('currency', () => {
    const fundsStore = useFundsStore();

    // Current effective currency - fund currency takes precedence over user currency
    const currentCurrency = computed<string>(() => {
        const currentFund = fundsStore.currentFund;
        if (currentFund && currentFund.defaultCurrency) {
            return currentFund.defaultCurrency;
        }
        // Fallback to user's default currency if no fund is selected
        return fundsStore.currentCurrency;
    });

    // Check if we're using fund currency (vs user currency)
    const isUsingFundCurrency = computed<boolean>(() => {
        const currentFund = fundsStore.currentFund;
        return !!(currentFund && currentFund.defaultCurrency);
    });

    return {
        // computed states
        currentCurrency,
        isUsingFundCurrency
    };
});