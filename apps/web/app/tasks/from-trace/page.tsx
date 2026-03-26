import TaskFromTraceForm from "../../../components/TaskFromTraceForm";

export default function TaskBuilderPage() {
  return (
    <main className="shell">
      <section className="toprail" aria-label="Tasks navigation">
        <a className="toprail-link" href="/">
          Dashboard
        </a>
        <a className="toprail-link" href="/tasks">
          Tasks
        </a>
        <a className="toprail-link" href="/tasks/from-trace">
          Task Builder
        </a>
      </section>

      <section className="hero-page hero-compact">
        <p className="eyebrow">Task Builder</p>
        <h1>Build Task from Trace</h1>
        <p className="lede">
          Generate a reusable evaluation task from an existing execution trace.
          The system extracts input parameters and expected outputs automatically.
        </p>
      </section>

      <div className="content-grid">
        <article className="panel">
          <div className="panel-header">
            <h2>Task Configuration</h2>
          </div>

          <TaskFromTraceForm />
        </article>
      </div>
    </main>
  );
}
