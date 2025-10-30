<template>
    <v-menu>
        <template v-slot:activator="{ props }">
            <v-btn
                variant="text"
                :prepend-icon="mdiFolderMultiple"
                v-bind="props"
                class="fund-select-button"
            >
                {{ currentFundName }}
                <v-icon :icon="mdiChevronDown" size="small" class="ml-1" />
            </v-btn>
        </template>
        
        <v-list density="compact" min-width="280">
            <v-list-item
                v-for="fund in allFunds"
                :key="fund.id"
                :value="fund.id"
                :active="fund.id === currentFundId"
            >
                <template #prepend>
                    <v-icon 
                        :icon="mdiCheck" 
                        size="small"
                        :style="{ visibility: fund.id === currentFundId ? 'visible' : 'hidden' }"
                    />
                </template>
                
                <div class="flex-grow-1" @click="selectFund(fund.id)" style="cursor: pointer;">
                    <v-list-item-title>{{ fund.name }}</v-list-item-title>
                    <v-list-item-subtitle>
                        <small class="text-caption">
                            {{ getFundRoleText(fund.role) }} • 
                            {{ fund.memberCount }} {{ tt(fund.memberCount === 1 ? 'member' : 'members') }}
                        </small>
                    </v-list-item-subtitle>
                </div>
                
                <template #append v-if="fund.role === FundRole.Owner">
                    <v-btn
                        icon
                        size="small"
                        variant="text"
                        @click.stop="manageFund(fund.id)"
                        :title="tt('Manage Fund')"
                    >
                        <v-icon :icon="mdiCog" size="small" />
                    </v-btn>
                </template>
            </v-list-item>
            
            <v-divider v-if="allFunds.length > 0" />
            
            <v-list-item @click="$emit('create-fund')">
                <template #prepend>
                    <v-icon :icon="mdiPlus" size="small" />
                </template>
                <v-list-item-title>{{ tt('Create Fund') }}</v-list-item-title>
            </v-list-item>
            

        </v-list>
    </v-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
import { useFundSwitching } from '@/composables/useFundSwitching.ts';

import { FundRole } from '@/models/fund.ts';

import {
    mdiCheck,
    mdiChevronDown,
    mdiFolderMultiple,
    mdiPlus,
    mdiCog
} from '@mdi/js';

const emit = defineEmits<{
    (e: 'create-fund'): void;
    (e: 'manage-fund', fundId: string): void;
}>();

const { tt } = useI18n();
const router = useRouter();
const fundsStore = useFundsStore();
const { switchToFund } = useFundSwitching();

const allFunds = computed(() => fundsStore.allFunds);
const currentFund = computed(() => fundsStore.currentFund);
const currentFundId = computed(() => fundsStore.currentFundId);

const currentFundName = computed(() => {
    return currentFund.value ? currentFund.value.name : tt('Select Fund hahahah');
});

function selectFund(fundId: string): void {
    switchToFund(fundId);
    // Navigate to dashboard after fund change to refresh all data
    router.push('/');
}

function manageFund(fundId: string): void {
    emit('manage-fund', fundId);
}

function getFundRoleText(role: FundRole): string {
    switch (role) {
        case FundRole.Owner:
            return tt('Owner');
        case FundRole.Member:
            return tt('Member');
        default:
            return tt('Unknown');
    }
}
</script>

<style scoped>
.fund-select-button {
    text-transform: none;
}
</style>