<screen name="dashboard" version="3">
  <derive key="derived.health" op="coalesce" source="service.health" fallback="unknown"/>
  <section title="Service Overview">
    <field label="Health" key="derived.health"/>
    <grid columns="2" collapse-below="60" gap="2">
      <metric label="Requests" value="{{stats.requests}}"/>
      <metric label="Users" value="{{stats.users}}"/>
    </grid>
    <progress label="Capacity" value="{{stats.capacity}}"/>
    <if field="service.health" equals="online">
      <badge>Online</badge>
      <otherwise><alert>Service not online</alert></otherwise>
    </if>
  </section>
</screen>
