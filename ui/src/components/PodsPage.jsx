import React from 'react';
import ActionMenu from './ActionMenu.jsx';

export default function PodsPage({
  podSearch,
  setPodSearch,
  podsAutoRefreshEnabled,
  setPodsAutoRefreshEnabled,
  podsAutoRefreshSeconds,
  setPodsAutoRefreshSeconds,
  sortedPods,
  toggleSort,
  sortMark,
  selectedPod,
  selectPod,
  isDegradedReady,
  showWarningPopover,
  scheduleWarningPopoverHide,
  makeAction,
  permissionInfo,
  safe,
  openManifestTab,
  openPodLogsTab,
  evictPodByRef,
  setStatus,
  refreshAll,
  deletePodByRef,
  setSelectedPod,
  allAllowed,
  openPodExecTab
}) {
  return (
    <>
      <div className="toolbar fixed-toolbar">
        <input value={podSearch} onChange={(e) => setPodSearch(e.target.value)} placeholder="Search pods..." />
        <label className="toggle-row">
          <input
            type="checkbox"
            checked={podsAutoRefreshEnabled}
            onChange={(e) => setPodsAutoRefreshEnabled(e.target.checked)}
          />
          <span>Auto refresh</span>
        </label>
        <select
          value={String(podsAutoRefreshSeconds)}
          onChange={(e) => setPodsAutoRefreshSeconds(Number(e.target.value))}
          disabled={!podsAutoRefreshEnabled}
        >
          <option value="1">1s</option>
          <option value="5">5s</option>
          <option value="15">15s</option>
        </select>
      </div>

      <div className="table-wrap pods-table-wrap">
        <table>
          <thead>
            <tr>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'name')}>Name {sortMark('pods', 'name')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'namespace')}>Namespace {sortMark('pods', 'namespace')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'phase')}>Status {sortMark('pods', 'phase')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'ready')}>Ready {sortMark('pods', 'ready')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'restarts')}>Restarts {sortMark('pods', 'restarts')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'node')}>Node {sortMark('pods', 'node')}</button></th>
              <th><button className="sort-btn" onClick={() => toggleSort('pods', 'age')}>Age {sortMark('pods', 'age')}</button></th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {sortedPods.map((p) => (
              <tr
                key={`${p.namespace}/${p.name}`}
                className={selectedPod?.namespace === p.namespace && selectedPod?.name === p.name ? 'active-row' : ''}
                onClick={() => selectPod(p)}
              >
                <td>{p.name}</td>
                <td>{p.namespace}</td>
                <td>{p.phase}</td>
                <td>
                  <div className="ready-cell">
                    <span>{p.ready}</span>
                    {isDegradedReady(p.ready) ? (
                      <button
                        className="warning-indicator"
                        title="Pod is not fully ready"
                        onMouseEnter={(e) => {
                          void showWarningPopover(e, { type: 'pod', key: `pod:${p.namespace}:${p.name}`, item: p });
                        }}
                        onMouseLeave={() => scheduleWarningPopoverHide(`pod:${p.namespace}:${p.name}`)}
                        onClick={(e) => e.stopPropagation()}
                      >
                        !
                      </button>
                    ) : null}
                  </div>
                </td>
                <td>{p.restarts}</td>
                <td>{p.node || '-'}</td>
                <td>{p.age}</td>
                <td className="actions-cell">
                  <ActionMenu
                    actions={[
                      makeAction('Manifest', permissionInfo('pods', 'view', p.namespace), () => safe(() => openManifestTab(p.namespace, 'pod', p.name))),
                      makeAction('Logs', permissionInfo('pods', 'view', p.namespace), () => safe(() => openPodLogsTab(p.namespace, p.name))),
                      makeAction(
                        'Evict',
                        permissionInfo('pods', 'edit', p.namespace),
                        () => safe(async () => {
                          await evictPodByRef(p.namespace, p.name);
                          setStatus(`Eviction requested for ${p.namespace}/${p.name}`);
                          await refreshAll();
                        })
                      ),
                      makeAction(
                        'Delete',
                        permissionInfo('pods', 'delete', p.namespace),
                        () => safe(async () => {
                          await deletePodByRef(p.namespace, p.name);
                          if (selectedPod?.namespace === p.namespace && selectedPod?.name === p.name) {
                            setSelectedPod(null);
                          }
                          await refreshAll();
                        })
                      ),
                      makeAction(
                        'Exec',
                        allAllowed(
                          permissionInfo('pods', 'view', p.namespace),
                          permissionInfo('exec', 'edit', p.namespace),
                          String(p.phase) === 'Running' && !isDegradedReady(p.ready)
                            ? { allowed: true, reason: '' }
                            : { allowed: false, reason: 'Exec is disabled for pods that are not fully running' }
                        ),
                        () => openPodExecTab(p.namespace, p.name)
                      )
                    ]}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}
