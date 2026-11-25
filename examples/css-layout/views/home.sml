<screen name="home" version="3">
    <box align="center">
        <grid>
            <panel title="Navigation">
                <button label="Home" action="nav.home"/>
                <button label="Settings" action="nav.settings"/>
            </panel>
            <panel title="System" class="active">
                <badge tone="success">Online</badge>
                <metric label="Requests" value="{{app.requests}}"/>
                <progress label="CPU" value="{{app.cpu}}"/>
            </panel>
        </grid>
    </box>
</screen>
