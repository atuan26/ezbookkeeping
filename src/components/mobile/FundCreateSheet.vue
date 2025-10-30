<template>
    <f7-sheet swipe-to-close swipe-handler=".swipe-handler"
              class="fund-create-sheet" :opened="show"
              @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar>
            <div class="swipe-handler"></div>
            <div class="left">
                <f7-link @click="close">{{ tt('Cancel') }}</f7-link>
            </div>
            <div class="right">
                <f7-link 
                    :class="{ 'disabled': !isFormValid || loading }"
                    @click="save"
                >
                    {{ isEdit ? tt('Save') : tt('Create') }}
                </f7-link>
            </div>
        </f7-toolbar>
        
        <f7-page-content>
            <f7-block-title>{{ isEdit ? tt('Edit Fund') : tt('Create Fund') }}</f7-block-title>
            
            <f7-list>
                <f7-list-input
                    v-model:value="fundForm.name"
                    :label="tt('Fund Name')"
                    type="text"
                    :maxlength="64"
                    :error-message="nameError"
                    :error-message-force="!!nameError"
                    required
                    clear-button
                    @input="validateName"
                />
                
                <f7-list-item
                    :title="tt('Default Currency')"
                    smart-select
                    :smart-select-params="currencySelectParams"
                    :disabled="isEdit"
                >
                    <select v-model="fundForm.defaultCurrency" name="currency">
                        <option 
                            v-for="currency in allCurrencies"
                            :key="currency.currencyCode"
                            :value="currency.currencyCode"
                        >
                            {{ currency.displayName }}
                        </option>
                    </select>
                </f7-list-item>
            </f7-list>
            
            <f7-block v-if="loading" class="text-align-center">
                <f7-preloader />
            </f7-block>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
// import { useUserStore } from '@/stores/user.ts';

import { Fund } from '@/models/fund.ts';

const props = defineProps<{
    show: boolean;
    fund?: Fund | null;
}>();

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'fund-saved', fund: Fund): void;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (e: 'error', error: any): void;
}>();

const { tt, getAllCurrencies } = useI18n();
const fundsStore = useFundsStore();
// const userStore = useUserStore();

const loading = ref<boolean>(false);
const nameError = ref<string>('');

const fundForm = ref({
    name: '',
    defaultCurrency: ''
});

const allCurrencies = computed(() => getAllCurrencies());

const isEdit = computed(() => !!props.fund);

const isFormValid = computed(() => {
    return fundForm.value.name.trim().length > 0 && 
           fundForm.value.name.length <= 64 &&
           fundForm.value.defaultCurrency.length > 0 &&
           !nameError.value;
});

const currencySelectParams = {
    searchbar: true,
    searchbarPlaceholder: tt('Search currencies'),
    closeOnSelect: true
};

watch(() => props.show, (newVal) => {
    if (newVal) {
        resetForm();
    }
});

watch(() => props.fund, (newVal) => {
    if (newVal) {
        fundForm.value = {
            name: newVal.name,
            defaultCurrency: newVal.defaultCurrency
        };
    }
});

function resetForm(): void {
    if (props.fund) {
        fundForm.value = {
            name: props.fund.name,
            defaultCurrency: props.fund.defaultCurrency
        };
    } else {
        fundForm.value = {
            name: '',
            defaultCurrency: fundsStore.currentCurrency
        };
    }
    nameError.value = '';
}

function validateName(): void {
    const name = fundForm.value.name;
    if (!name || name.trim().length === 0) {
        nameError.value = tt('Fund name is required');
    } else if (name.length > 64) {
        nameError.value = tt('Fund name must be less than 64 characters');
    } else {
        nameError.value = '';
    }
}

function close(): void {
    emit('update:show', false);
}

function onSheetOpen(): void {
    resetForm();
}

function onSheetClosed(): void {
    close();
}

async function save(): Promise<void> {
    if (!isFormValid.value) {
        return;
    }

    loading.value = true;

    try {
        let fund: Fund;
        
        if (isEdit.value && props.fund) {
            fund = new Fund({
                id: props.fund.id,
                name: fundForm.value.name.trim(),
                role: props.fund.role,
                memberCount: props.fund.memberCount,
                defaultCurrency: fundForm.value.defaultCurrency,
                createdAt: props.fund.createdAt
            });
        } else {
            fund = new Fund({
                id: '0', // Will be set by server
                name: fundForm.value.name.trim(),
                role: 1, // Owner
                memberCount: 1,
                defaultCurrency: fundForm.value.defaultCurrency,
                createdAt: Date.now()
            });
        }

        const savedFund = await fundsStore.saveFund({
            fund,
            isEdit: isEdit.value
        });

        emit('fund-saved', savedFund);
        close();
    } catch (error) {
        emit('error', error);
    } finally {
        loading.value = false;
    }
}
</script>

<style>
.fund-create-sheet {
    height: 400px;
}
</style>