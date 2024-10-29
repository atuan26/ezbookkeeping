import { watch } from 'vue';
import { useFundsStore } from '@/stores/fund.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useStatisticsStore } from '@/stores/statistics.ts';
import { useOverviewStore } from '@/stores/overview.ts';

export function useFundSwitching() {
    const fundsStore = useFundsStore();
    const accountsStore = useAccountsStore();
    const transactionsStore = useTransactionsStore();
    const transactionCategoriesStore = useTransactionCategoriesStore();
    const transactionTagsStore = useTransactionTagsStore();
    const transactionTemplatesStore = useTransactionTemplatesStore();
    const statisticsStore = useStatisticsStore();
    const overviewStore = useOverviewStore();

    function invalidateStoresForFundChange(): void {
        // Invalidate accounts
        accountsStore.updateAccountListInvalidState(true);
        
        // Invalidate transactions
        transactionsStore.updateTransactionListInvalidState(true);
        
        // Invalidate categories
        transactionCategoriesStore.updateTransactionCategoryListInvalidState(true);
        
        // Invalidate tags
        transactionTagsStore.updateTransactionTagListInvalidState(true);
        
        // Invalidate templates (invalidate all template types)
        transactionTemplatesStore.updateTransactionTemplateListInvalidState(0, true); // Income templates
        transactionTemplatesStore.updateTransactionTemplateListInvalidState(1, true); // Expense templates
        transactionTemplatesStore.updateTransactionTemplateListInvalidState(2, true); // Transfer templates
        
        // Invalidate statistics
        statisticsStore.updateTransactionStatisticsInvalidState(true);
        
        // Invalidate overview
        overviewStore.updateTransactionOverviewInvalidState(true);
    }

    // Watch for fund changes and invalidate related stores
    watch(() => fundsStore.currentFundId, (newFundId, oldFundId) => {
        if (newFundId !== oldFundId && oldFundId !== null) {
            // Invalidate all fund-dependent data
            invalidateStoresForFundChange();
        }
    });

    function switchToFund(fundId: string): void {
        const changed = fundsStore.setCurrentFund(fundId);
        console.log("Fund: ", fundId, changed)
        if (changed) {
            // Currency will automatically update via the currency store computed property
            // No need to call API - it's purely UI-based now
            const newFund = fundsStore.allFundsMap[fundId];
            if (newFund) {
                console.log('Fund switched to:', newFund.name, 'Currency will be:', newFund.defaultCurrency);
            }
            
            invalidateStoresForFundChange();
            // Force reload data for the new fund
            forceReloadAllData();
        }
    }

    function forceReloadAllData(): void {
        // Force reload basic data that's used across all pages
        accountsStore.loadAllAccounts({ force: false }).catch(console.error);
        transactionCategoriesStore.loadAllCategories({ force: false }).catch(console.error);
        transactionTagsStore.loadAllTags({ force: false }).catch(console.error);
        
        // Only reload overview data (for dashboard) - don't reload statistics
        // Statistics will be loaded when user navigates to statistics page
        overviewStore.loadTransactionOverview({ force: true }).catch(console.error);
        
        // Note: Statistics are NOT reloaded here to avoid unnecessary API calls
        // They will be loaded on-demand when user visits the statistics page
    }

    return {
        switchToFund,
        invalidateStoresForFundChange,
        forceReloadAllData
    };
}