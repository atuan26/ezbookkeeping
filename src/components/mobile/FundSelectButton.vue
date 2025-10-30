<template>
    <f7-button 
        fill 
        class="fund-select-button"
        @click="showFundSelection = true"
    >
        <f7-icon f7="folder_fill" class="margin-right-half" />
        {{ currentFundName }}
        <f7-icon f7="chevron_down" size="14" class="margin-left-half" />
    </f7-button>

    <FundSelectionSheet
        v-model="selectedFundId"
        v-model:show="showFundSelection"
    />
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
import { useFundSwitching } from '@/composables/useFundSwitching.ts';

import FundSelectionSheet from './FundSelectionSheet.vue';

const { tt } = useI18n();
const fundsStore = useFundsStore();
const { switchToFund } = useFundSwitching();

const showFundSelection = ref<boolean>(false);

const currentFund = computed(() => fundsStore.currentFund);
const selectedFundId = computed({
    get: () => fundsStore.currentFundId,
    set: (value: string | null) => {
        if (value) {
            switchToFund(value);
        }
    }
});

const currentFundName = computed(() => {
    return currentFund.value ? currentFund.value.name : tt('Select Fund');
});
</script>

<style scoped>
.fund-select-button {
    --f7-button-text-transform: none;
    margin: 8px;
}
</style>