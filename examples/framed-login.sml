<screen name="framed-login" version="3">
  <panel title="{{login.brand}} / LOGIN" border="round" padding="1" border-gradient="#5cddea,#9781fd,#f77dc1,#5cddea">
    <text foreground="#76d8d7">{{login.heading}}</text>
    <text foreground="#71717a">LANGUAGE / {{login.language}} · Ctrl+L / F2</text>
    <divider/>
    <text align="center" foreground-gradient="#5cddea,#9781fd,#f77dc1">{{login.hero}}</text>
    <text align="center" foreground="#a1a1aa">{{login.subtitle}}</text>
    <text align="center" foreground="#52525b">{{login.clock}}</text>
    <spacer/>
    <box width="40" align="center">
      <if field="login.focus" equals="username">
        <text align="center" foreground="#76d8d7">{{login.username_label}}</text>
        <input id="login-username" key="state.login_username" label="Username" appearance="boxed" border="round" border-color="#76d8d7"/>
        <otherwise>
          <text align="center" foreground="#71717a">{{login.username_label}}</text>
          <input id="login-username" key="state.login_username" label="Username" appearance="boxed" border="round" border-color="#71717a"/>
        </otherwise>
      </if>
      <if field="login.focus" equals="password">
        <text align="center" foreground="#76d8d7">{{login.password_label}}</text>
        <input id="login-password" key="state.login_password_mask" label="Password" appearance="boxed" border="round" border-color="#76d8d7"/>
        <otherwise>
          <text align="center" foreground="#71717a">{{login.password_label}}</text>
          <input id="login-password" key="state.login_password_mask" label="Password" appearance="boxed" border="round" border-color="#71717a"/>
        </otherwise>
      </if>
    </box>
    <spacer/>
    <button id="login-submit" action="login.submit" label="Sign in" align="center" foreground="#76d8d7"/>
    <text align="center" foreground="#52525b">{{login.hint}}</text>
    <divider/>
    <text foreground="#71717a">{{login.status}}</text>
  </panel>
</screen>
