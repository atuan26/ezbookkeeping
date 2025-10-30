<template>
    <v-autocomplete
        v-model="selectedMemberIds"
        :items="allMembers"
        item-title="name"
        item-value="memberId"
        multiple
        chips
        closable-chips
        :label="label"
        :placeholder="placeholder"
        :disabled="disabled"
        :no-data-text="tt('No members found')"
        :custom-filter="filterMember"
    >
        <template #chip="{ props, item }">
            <v-chip
                v-bind="props"
                :text="item.title"
                :prepend-icon="item.raw.role === FundRole.Owner ? mdiCrown : mdiAccount"
                size="small"
            />
        </template>

        <template #item="{ props, item }">
            <v-list-item v-bind="props">
                <template #prepend>
                    <v-icon 
                        :icon="item.raw.role === FundRole.Owner ? mdiCrown : mdiAccount"
                        :color="item.raw.role === FundRole.Owner ? 'amber' : 'grey'"
                        size="small"
                    />
                </template>
                
                <v-list-item-title>{{ item.title }}</v-list-item-title>
                <v-list-item-subtitle v-if="item.raw.email">
                    {{ item.raw.email }}
                </v-list-item-subtitle>
                
                <template #append>
                    <v-icon 
                        :icon="mdiCheck" 
                        v-if="selectedMemberIds.includes(item.value)"
                        color="primary"
                        size="small"
                    />
                </template>
            </v-list-item>
        </template>
    </v-autocomplete>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';

import { FundMember, FundRole } from '@/models/fund.ts';

import {
    mdiAccount,
    mdiCrown,
    mdiCheck
} from '@mdi/js';

const props = defineProps<{
    modelValue: string[];
    fundId?: string;
    label?: string;
    placeholder?: string;
    disabled?: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string[]): void;
}>();

const { tt } = useI18n();
const fundsStore = useFundsStore();

const allMembers = ref<FundMember[]>([]);
const loading = ref<boolean>(false);

const selectedMemberIds = computed<string[]>({
    get: () => props.modelValue,
    set: (value: string[]) => emit('update:modelValue', value)
});

watch(() => props.fundId, (newFundId) => {
    if (newFundId) {
        loadMembers(newFundId);
    } else {
        allMembers.value = [];
    }
}, { immediate: true });

async function loadMembers(fundId: string): Promise<void> {
    if (loading.value) return;
    
    loading.value = true;
    try {
        allMembers.value = await fundsStore.getFundMembers({ fundId });
    } catch (error) {
        console.error('Failed to load fund members:', error);
        allMembers.value = [];
    } finally {
        loading.value = false;
    }
}

function filterMember(value: string, query: string, item?: { value: unknown, raw: FundMember }): boolean {
    if (!item) {
        return false;
    }

    const lowerCaseFilterContent = query.toLowerCase() || '';

    if (!lowerCaseFilterContent) {
        return true;
    }

    const nameMatch = item.raw.name.toLowerCase().indexOf(lowerCaseFilterContent) >= 0;
    const emailMatch = item.raw.email && item.raw.email.toLowerCase().indexOf(lowerCaseFilterContent) >= 0;
    
    return nameMatch || Boolean(emailMatch);
}
</script>