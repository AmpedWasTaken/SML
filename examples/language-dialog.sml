<screen name="language-dialog" version="3">
  <box width="54" align="center">
    <panel border="round" padding="1" background="#131b2c" foreground="#eaf2fb" border-gradient="#5cddea,#9781fd,#f77dc1,#5cddea">
      <text foreground-gradient="#5cddea,#9781fd,#f77dc1">◇  SELECT LANGUAGE</text>
      <text foreground="#8b99b2">Choose your preferred interface language</text>
      <divider/>
      <if field="language.selected" equals="en">
        <text background="#253e59" foreground="#81e2f6">› English [en]</text>
        <otherwise><text>  English [en]</text></otherwise>
      </if>
      <if field="language.selected" equals="zh">
        <text background="#253e59" foreground="#81e2f6">› 中文 (Chinese) [zh]</text>
        <otherwise><text>  中文 (Chinese) [zh]</text></otherwise>
      </if>
      <if field="language.selected" equals="ru">
        <text background="#253e59" foreground="#81e2f6">› Русский (Russian) [ru]</text>
        <otherwise><text>  Русский (Russian) [ru]</text></otherwise>
      </if>
      <spacer/>
      <text foreground="#8b99b2">Up/Down navigate · Enter confirm · 1-3 select</text>
    </panel>
  </box>
</screen>
