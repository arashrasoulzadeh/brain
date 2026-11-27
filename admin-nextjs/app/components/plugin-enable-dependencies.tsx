import type { DependencyEnablePlan } from "../lib/plugin-version-review";

export function PluginEnableDependencies({ plan }: { plan: DependencyEnablePlan }) {
  return <section aria-label="Required dependency activation">
    <h3>Required dependencies</h3>
    {plan.enable.length ? <><p>Enabling this plugin will also enable these exact required versions in dependency order:</p><ul>{plan.enable.map(key => <li key={key}><code>{key}</code> — will enable</li>)}</ul></> : <p>No additional dependencies will be enabled.</p>}
    {!!plan.alreadyEnabled.length && <><p>Already enabled; unchanged:</p><ul>{plan.alreadyEnabled.map(key => <li key={key}><code>{key}</code></li>)}</ul></>}
    {!!plan.blockers.length && <div role="alert"><p>Resolve these requirements before enabling:</p><ul>{plan.blockers.map((blocker, index) => <li key={index}>{blocker}</li>)}</ul></div>}
    <p className="muted">Only required, exact, already installed and granted versions are included. Optional dependencies, grants, other projects and promoted versions are not changed.</p>
  </section>;
}
