export enum FundRole {
    Owner = 1,
    Member = 2
}

export interface FundInfoResponse {
    id: string;
    name: string;
    role: FundRole;
    memberCount: number;
    defaultCurrency: string;
    createdAt: number;
}

export interface FundMemberResponse {
    memberId: string;
    name: string;
    email: string;
    role: FundRole;
    linkedUid: string;
    isLinked: boolean;
}

export interface FundCreateRequest {
    name: string;
    defaultCurrency: string;
    clientSessionId?: string;
}

export interface FundModifyRequest {
    id: string;
    name: string;
    defaultCurrency: string;
}

export interface FundDeleteRequest {
    id: string;
}

export interface FundMemberCreateRequest {
    fundId: string;
    name: string;
    email?: string;
    clientSessionId?: string;
}

export interface FundMemberLinkRequest {
    memberId: string;
    linkedUid: string;
}

export interface FundMemberDeleteRequest {
    fundId: string;
    memberId: string;
}

export class Fund {
    id: string;
    name: string;
    role: FundRole;
    memberCount: number;
    defaultCurrency: string;
    createdAt: number;

    constructor(data: FundInfoResponse) {
        this.id = data.id;
        this.name = data.name;
        this.role = data.role;
        this.memberCount = data.memberCount;
        this.defaultCurrency = data.defaultCurrency;
        this.createdAt = data.createdAt;
    }

    static of(data: FundInfoResponse): Fund {
        return new Fund(data);
    }

    static ofMulti(data: FundInfoResponse[]): Fund[] {
        return data.map(item => Fund.of(item));
    }

    toCreateRequest(): FundCreateRequest {
        return {
            name: this.name,
            defaultCurrency: this.defaultCurrency
        };
    }

    toModifyRequest(): FundModifyRequest {
        return {
            id: this.id,
            name: this.name,
            defaultCurrency: this.defaultCurrency
        };
    }
}

export class FundMember {
    memberId: string;
    name: string;
    email: string;
    role: FundRole;
    linkedUid: string;
    isLinked: boolean;

    constructor(data: FundMemberResponse) {
        this.memberId = data.memberId;
        this.name = data.name;
        this.email = data.email;
        this.role = data.role;
        this.linkedUid = data.linkedUid;
        this.isLinked = data.isLinked;
    }

    static of(data: FundMemberResponse): FundMember {
        return new FundMember(data);
    }

    static ofMulti(data: FundMemberResponse[]): FundMember[] {
        return data.map(item => FundMember.of(item));
    }

    get isOwner(): boolean {
        return this.role === FundRole.Owner;
    }

    get isMember(): boolean {
        return this.role === FundRole.Member;
    }
}