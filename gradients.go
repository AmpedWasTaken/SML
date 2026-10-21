package sml

import (
 "errors"
 "fmt"
 "strconv"
 "strings"

 "github.com/AmpedWasTaken/SML/terminal"
)

var gradientTargets=map[string]bool{"logo":true,"banner":true,"prompt":true,"tui":true,"body":true}
var gradientModes=map[string]bool{"loop":true,"linear":true,"pingpong":true}
var gradientDirections=map[string]bool{"right":true,"left":true,"down":true,"up":true}

// parseGradientProfile rejects unknown styling attributes to make editor
// mistakes visible at compile/reload time rather than silently ignored.
func parseGradientProfile(node Node) (string,terminal.GradientStyle,error) {
 var zero terminal.GradientStyle
 if node.Kind!="gradient" {return "",zero,errors.New("expected gradient element")}
 if len(node.Children)!=0 || strings.TrimSpace(node.Text)!="" {
  return "",zero,errors.New("gradient must be empty")
 }
 accepted:=map[string]bool{"target":true,"colors":true,"mode":true,"direction":true,"speed":true,"fps":true,"animated":true}
 for key:=range node.Attrs{if !accepted[key] {return "",zero,fmt.Errorf("unknown gradient attribute %q",key)}}
 target:=node.Attr("target","")
 if !gradientTargets[target]{return "",zero,fmt.Errorf("unknown gradient target %q",target)}
 tokens:=strings.Split(node.Attr("colors",""),",")
 if len(tokens)<2||len(tokens)>8{return "",zero,errors.New("gradient needs 2-8 hex colors")}
 colors:=make([]string,0,len(tokens))
 for _,color:=range tokens{
  color=strings.TrimSpace(color)
  if len(color)!=7||color[0]!='#'{return "",zero,errors.New("gradient colors must use #RRGGBB")}
  colors=append(colors,color)
 }
 rgb,err:=terminal.Palette(colors)
 if err!=nil{return "",zero,err}
 mode:=node.Attr("mode","loop")
 if !gradientModes[mode]{return "",zero,errors.New("invalid gradient mode")}
 direction:=node.Attr("direction","right")
 if !gradientDirections[direction]{return "",zero,errors.New("invalid gradient direction")}
 speed,err:=strconv.ParseFloat(node.Attr("speed","0.15"),64)
 if err!=nil||speed<0||speed>5||speed!=speed{return "",zero,errors.New("gradient speed must be between 0 and 5")}
 fps,err:=strconv.Atoi(node.Attr("fps","12"))
 if err!=nil||fps<1||fps>30{return "",zero,errors.New("gradient fps must be between 1 and 30")}
 animatedText:=node.Attr("animated","true")
 if animatedText!="true"&&animatedText!="false"{return "",zero,errors.New("gradient animated must be true or false")}
 return target,terminal.GradientStyle{
  Colors:rgb, Mode:mode,Direction:direction,Speed:speed,
  FPS:fps,Animated:animatedText=="true",
 },nil
}

// GradientStyles is the immutable per-surface theme read from
// assets/views/appearance.sml. No runtime SML interpretation occurs.
func (v View) GradientStyles() (map[string]terminal.GradientStyle,error) {
 result:=map[string]terminal.GradientStyle{}
 if v.Name!="appearance" {return result,nil}
 for _,node:=range v.Nodes{
  target,style,err:=parseGradientProfile(node)
  if err!=nil{return nil,err}
  if _,exists:=result[target];exists{return nil,fmt.Errorf("duplicate gradient target %s",target)}
  result[target]=style
 }
 return result,nil
}
