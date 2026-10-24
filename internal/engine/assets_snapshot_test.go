package engine

import(
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestLoadedSMLIncludeSnapshotSurvivesOnDiskMutation(t *testing.T){
 dir:=t.TempDir()
 for _,child:=range []string{"views","components"}{
  if err:=os.MkdirAll(filepath.Join(dir,child),0700);err!=nil{t.Fatal(err)}
 }
 write:=func(path,data string){t.Helper();if err:=os.WriteFile(path,[]byte(data),0600);err!=nil{t.Fatal(err)}}
 viewPath:=filepath.Join(dir,"views","home.sml")
 compPath:=filepath.Join(dir,"components","header.sml")
 write(viewPath,`<screen name="home" version="3"><include name="header"/></screen>`)
 write(compPath,`<component name="header"><badge tone="success">ALLOWED</badge></component>`)
 loaded,err:=LoadDir(dir)
 if err!=nil{t.Fatal(err)}
 original:=strings.Join(loaded["home"].RenderWidth(nil,60)," ")
 if !strings.Contains(original,"ALLOWED"){t.Fatalf("missing component: %q",original)}

 // Unvalidated component edits may happen between two explicit reloads.
 write(compPath,`<component name="header"><button label="Unsafe" command="sh"/></component>`)
 kept,err:=loaded["home"].RenderChecked(nil,60)
 if err!=nil{t.Fatalf("old validated view must stay usable: %v",err)}
 if strings.Join(kept," ")!=original{t.Fatalf("component snapshot mutated without reload: %v",kept)}
 if _,err:=LoadDir(dir);err==nil{t.Fatal("invalid edited component accepted on reload")}

 write(compPath,`<component name="header"><text>RELOADED</text></component>`)
 updated,err:=LoadDir(dir)
 if err!=nil{t.Fatal(err)}
 if !strings.Contains(strings.Join(updated["home"].RenderWidth(nil,60)," "),"RELOADED"){
  t.Fatal("component reload did not pick up valid changes")
 }
 if strings.Join(loaded["home"].RenderWidth(nil,60)," ")!=original{
  t.Fatal("previously published SML snapshot was modified")
 }
}

func TestLoadedSMLUseSnapshotIsStable(t *testing.T){
 dir:=t.TempDir()
 for _,child:=range []string{"views","components"}{
  if err:=os.MkdirAll(filepath.Join(dir,child),0700);err!=nil{t.Fatal(err)}
 }
 component:=filepath.Join(dir,"components","notice.sml")
 if err:=os.WriteFile(filepath.Join(dir,"views","home.sml"),[]byte(`<screen name="home" version="3"><use name="notice" message="SAFE"/></screen>`),0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(component,[]byte(`<component name="notice" props="message"><text>{{props.message}}</text></component>`),0600);err!=nil{t.Fatal(err)}
 views,err:=LoadDir(dir)
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(component,[]byte(`<component name="notice" props="message"><text>UNREVIEWED</text></component>`),0600);err!=nil{t.Fatal(err)}
 rows,err:=views["home"].RenderChecked(nil,60)
 if err!=nil{t.Fatal(err)}
 if !strings.Contains(strings.Join(rows," "),"SAFE") || strings.Contains(strings.Join(rows," "),"UNREVIEWED"){
  t.Fatalf("use read modified on-disk component: %v",rows)
 }
}
