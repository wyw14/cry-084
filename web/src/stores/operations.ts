import { defineStore } from 'pinia'
import { listAssets,listHazards,listTasks } from '../services/api'
import type { Asset,Hazard,InspectionTask } from '../types/domain'
export const useOperationsStore=defineStore('operations',{state:()=>({assets:[] as Asset[],hazards:[] as Hazard[],tasks:[] as InspectionTask[],loading:false,lastUpdated:''}),getters:{criticalHazards:s=>s.hazards.filter(h=>h.priority==='critical'),disabledAssets:s=>s.assets.filter(a=>a.status==='disabled'),activeTasks:s=>s.tasks.filter(t=>t.status==='running')},actions:{async refresh(){this.loading=true;try{[this.assets,this.hazards,this.tasks]=await Promise.all([listAssets(),listHazards(),listTasks()]);this.lastUpdated=new Date().toISOString()}finally{this.loading=false}}}})
