<template>
    <f7-sheet swipe-to-close swipe-handler=".swipe-handler"
              class="fund-selection-sheet" :opened="show"
              @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar>
            <div class="swipe-handler"></div>
            <div class="left"></div>
            <div class="right">
                <f7-link sheet-close :text="tt('Done')"></f7-link>
            </div>
        </f7-toolbar>
        <f7-page-content>
            <f7-list dividers class="no-margin-vertical">
                <f7-list-item link="#" no-chevron
                              :title="fund.name"
                              :after="fund.memberCount + ' ' + tt(fund.memberCount === 1 ? 'member' : 'members')"
                              :class="{ 'list-item-selected': isSelected(fund) }"
                              :key="fund.id"
                              v-for="fund in allFunds"
                              @click="onFundClicked(fund)">
                    <template #content-start>
                        <f7-icon class="list-item-checked-icon" f7="checkmark_alt" 
                                 :style="{ 'color': isSelected(fund) ? '' : 'transparent' }"></f7-icon>
                    </template>
                    <template #subtitle>
                        <small class="text-caption">{{ getFundRoleText(fund.role) }}</small>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
import { useFundSwitching } from '@/composables/useFundSwitching.ts';

import { Fund, FundRole } from '@/models/fund.ts';

import { type Framework7Dom, scrollToSelectedItem } from '@/lib/ui/mobile.ts';

const props = defineProps<{
    modelValue: string | null;
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string | null): void;
    (e: 'update:show', value: boolean): void;
    (e: 'fund-changed', fundId: string): void;
}>();

const { tt } = useI18n();
const fundsStore = useFundsStore();
const { switchToFund } = useFundSwitching();

const currentValue = ref<string | null>(props.modelValue);

const allFunds = computed<Fund[]>(() => fundsStore.allFunds);

function isSelected(fund: Fund): boolean {
    return currentValue.value === fund.id;
}

function close(): void {
    emit('update:show', false);
}

function onFundClicked(fund: Fund): void {
    currentValue.value = fund.id;
    emit('update:modelValue', fund.id);
    switchToFund(fund.id);
    emit('fund-changed', fund.id);
    close();
}

function onSheetOpen(event: { $el: Framework7Dom }): void {
    currentValue.value = props.modelValue;
    scrollToSelectedItem(event.$el, '.page-content', 'li.list-item-selected');
}

function onSheetClosed(): void {
    close();
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

<style>
.fund-selection-sheet {
    height: 400px;
}
</style>