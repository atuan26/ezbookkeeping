<template>
    <v-dialog v-model="showDialog" max-width="500px" persistent>
        <v-card>
            <v-card-title>
                <span class="text-h6">{{ isEdit ? tt('Edit Fund') : tt('Create Fund') }}</span>
            </v-card-title>
            
            <v-card-text>
                <v-form ref="form" v-model="formValid">
                    <v-text-field
                        v-model="fundForm.name"
                        :label="tt('Fund Name')"
                        :rules="nameRules"
                        :counter="64"
                        required
                        autofocus
                    />
                    
                    <CurrencySelect
                        v-model="fundForm.defaultCurrency"
                        :label="tt('Default Currency')"
                        :disabled="isEdit"
                    />
                </v-form>
            </v-card-text>
            
            <v-card-actions>
                <v-spacer />
                <v-btn
                    color="grey-darken-1"
                    variant="text"
                    @click="close"
                >
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn
                    color="primary"
                    variant="text"
                    :disabled="!formValid || loading"
                    :loading="loading"
                    @click="save"
                >
                    {{ isEdit ? tt('Save') : tt('Create') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';
// import { useUserStore } from '@/stores/user.ts';

import { Fund } from '@/models/fund.ts';
import CurrencySelect from './CurrencySelect.vue';

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

const { tt } = useI18n();
const fundsStore = useFundsStore();
// const userStore = useUserStore();

const form = ref();
const formValid = ref<boolean>(false);
const loading = ref<boolean>(false);

const fundForm = ref({
    name: '',
    defaultCurrency: ''
});

const showDialog = computed({
    get: () => props.show,
    set: (value: boolean) => emit('update:show', value)
});

const isEdit = computed(() => !!props.fund);

const nameRules = [
    (v: string) => !!v || tt('Fund name is required'),
    (v: string) => (v && v.length <= 64) || tt('Fund name must be less than 64 characters'),
    (v: string) => (v && v.trim().length > 0) || tt('Fund name cannot be empty')
];

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
    
    if (form.value) {
        form.value.resetValidation();
    }
}

function close(): void {
    showDialog.value = false;
}

async function save(): Promise<void> {
    if (!form.value.validate()) {
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