<template>
    <f7-sheet swipe-to-close swipe-handler=".swipe-handler"
              class="fund-member-selection-sheet" :opened="show"
              @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar>
            <div class="swipe-handler"></div>
            <div class="left"></div>
            <div class="right">
                <f7-link sheet-close :text="tt('Done')"></f7-link>
            </div>
        </f7-toolbar>
        
        <f7-page-content>
            <f7-block-title>{{ tt('Select Members') }}</f7-block-title>
            
            <f7-list dividers class="no-margin-vertical">
                <f7-list-item
                    checkbox
                    :title="member.name"
                    :after="member.email || ''"
                    :checked="isSelected(member.memberId)"
                    :key="member.memberId"
                    v-for="member in allMembers"
                    @change="onMemberToggle(member.memberId, $event)"
                >
                    <template #media>
                        <f7-icon 
                            :f7="member.role === FundRole.Owner ? 'crown_fill' : 'person_fill'"
                            :color="member.role === FundRole.Owner ? 'orange' : 'gray'"
                        />
                    </template>
                    <template #subtitle>
                        <small class="text-caption">
                            {{ member.role === FundRole.Owner ? tt('Owner') : tt('Member') }}
                        </small>
                    </template>
                </f7-list-item>
            </f7-list>
            
            <f7-block v-if="allMembers.length === 0 && !loading">
                <p class="text-align-center">{{ tt('No members found') }}</p>
            </f7-block>
            
            <f7-block v-if="loading" class="text-align-center">
                <f7-preloader />
            </f7-block>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';

import { FundMember, FundRole } from '@/models/fund.ts';

const props = defineProps<{
    modelValue: string[];
    fundId?: string;
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string[]): void;
    (e: 'update:show', value: boolean): void;
}>();

const { tt } = useI18n();
const fundsStore = useFundsStore();

const allMembers = ref<FundMember[]>([]);
const loading = ref<boolean>(false);
const currentValue = ref<string[]>([]);

watch(() => props.show, (newVal) => {
    if (newVal) {
        currentValue.value = [...props.modelValue];
        if (props.fundId) {
            loadMembers(props.fundId);
        }
    }
});

watch(() => props.fundId, (newFundId) => {
    if (newFundId && props.show) {
        loadMembers(newFundId);
    } else {
        allMembers.value = [];
    }
});

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

function isSelected(memberId: string): boolean {
    return currentValue.value.includes(memberId);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onMemberToggle(memberId: string, event: any): void {
    const isChecked = event.target.checked;
    
    if (isChecked) {
        if (!currentValue.value.includes(memberId)) {
            currentValue.value.push(memberId);
        }
    } else {
        const index = currentValue.value.indexOf(memberId);
        if (index > -1) {
            currentValue.value.splice(index, 1);
        }
    }
    
    emit('update:modelValue', [...currentValue.value]);
}

function close(): void {
    emit('update:show', false);
}

function onSheetOpen(): void {
    currentValue.value = [...props.modelValue];
}

function onSheetClosed(): void {
    close();
}
</script>

<style>
.fund-member-selection-sheet {
    height: 400px;
}
</style>