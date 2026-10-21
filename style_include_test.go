package sml

import (
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func TestInlineStylingInReusableComponent(t *testing.T){
 dir:=t.TempDir()
 for _,folder:=range []string{"views","components"} {
  if err:=os.MkdirAll(filepath.Join(dir,folder),0700);err!=nil{t.Fatal(err)}
 }
 component:=`<component name="header"><badge tone="success" variant="outline">Ready</badge></component>`
 screen:=`<screen name="home" version="3"><include name="header"/></screen>`
 if err:=os.WriteFile(filepath.Join(dir,"components","header.sml"),[]byte(component),0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(dir,"views","home.sml"),[]byte(screen),0600);err!=nil{t.Fatal(err)}
 loaded,err:=LoadDir(dir)
 if err!=nil{t.Fatal(err)}
 result:=strings.Join(loaded["home"].RenderWidth(nil,80),"\n")
 if !strings.Contains(result,"[OK] [ Ready ]"){t.Fatalf("component inline styling not applied: %q",result)}
}

func TestInvalidInlineStylingInComponentRejected(t *testing.T){
 dir:=t.TempDir()
 for _,folder:=range []string{"views","components"} {
  if err:=os.MkdirAll(filepath.Join(dir,folder),0700);err!=nil{t.Fatal(err)}
 }
 if err:=os.WriteFile(filepath.Join(dir,"components","header.sml"),[]byte(`<component name="header"><text padding="100">Invalid</text></component>`),0600);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(dir,"views","home.sml"),[]byte(`<screen name="home" version="3"><include name="header"/></screen>`),0600);err!=nil{t.Fatal(err)}
 if _,err:=LoadDir(dir);err==nil{t.Fatal("invalid inline style in included component accepted")}
}
