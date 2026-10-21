package sml

import (
 "strings"
 "testing"
)

func TestSMLAppearanceGradientProfiles(t *testing.T) {
 content:=[]byte(`<screen name="appearance">
 <gradient target="logo" colors="#ff0000,#00ff00,#0000ff" mode="pingpong" direction="down" speed="0.32" fps="20" animated="true"/>
 <gradient target="prompt" colors="#ffffff,#000000" mode="linear" direction="left" speed="0" fps="12" animated="false"/>
 </screen>`)
 view,err:=Parse(content)
 if err!=nil{t.Fatal(err)}
 styles,err:=view.GradientStyles()
 if err!=nil{t.Fatal(err)}
 if len(styles)!=2||len(styles["logo"].Colors)!=3||styles["logo"].FPS!=20||styles["logo"].Mode!="pingpong"||styles["logo"].Direction!="down"||styles["prompt"].Animated {
  t.Fatalf("invalid appearance profiles: %+v",styles)
 }
 if text:=strings.Join(view.Render(nil),"");text!="" {t.Fatalf("metadata leaked into screen: %q",text)}
}

func TestSMLGradientStrictValidation(t *testing.T) {
 cases:=[]string{
  `<screen name="appearance"><gradient target="logo" colors="#ffffff" /></screen>`,
  `<screen name="appearance"><gradient target="unknown" colors="#ffffff,#000000"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" speed="-1"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" fps="200"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" animated="sometimes"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" direction="diagonal"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" mode="rainbow"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000" unexpected="test"/></screen>`,
  `<screen name="appearance"><gradient target="tui" colors="#ffffff,#000000"/><gradient target="tui" colors="#111111,#222222"/></screen>`,
  `<screen name="appearance"><text>Unexpected non-metadata</text></screen>`,
  `<screen name="home"><gradient target="tui" colors="#ffffff,#000000"/></screen>`,
 }
 for _,sample:=range cases{
  if _,err:=Parse([]byte(sample));err==nil {t.Errorf("accepted invalid SML %q",sample)}
 }
}
