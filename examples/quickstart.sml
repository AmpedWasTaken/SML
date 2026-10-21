<screen name="quickstart" version="3">
  <derive key="derived.health" op="coalesce"
          source="app.health" fallback="unknown"/>
  <section title="APPLICATION DASHBOARD">
    <grid columns="2" gap="2" collapse-below="60">
      <metric label="Requests" value="{{app.requests}}"/>
      <metric label="Errors" value="{{app.errors}}"/>
    </grid>
    <panel title="Service">
      <field label="Status" key="derived.health"/>
      <progress label="CPU" value="{{app.cpu}}"/>
    </panel>
  </section>
</screen>
