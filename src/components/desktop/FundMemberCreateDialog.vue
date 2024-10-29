<template>
    <v-dialog v-model="showDialog" max-width="500px" persistent>
        <v-card>
            <v-card-title>
                <span class="text-h6">{{ tt('Add Member') }}</span>
            </v-card-title>
            
            <v-card-text>
                <v-form ref="form" v-model="formValid">
                    <v-text-field
                        v-model="memberForm.name"
                        :label="tt('Member Name')"
                        :rules="nameRules"
                        :counter="64"
                        required
                        autofocus
                    />
                    
                    <v-text-field
                        v-model="memberForm.email"
                        :label="tt('Email (Optional)')"
                        :rules="emailRules"
                        type="email"
                    />
                    
                    <v-alert type="info" class="mt-3">
                        {{ tt('Members will have read-only access to the fund. They can be linked to existing users later.') }}
                    </v-alert>
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
                    {{ tt('Add Member') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';

import { FundMember } from '@/models/fund.ts';

const props = defineProps<{
    show: boolean;
    fundId?: string;
}>();

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'member-added', member: FundMember): void;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (e: 'error', error: any): void;
}>();

const { tt } = useI18n();
const fundsStore = useFundsStore();

const form = ref();
const formValid = ref<boolean>(false);
const loading = ref<boolean>(false);

const memberForm = ref({
    name: '',
    email: ''
});

const showDialog = computed({
    get: () => props.show,
    set: (value: boolean) => emit('update:show', value)
});

const nameRules = [
    (v: string) => !!v || tt('Member name is required'),
    (v: string) => (v && v.length <= 64) || tt('Member name must be less than 64 characters'),
    (v: string) => (v && v.trim().length > 0) || tt('Member name cannot be empty')
];

const emailRules = [
    (v: string) => !v || /.+@.+\..+/.test(v) || tt('Email must be valid')
];

watch(() => props.show, (newVal) => {
    if (newVal) {
        resetForm();
    }
});

function resetForm(): void {
    memberForm.value = {
        name: '',
        email: ''
    };
    
    if (form.value) {
        form.value.resetValidation();
    }
}

function close(): void {
    showDialog.value = false;
}

async function save(): Promise<void> {
    if (!form.value.validate() || !props.fundId) {
        return;
    }

    loading.value = true;

    try {
        const member = await fundsStore.addFundMember({
            fundId: props.fundId,
            name: memberForm.value.name.trim(),
            email: memberForm.value.email.trim() || undefined
        });

        emit('member-added', member);
        close();
    } catch (error) {
        emit('error', error);
    } finally {
        loading.value = false;
    }
}
</script>