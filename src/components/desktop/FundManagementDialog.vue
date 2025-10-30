<template>
    <v-dialog v-model="showDialog" max-width="600px" persistent>
        <v-card v-if="managingFund">
            <v-card-title>
                <span class="text-h6">{{ tt('Manage Fund') }}: {{ managingFund.name }}</span>
            </v-card-title>
            
            <v-card-text>
                <v-tabs v-model="activeTab">
                    <v-tab value="settings">{{ tt('Settings') }}</v-tab>
                    <v-tab value="members" v-if="managingFund.role === FundRole.Owner">
                        {{ tt('Members') }}
                    </v-tab>
                </v-tabs>
                
                <v-tabs-window v-model="activeTab">
                    <!-- Settings Tab -->
                    <v-tabs-window-item value="settings">
                        <div class="mt-4">
                            <v-row class="mb-3">
                                <v-col>
                                    <h3 class="text-h6 mb-3">{{ tt('Fund Information') }}</h3>
                                    <v-list>
                                        <v-list-item>
                                            <v-list-item-title>{{ tt('Name') }}</v-list-item-title>
                                            <v-list-item-subtitle>{{ managingFund.name }}</v-list-item-subtitle>
                                            <template #append v-if="managingFund.role === FundRole.Owner">
                                                <v-btn
                                                    icon
                                                    size="small"
                                                    variant="text"
                                                    @click="editFund(managingFund)"
                                                >
                                                    <v-icon :icon="mdiPencil" />
                                                </v-btn>
                                            </template>
                                        </v-list-item>
                                        <v-list-item>
                                            <v-list-item-title>{{ tt('Role') }}</v-list-item-title>
                                            <v-list-item-subtitle>{{ getFundRoleText(managingFund.role) }}</v-list-item-subtitle>
                                        </v-list-item>
                                        <v-list-item>
                                            <v-list-item-title>{{ tt('Members') }}</v-list-item-title>
                                            <v-list-item-subtitle>
                                                {{ managingFund.memberCount }} {{ tt(managingFund.memberCount === 1 ? 'member' : 'members') }}
                                            </v-list-item-subtitle>
                                        </v-list-item>
                                    </v-list>
                                </v-col>
                            </v-row>
                            
                            <v-row v-if="managingFund.role === FundRole.Owner && allFunds.length > 1">
                                <v-col>
                                    <h3 class="text-h6 mb-3">{{ tt('Danger Zone') }}</h3>
                                    <v-btn
                                        color="error"
                                        variant="outlined"
                                        prepend-icon="mdi-delete"
                                        @click="confirmDeleteFund(managingFund)"
                                    >
                                        {{ tt('Delete Fund') }}
                                    </v-btn>
                                </v-col>
                            </v-row>
                        </div>
                    </v-tabs-window-item>
                    
                    <!-- Members Tab -->
                    <v-tabs-window-item value="members" v-if="managingFund && managingFund.role === FundRole.Owner">
                        <div class="mt-4">
                            <v-row class="mb-3">
                                <v-col>
                                    <v-btn
                                        color="primary"
                                        prepend-icon="mdi-account-plus"
                                        @click="showAddMember = true"
                                        :disabled="!managingFund"
                                    >
                                        {{ tt('Add Member') }}
                                    </v-btn>
                                </v-col>
                            </v-row>
                            
                            <v-list v-if="fundMembers.length > 0">
                                <v-list-item
                                    v-for="member in fundMembers"
                                    :key="member.memberId"
                                >
                                    <template #prepend>
                                        <v-icon 
                                            :icon="member.role === FundRole.Owner ? mdiCrown : mdiAccount"
                                            :color="member.role === FundRole.Owner ? 'amber' : 'grey'"
                                        />
                                    </template>
                                    
                                    <v-list-item-title>{{ member.name }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ member.email || tt('No email') }} • {{ getFundRoleText(member.role) }}
                                        <span v-if="member.isLinked" class="text-success">
                                            • {{ tt('Linked') }}
                                        </span>
                                    </v-list-item-subtitle>
                                    
                                    <template #append v-if="member.role !== FundRole.Owner">
                                        <v-btn
                                            icon
                                            size="small"
                                            variant="text"
                                            color="error"
                                            @click="confirmRemoveMember(member)"
                                        >
                                            <v-icon :icon="mdiDelete" />
                                        </v-btn>
                                    </template>
                                </v-list-item>
                            </v-list>
                            
                            <v-alert v-else type="info" class="mt-4">
                                {{ tt('No members found. Add members to collaborate on this fund.') }}
                            </v-alert>
                        </div>
                    </v-tabs-window-item>
                </v-tabs-window>
            </v-card-text>
            
            <v-card-actions>
                <v-spacer />
                <v-btn
                    color="grey-darken-1"
                    variant="text"
                    @click="close"
                >
                    {{ tt('Close') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Create/Edit Fund Dialog -->
    <FundCreateDialog
        v-model:show="showCreateFund"
        :fund="editingFund"
        @fund-saved="onFundSaved"
        @error="onError"
    />

    <!-- Add Member Dialog -->
    <FundMemberCreateDialog
        v-model:show="showAddMember"
        :fund-id="managingFund?.id"
        @member-added="onMemberAdded"
        @error="onError"
    />

    <!-- Confirm Delete Fund Dialog -->
    <v-dialog v-model="showDeleteConfirm" max-width="400px">
        <v-card>
            <v-card-title>{{ tt('Delete Fund') }}</v-card-title>
            <v-card-text>
                {{ tt('Are you sure you want to delete this fund? This action cannot be undone.') }}
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn color="grey-darken-1" variant="text" @click="showDeleteConfirm = false">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn color="error" variant="text" @click="deleteFund" :loading="deleteLoading">
                    {{ tt('Delete') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Confirm Remove Member Dialog -->
    <v-dialog v-model="showRemoveMemberConfirm" max-width="400px">
        <v-card>
            <v-card-title>{{ tt('Remove Member') }}</v-card-title>
            <v-card-text>
                {{ tt('Are you sure you want to remove this member from the fund?') }}
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn color="grey-darken-1" variant="text" @click="showRemoveMemberConfirm = false">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn color="error" variant="text" @click="removeMember" :loading="removeLoading">
                    {{ tt('Remove') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useFundsStore } from '@/stores/fund.ts';

import { Fund, FundMember, FundRole } from '@/models/fund.ts';
import FundCreateDialog from './FundCreateDialog.vue';
import FundMemberCreateDialog from './FundMemberCreateDialog.vue';

import {
    mdiAccount,
    mdiCrown,
    mdiPencil,
    mdiDelete
} from '@mdi/js';

const props = defineProps<{
    show: boolean;
    fundId: string | null;
}>();

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (e: 'error', error: any): void;
}>();

const { tt } = useI18n();

const fundsStore = useFundsStore();

const activeTab = ref<string>('settings');
const fundMembers = ref<FundMember[]>([]);

const showCreateFund = ref<boolean>(false);
const showAddMember = ref<boolean>(false);
const showDeleteConfirm = ref<boolean>(false);
const showRemoveMemberConfirm = ref<boolean>(false);

const editingFund = ref<Fund | null>(null);
const deletingFund = ref<Fund | null>(null);
const removingMember = ref<FundMember | null>(null);

const deleteLoading = ref<boolean>(false);
const removeLoading = ref<boolean>(false);

const showDialog = computed({
    get: () => props.show,
    set: (value: boolean) => emit('update:show', value)
});

const allFunds = computed(() => fundsStore.allFunds);

const managingFund = computed(() => {
    if (!props.fundId) return null;
    return fundsStore.allFundsMap[props.fundId] || null;
});

watch(() => props.show, (newVal) => {
    if (newVal && managingFund.value) {
        loadFundMembers();
    }
});

// Remove selectFund function as it's no longer needed

function editFund(fund: Fund): void {
    editingFund.value = fund;
    showCreateFund.value = true;
}

function confirmDeleteFund(fund: Fund): void {
    deletingFund.value = fund;
    showDeleteConfirm.value = true;
}

function confirmRemoveMember(member: FundMember): void {
    removingMember.value = member;
    showRemoveMemberConfirm.value = true;
}

async function deleteFund(): Promise<void> {
    if (!deletingFund.value) return;

    const fundToDelete = deletingFund.value;
    deleteLoading.value = true;
    try {
        await fundsStore.deleteFund({ fundId: fundToDelete.id });
        showDeleteConfirm.value = false;
        deletingFund.value = null;
        
        // If the deleted fund was the current fund, switch to another fund
        if (fundsStore.currentFundId === fundToDelete.id) {
            const remainingFunds = fundsStore.allFunds.filter(f => f.id !== fundToDelete.id);
            if (remainingFunds.length > 0 && remainingFunds[0]) {
                fundsStore.setCurrentFund(remainingFunds[0].id);
            }
        }
        
        // Close the dialog since the fund was deleted
        showDialog.value = false;
    } catch (error) {
        emit('error', error);
    } finally {
        deleteLoading.value = false;
    }
}

async function removeMember(): Promise<void> {
    if (!removingMember.value || !managingFund.value) return;

    removeLoading.value = true;
    try {
        await fundsStore.removeFundMember({
            fundId: managingFund.value.id,
            memberId: removingMember.value.memberId
        });
        
        showRemoveMemberConfirm.value = false;
        removingMember.value = null;
        await loadFundMembers();
    } catch (error) {
        emit('error', error);
    } finally {
        removeLoading.value = false;
    }
}

async function loadFundMembers(): Promise<void> {
    if (!managingFund.value || managingFund.value.role !== FundRole.Owner) {
        fundMembers.value = [];
        return;
    }

    try {
        fundMembers.value = await fundsStore.getFundMembers({ fundId: managingFund.value.id });
    } catch (error) {
        emit('error', error);
    }
}

function onFundSaved(): void {
    editingFund.value = null;
    // Refresh fund list
    fundsStore.updateFundListInvalidState(true);
    fundsStore.loadAllFunds({ force: false });
}

function onMemberAdded(): void {
    loadFundMembers();
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function onError(error: any): void {
    emit('error', error);
}

function close(): void {
    showDialog.value = false;
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