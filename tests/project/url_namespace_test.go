package project_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "brain/internal/project"
)

func TestURLNamespaceResolvesHostProjectAndRejectsClientConflicts(t *testing.T) {
    store:=project.NewStore(project.Project{ID:"internal-alpha",Slug:"alpha",Active:true},project.Project{ID:"beta",Active:false})
    handler:=project.Resolver{Store:store}.URLMiddleware(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        identity,ok:=project.FromContext(r.Context())
        if !ok||identity.ID!="internal-alpha"{t.Fatalf("project context %#v, %v",identity,ok)}
        w.WriteHeader(http.StatusNoContent)
    }))
    cases:=[]struct{path,header string;want int}{
        {"/alpha/rooms","",204},
        {"/alpha/rooms?project_id=beta","",400},
        {"/alpha/rooms","beta",400},
        {"/alpha/rooms?project_id=internal-alpha","",204},
        {"/missing/rooms","",404},
        {"/beta/rooms","",403},
        {"/livez","",404},
        {"/alpha/%2e%2e/rooms","",404},
        {"/alpha/%2frooms","",404},
    }
    for _,tc:=range cases{
        req:=httptest.NewRequest("GET",tc.path,nil)
        if tc.header!=""{req.Header.Set("X-Project-ID",tc.header)}
        rec:=httptest.NewRecorder()
        handler.ServeHTTP(rec,req)
        if rec.Code!=tc.want{t.Errorf("%s: got %d want %d",tc.path,rec.Code,tc.want)}
    }
}

func TestProjectSlugUniqueness(t *testing.T) {
    store:=project.NewStore()
    if err:=store.CreateWithDetails("immutable-one","Alpha","alpha",map[string]any{"region":"eu"});err!=nil{t.Fatal(err)}
    if err:=store.CreateWithDetails("immutable-two","Second","alpha",nil);err!=project.ErrDuplicate{t.Fatalf("duplicate slug: %v",err)}
    got,err:=store.ResolveNamespace("alpha")
    if err!=nil||got.ID!="immutable-one"||got.Name!="Alpha"{t.Fatalf("resolved: %#v %v",got,err)}
    if _,err:=store.ResolveNamespace("_system");err==nil{t.Fatal("reserved namespace accepted")}
}
