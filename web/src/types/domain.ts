export type AssetStatus='active'|'warning'|'disabled'|'repairing'|'retired'
export interface Asset { id:string; code:string; kind:string; zone:string; status:AssetStatus; expiresAt:string }
export interface Hazard { id:string; assetCode:string; priority:'low'|'medium'|'high'|'critical'; status:string; owner:string; dueAt:string; evidence:number }
export interface InspectionTask { id:string; route:string; inspector:string; progress:number; status:string; window:string }
