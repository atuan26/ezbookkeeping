<template>
    <v-autocomplete item-title="name" item-value="id" auto-select-first persistent-placeholder :disabled="disabled"
        :label="label" :placeholder="placeholder" :items="allFunds" :no-data-text="tt('No results')"
        :custom-filter="filterFund" v-model="currentFundValue">
        <template #append-inner>
            <small class="text-field-append-text smaller">{{ currentFundDisplayText }}</small>
        </template>

        <template #item="{ props, item }">
            <v-list-item :value="item.value" v-bind="props">
                <template #title>
                    <v-list-item-title>
                        <div class="d-flex align-center">
                            <span>{{ item.title }}</span>
                            <v-spacer style="min-width: 40px" />
                            <v-icon :icon="mdiCheck" v-if="currentFundValue === item.raw.id" />
                            <small class="text-field-append-text" v-if="currentFundValue !== item.raw.id">
                                {{ item.raw.memberCount }} {{ tt(item.raw.memberCount === 1 ? 'member' : 'members') }}
                            </small>
                        </div>
                    </v-list-item-title>
                </template>
                <template #subtitle v-if="item.raw.role">
                    <small class="text-caption">{{ getFundRoleText(item.raw.role) }}</small>
                </template>
            </v-list-item>
        </template>
    </v-autocomplete>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
import { useFundSwitching } from '@/composables/useFundSwitching.ts';

import { Fund, FundRole } from '@/models/fund.ts';

import {
    mdiCheck
} from '@mdi/js';

const props = defineProps<{
    disabled?: boolean;
    label?: string;
    placeholder?: string;
    modelValue: string | null;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string | null): void;
}>();

const { tt } = useI18n();
const fundsStore = useFundsStore();
const { switchToFund } = useFundSwitching();

const allFunds = computed<Fund[]>(() => fundsStore.allFunds);

const currentFundValue = computed<string | null>({
    get: () => props.modelValue,
    set: (value: string | null) => {
        emit('update:modelValue', value);
        if (value) {
            switchToFund(value);
        }
    }
});

const currentFundDisplayText = computed<string>(() => {
    if (!currentFundValue.value) return '';

    const fund = fundsStore.allFundsMap[currentFundValue.value];
    if (!fund) return '';

    return `${fund.memberCount} ${tt(fund.memberCount === 1 ? 'member' : 'members')}`;
});

function filterFund(value: string, query: string, item?: { value: unknown, raw: Fund }): boolean {
    if (!item) {
        return false;
    }

    const lowerCaseFilterContent = query.toLowerCase() || '';

    if (!lowerCaseFilterContent) {
        return true;
    }

    return item.raw.name.toLowerCase().indexOf(lowerCaseFilterContent) >= 0;
}

function getFundRoleText(role: FundRole): string {
    switch (role) {
        case FundRole.Owner:
            return tt('Owner');
        case FundRole.Member:
            return tt('Member');
        default:
            return '';
    }
}
</script>