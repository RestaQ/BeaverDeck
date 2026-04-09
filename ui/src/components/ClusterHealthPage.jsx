import React from 'react';

export default function ClusterHealthPage({
  nodes,
  pods,
  workloads,
  services,
  ingresses,
  pvcs,
  pvs,
  storageClasses,
  events,
  runningPodsHealth,
  formatMilliValue,
  formatByteValue,
  formatGPURequestLabel
}) {
  return (
    <>
      <div className="health-grid">
        <div className="health-card">
          <div className="small-label">Nodes</div>
          <div className="health-value">{nodes.length}</div>
          <div className="small-hint">
            Ready: {nodes.filter((n) => (n.status || '').toLowerCase().includes('ready')).length} / NotReady: {nodes.filter((n) => !(n.status || '').toLowerCase().includes('ready')).length}
          </div>
        </div>
        <div className="health-card">
          <div className="small-label">Pods</div>
          <div className="health-value">{pods.length}</div>
          <div className="small-hint">
            Running: {pods.filter((p) => p.phase === 'Running').length} | Pending: {pods.filter((p) => p.phase === 'Pending').length} | Failed: {pods.filter((p) => p.phase === 'Failed').length}
          </div>
        </div>
        <div className="health-card">
          <div className="small-label">Workloads</div>
          <div className="health-value">{workloads.length}</div>
          <div className="small-hint">Deploy/Stateful/Daemon/Jobs/CronJobs in selected namespaces</div>
        </div>
        <div className="health-card">
          <div className="small-label">Network Objects</div>
          <div className="health-value">{services.length + ingresses.length}</div>
          <div className="small-hint">Services: {services.length} | Ingresses: {ingresses.length}</div>
        </div>
        <div className="health-card">
          <div className="small-label">Storage</div>
          <div className="health-value">{pvcs.length + pvs.length}</div>
          <div className="small-hint">PVC: {pvcs.length} | PV: {pvs.length} | SC: {storageClasses.length}</div>
        </div>
        <div className="health-card">
          <div className="small-label">Recent Events</div>
          <div className="health-value">{events.length}</div>
          <div className="small-hint">Loaded events in memory for current namespace filter</div>
        </div>
      </div>
      <div className="table-wrap cluster-health-table">
        <table>
          <thead>
            <tr>
              <th>Pod</th>
              <th>Namespace</th>
              <th>CPU (used / request / limit)</th>
              <th>Memory (used / request / limit)</th>
              <th>GPU (used / requested)</th>
            </tr>
          </thead>
          <tbody>
            {runningPodsHealth.map((pod) => (
              <tr key={`${pod.namespace}/${pod.name}`}>
                <td>{pod.name}</td>
                <td>{pod.namespace}</td>
                <td>{`${pod.metrics_available ? formatMilliValue(pod.cpu_used_milli) : '-'} / ${formatMilliValue(pod.cpu_request_milli)} / ${formatMilliValue(pod.cpu_limit_milli)}`}</td>
                <td>{`${pod.metrics_available ? formatByteValue(pod.memory_used_bytes) : '-'} / ${formatByteValue(pod.memory_request_bytes)} / ${formatByteValue(pod.memory_limit_bytes)}`}</td>
                <td>
                  <div className="metric-cell">
                    <span>{`${pod.gpu || '-'} / ${formatGPURequestLabel(pod.gpu_request_count)}`}</span>
                    {pod.gpu_request_count > 0 && !pod.gpu_metrics_available ? (
                      <span className="metric-warning" title="GPU metrics are unavailable">!</span>
                    ) : null}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
