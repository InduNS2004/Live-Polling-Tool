const API=import.meta.env.VITE_API_URL||'http://localhost:8080';
export const apiBase=API;
export const setToken=(_t:string|null)=>{};
export async function api<T>(path:string,options:RequestInit={}){const headers=new Headers(options.headers);headers.set('Content-Type','application/json');const r=await fetch(`${API}${path}`,{...options,headers,credentials:'include'});let body:any=null;try{body=await r.json()}catch{}if(!r.ok)throw new Error(body?.error?.message||'Request failed');return body?.data as T;}
