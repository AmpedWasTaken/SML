package sml

import (
 "errors"
 "fmt"
 "regexp"
 "strconv"
 "strings"
 "unicode/utf8"
)

var inputDefault=regexp.MustCompile(`^[ -~]{0,64}$`)

func validateSessionInput(n Node) error {
 if n.Kind!="input" && n.Kind!="number" {return nil}
 if len(n.Children)>0 || strings.TrimSpace(n.Text)!="" {return errors.New("SML input cannot contain children")}
 for name:=range n.Attrs {
  switch name {
  case "label","key","placeholder","min","max":
  default: return fmt.Errorf("unsupported SML input attribute %q",name)
  }
 }
 if !uiLabel.MatchString(n.Attr("label","")) || !sessionKey.MatchString(n.Attr("key","")) {
  return errors.New("invalid input label or local state key")
 }
 if !inputDefault.MatchString(n.Attr("placeholder","")) {return errors.New("unsafe SML input placeholder")}
 if n.Kind=="input" {
  if _,ok:=n.Attrs["min"];ok{return errors.New("min is only supported on number")}
  if _,ok:=n.Attrs["max"];ok{return errors.New("max is only supported on number")}
 } else {
  lo,err:=strconv.Atoi(n.Attr("min","0"));if err!=nil||lo<0||lo>1000000{return errors.New("invalid number min")}
  hi,err:=strconv.Atoi(n.Attr("max","100"));if err!=nil||hi<lo||hi>1000000{return errors.New("invalid number max")}
 }
 return nil
}
func renderSessionInput(n Node,values map[string]string,width int)string{
 v:=values[n.Attr("key","")]
 if v==""{v=n.Attr("placeholder","")}
 if v=="" {v="(empty)"}
 if utf8.RuneCountInString(v)>64 {v=string([]rune(v)[:64])}
 return trim("[ "+n.Attr("label","")+": "+safe(v)+" ]",width)
}
