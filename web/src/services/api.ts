import type { Asset,Hazard,InspectionTask } from '../types/domain'
const headers={'Authorization':'Bearer demo','X-Actor-ID':'user-demo','X-Mall-ID':'mall-demo','X-Team-ID':'team-inspect','X-Roles':'safety_manager,inspection_lead','Content-Type':'application/json'}
export async function listAssets():Promise<Asset[]>{return demoAssets}
export async function listHazards():Promise<Hazard[]>{return demoHazards}
export async function listTasks():Promise<InspectionTask[]>{return demoTasks}
export async function post<T>(path:string,body:unknown):Promise<T>{const response=await fetch(`/api/v1${path}`,{method:'POST',headers,body:JSON.stringify(body)});if(!response.ok)throw await response.json();return response.json()}
export const demoAssets:Asset[]=[{id:'a1',code:'F1-EXT-001',kind:'灭火器',zone:'一层东区',status:'active',expiresAt:'2027-05-20'},{id:'a2',code:'B1-HYD-004',kind:'消火栓',zone:'地下一层',status:'warning',expiresAt:'2028-01-01'},{id:'a3',code:'F3-SMK-021',kind:'烟感',zone:'三层餐饮区',status:'disabled',expiresAt:'2030-08-12'}]
export const demoHazards:Hazard[]=[{id:'h1',assetCode:'B1-HYD-004',priority:'high',status:'assigned',owner:'维修二组',dueAt:'今天 18:00',evidence:1},{id:'h2',assetCode:'F3-SMK-021',priority:'critical',status:'awaiting_review',owner:'弱电组',dueAt:'已逾期 45 分钟',evidence:3}]
export const demoTasks:InspectionTask[]=[{id:'t1',route:'一层晨检路线',inspector:'陈屿',progress:72,status:'running',window:'08:00 - 10:00'},{id:'t2',route:'地下设备间',inspector:'待认领',progress:0,status:'pending',window:'10:00 - 12:00'}]
