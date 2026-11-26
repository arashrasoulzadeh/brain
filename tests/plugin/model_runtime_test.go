package plugin_test

import (
 "encoding/json"
 "errors"
 "path/filepath"
 "testing"

 "brain/internal/plugin"
)

func TestDeclaredModelPersistenceAndIsolation(t *testing.T) {
 storage,err:=plugin.OpenSQLiteStorage(filepath.Join(t.TempDir(),"models.sqlite"));if err!=nil { t.Fatal(err) };defer storage.Close()
 r:=plugin.NewRegistry(1)
 if err:=r.BindSQLiteStorage(storage);err!=nil { t.Fatal(err) }
 m:=plugin.Manifest{ID:"modeltest",Version:"1",HostAPI:1,Capabilities:[]string{"items"},Models:map[string]plugin.ModelSpec{"item":{Version:1,Fields:map[string]string{"name":"string"},Initial:"draft",Transitions:map[string]plugin.ModelTransition{"publish":{From:"draft",To:"published"}}}}}
 for _,action:=range []string{"create","read","list","update","delete","transition"} {
  route:=&plugin.ModelRoute{Name:"item",Action:action};if action=="transition" {route.Transition="publish"}
  m.Routes=append(m.Routes,plugin.Route{Operation:action,Function:action,Capability:"items",Helper:"host.model",Model:route})
 }
 if err:=r.Register(m);err!=nil { t.Fatal(err) }
 if err:=r.SetReadiness(m.ID,m.Version,true);err!=nil {t.Fatal(err)}
 for _,project:=range []string{"alpha","beta"} {
  if err:=r.Install(project,m.ID,m.Version);err!=nil {t.Fatal(err)}
  if err:=r.Grant(project,m.ID,m.Version,"items");err!=nil {t.Fatal(err)}
  if err:=r.Enable(project,m.ID,m.Version);err!=nil {t.Fatal(err)}
  if err:=r.BindWorker(project,m.ID,m.Version,&contractWorker{ready:true});err!=nil {t.Fatal(err)}
 }
 invoke:=func(project,action,payload string)(plugin.WorkerResponse,error){ return r.Invoke(plugin.WorkerRequest{Protocol:1,ProjectID:project,PluginID:m.ID,Version:m.Version,Operation:action,Payload:[]byte(payload)}) }
 created,err:=invoke("alpha","create",`{"id":"one","fields":{"name":"first"}}`);if err!=nil || !created.Success {t.Fatalf("create: %+v %v",created,err)}
 if _,err:=invoke("beta","read",`{"id":"one"}`);!errors.Is(err,plugin.ErrModelConflict) {t.Fatalf("cross-project: %v",err)}
 var record struct{ Revision string; State string }
 if err:=json.Unmarshal(created.Payload,&record);err!=nil || record.State!="draft" || record.Revision=="" {t.Fatalf("invalid record: %+v %v",record,err)}
 payload:=`{"id":"one","revision":"`+record.Revision+`"}`
 published,err:=invoke("alpha","transition",payload);if err!=nil || !published.Success {t.Fatalf("publish: %+v %v",published,err)}
 if _,err:=invoke("alpha","delete",payload);!errors.Is(err,plugin.ErrModelConflict) {t.Fatalf("stale revision: %v",err)}
 if err:=r.Revoke("alpha",m.ID,m.Version,"items");err!=nil {t.Fatal(err)}
 if _,err:=invoke("alpha","read",`{"id":"one"}`);err==nil {t.Fatal("revoked grant accepted")}
}
