<template>
    <div class="transaction-member-chips" v-if="members.length > 0">
        <template v-if="isDesktop">
            <v-chip
                v-for="member in displayMembers"
                :key="member.memberId"
                size="x-small"
                variant="outlined"
                :prepend-icon="member.role === FundRole.Owner ? mdiCrown : mdiAccount"
                class="mr-1 mb-1"
            >
                {{ member.name }}
            </v-chip>
            <v-chip
                v-if="remainingCount > 0"
                size="x-small"
                variant="outlined"
                class="mr-1 mb-1"
            >
                +{{ remainingCount }}
            </v-chip>
        </template>
        
        <template v-else>
            <span class="member-chips-mobile" v-if="members.length > 0">
                <f7-icon 
                    :f7="members[0]?.role === FundRole.Owner ? 'crown_fill' : 'person_fill'"
                    size="12"
                    :color="members[0]?.role === FundRole.Owner ? 'orange' : 'gray'"
                />
                {{ displayText }}
            </span>
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
// import { useI18n } from '@/locales/helpers.ts';

import { FundMember, FundRole } from '@/models/fund.ts';

import {
    mdiAccount,
    mdiCrown
} from '@mdi/js';

const props = defineProps<{
    members: FundMember[];
    maxDisplay?: number;
    isDesktop?: boolean;
}>();

// const { tt } = useI18n();

const maxDisplay = computed(() => props.maxDisplay || 2);

const displayMembers = computed(() => {
    return props.members.slice(0, maxDisplay.value);
});

const remainingCount = computed(() => {
    return Math.max(0, props.members.length - maxDisplay.value);
});

const displayText = computed(() => {
    if (props.members.length === 0) return '';
    
    if (props.members.length === 1) {
        return props.members[0]?.name || '';
    } else if (props.members.length === 2) {
        return `${props.members[0]?.name || ''}, ${props.members[1]?.name || ''}`;
    } else {
        return `${props.members[0]?.name || ''} +${props.members.length - 1}`;
    }
});
</script>

<style scoped>
.transaction-member-chips {
    display: inline-block;
}

.member-chips-mobile {
    font-size: 0.75rem;
    color: var(--f7-text-color);
    display: inline-flex;
    align-items: center;
    gap: 4px;
}
</style>