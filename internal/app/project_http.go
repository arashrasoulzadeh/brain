package app

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "strings"

    "brain/internal/admin"
    "brain/internal/operator"
    "brain/internal/plugin"
    "brain/internal/project"
)

func projectIdentity(r *http.Request) (project.Project,bool) {
    p,ok:=project.FromContext(r.Context())
    return p,ok && p.ID!="" && p.Active
}

func checkedIdentity(p project.Project, values map[string]any) bool {
    for _,key:=range []string{"project_id","projectId","project"}{
        value,exists:=values[key]
        if !exists{continue}
        text,ok:=value.(string)
        if !ok || (text!=p.ID && text!=p.Slug){return false}
    }
    return true
}

// ProjectCommandHandler binds the command's project ID to host-resolved context
// before an existing operator or functions handler sees request JSON.
func ProjectCommandHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        p,ok:=projectIdentity(r)
        if !ok{http.NotFound(w,r);return}
        data,err:=io.ReadAll(io.LimitReader(r.Body,1<<20+1))
        if err!=nil||len(data)>1<<20{
            http.Error(w,"invalid request body",http.StatusBadRequest);return
        }
        var command map[string]any
        if err:=json.Unmarshal(data,&command);err!=nil||command==nil||!checkedIdentity(p,command){
            http.Error(w,"invalid project command",http.StatusBadRequest);return
        }
        command["project_id"]=p.ID
        bound,err:=json.Marshal(command)
        if err!=nil{http.Error(w,"invalid request",http.StatusBadRequest);return}
        r.Body=io.NopCloser(bytes.NewReader(bound))
        r.ContentLength=int64(len(bound))
        next.ServeHTTP(w,r)
    })
}

// ProjectQueryHandler removes any ambiguity between the URL namespace and
// legacy query fields consumed internally by the operator services.
func ProjectQueryHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        p,ok:=projectIdentity(r)
        if !ok{http.NotFound(w,r);return}
        clone:=r.Clone(r.Context())
        query:=clone.URL.Query()
        query.Set("project_id",p.ID)
        clone.URL.RawQuery=query.Encode()
        next.ServeHTTP(w,clone)
    })
}

func AuthenticatedProjectRoutesHandler(service *admin.LifecycleService,sessions *operator.Store) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        p,ok:=projectIdentity(r)
        if !ok{http.NotFound(w,r);return}
        if _,ok:=operatorActorFromRequest(sessions,r);!ok{
            http.Error(w,"forbidden",http.StatusForbidden);return
        }
        w.Header().Set("Content-Type","application/json")
        _=json.NewEncoder(w).Encode(service.ProjectHTTPRoutes(p.ID))
    })
}

func PluginHTTPHandler(service *admin.LifecycleService,sessions *operator.Store) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        p,ok:=projectIdentity(r)
        if !ok{http.NotFound(w,r);return}
        _,relative,ok:=strings.Cut(strings.TrimPrefix(r.URL.Path,"/"),"/")
        if !ok{http.NotFound(w,r);return}
        route,found:=service.MatchHTTPRoute(p.ID,r.Method,"/"+relative)
        if !found{http.NotFound(w,r);return}
        actorID:=""
        if route.Auth=="operator" {
            actor,authenticated:=operatorActorFromRequest(sessions,r)
            if !authenticated{http.Error(w,"forbidden",http.StatusForbidden);return}
            actorID=actor.ID
        }else if route.Auth!="public" {
            http.Error(w,"forbidden",http.StatusForbidden);return
        }
        fields:=make(map[string]any)
        for key,values:=range r.URL.Query(){
            if len(values)!=1{http.Error(w,"invalid query",http.StatusBadRequest);return}
            fields[key]=values[0]
        }
        if r.Method!="GET" && r.Method!="HEAD" {
            input,err:=io.ReadAll(io.LimitReader(r.Body,1<<20+1))
            if err!=nil||len(input)>1<<20{
                http.Error(w,"invalid request body",http.StatusBadRequest);return
            }
            if len(bytes.TrimSpace(input))>0 {
                var body map[string]any
                if json.Unmarshal(input,&body)!=nil||body==nil{
                    http.Error(w,"invalid JSON object",http.StatusBadRequest);return
                }
                for key,value:=range body{
                    if _,duplicate:=fields[key];duplicate{
                        http.Error(w,"conflicting request fields",http.StatusBadRequest);return
                    }
                    fields[key]=value
                }
            }
        }
        for key,value:=range route.Params{
            if supplied,exists:=fields[key];exists&&supplied!=value{
                http.Error(w,"conflicting route parameter",http.StatusBadRequest);return
            }
            fields[key]=value
        }
        if !checkedIdentity(p,fields){
            http.Error(w,"conflicting project identity",http.StatusBadRequest);return
        }
        // Never import identity or grants from JSON into the worker request.
        delete(fields,"project_id")
        delete(fields,"projectId")
        delete(fields,"project")
        payload,err:=json.Marshal(fields)
        if err!=nil{http.Error(w,"invalid payload",http.StatusBadRequest);return}
        response,err:=service.InvokeHTTPRoute(p.ID,route,payload,actorID)
        if err!=nil{
            status:=http.StatusForbidden
            if errors.Is(err,plugin.ErrNotInstalled)||errors.Is(err,plugin.ErrDisabled)||errors.Is(err,plugin.ErrNotReady)||errors.Is(err,plugin.ErrNotRegistered)||errors.Is(err,plugin.ErrProjectInvalid){
                status=http.StatusNotFound
            }
            http.Error(w,http.StatusText(status),status);return
        }
        if !response.Success||response.ErrorCode!=""{
            http.Error(w,"plugin invocation rejected",http.StatusBadRequest);return
        }
        if len(response.Payload)>0&&!json.Valid(response.Payload){
            http.Error(w,"invalid plugin response",http.StatusBadGateway);return
        }
        w.Header().Set("Content-Type","application/json")
        if len(response.Payload)==0{_,_=w.Write([]byte("{}"));return}
        _,_=w.Write(response.Payload)
    })
}
