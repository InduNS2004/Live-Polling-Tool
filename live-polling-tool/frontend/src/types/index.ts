export type Option={id:string;text:string;votes:number};
export type Poll={id:string;question:string;options:Option[];status:'active'|'closed';createdAt:string};
export type Result={pollId:string;question:string;options:Option[];totalVotes:number};
export type User={id:string;email:string};
