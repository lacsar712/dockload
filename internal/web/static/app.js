(function () {
  'use strict';

  const POLL_MS = 3000;

  function $(sel) { return document.querySelector(sel); }
  function $$(sel) { return document.querySelectorAll(sel); }

  async function fetchJSON(url, opts) {
    const res = await fetch(url, opts);
    const data = await res.json().catch(function () { return {}; });
    if (!res.ok) throw new Error(data.error || res.statusText);
    return data;
  }

  function badgeClass(state) {
    if (state === 'Stable' || state === 'Published') return 'badge badge-ok';
    if (state === 'Collecting') return 'badge badge-warn';
    return 'badge';
  }

  function renderCalibTable(items) {
    const tbody = $('#calib-table tbody');
    if (!items || items.length === 0) {
      tbody.innerHTML = '<tr><td colspan="4" class="empty">暂无标定记录</td></tr>';
      return;
    }
    tbody.innerHTML = items.map(function (row) {
      const c = row.calib || row.Calib || row;
      const id = row.hookID || row.HookID || '';
      return '<tr><td>' + id + '</td><td>' + c.tare + '</td><td>' + c.span + '</td><td>' + c.maxLoadKg + '</td></tr>';
    }).join('');
  }

  function renderWeighsTable(items) {
    const tbody = $('#weighs-table tbody');
    if (!items || items.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="empty">暂无称重事件</td></tr>';
      return;
    }
    tbody.innerHTML = items.map(function (w) {
      return '<tr><td>' + w.timestamp + '</td><td>' + w.hookID + '</td><td>' +
        w.netKg.toFixed(2) + '</td><td>' + w.rawCounts + '</td><td>' +
        (w.stdDev != null ? w.stdDev.toFixed(2) : '—') + '</td></tr>';
    }).join('');
  }

  function renderStatusTable(items) {
    const tbody = $('#status-table tbody');
    if (!items || items.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" class="empty">暂无活跃吊具</td></tr>';
      return;
    }
    tbody.innerHTML = items.map(function (s) {
      return '<tr><td>' + s.hookID + '</td><td><span class="' + badgeClass(s.state) + '">' +
        s.state + '</span></td><td>' + (s.stable ? '是' : '否') + '</td><td>' +
        s.meanRaw.toFixed(1) + '</td><td>' + s.stdDev.toFixed(2) + '</td></tr>';
    }).join('');
  }

  async function refreshCalib() {
    try {
      const data = await fetchJSON('/v1/calib');
      renderCalibTable(data.items || []);
    } catch (e) { console.warn('calib refresh', e); }
  }

  async function refreshWeighs() {
    try {
      const data = await fetchJSON('/v1/weighs/recent?limit=50');
      renderWeighsTable(data.items || []);
    } catch (e) { console.warn('weighs refresh', e); }
  }

  async function refreshStatus() {
    try {
      const data = await fetchJSON('/v1/hooks/status');
      renderStatusTable(data.items || []);
    } catch (e) { console.warn('status refresh', e); }
  }

  async function refreshHealth() {
    try {
      await fetchJSON('/health');
      $('#health-indicator').style.color = 'var(--ok)';
    } catch (e) {
      $('#health-indicator').style.color = 'var(--danger)';
    }
  }

  function refreshAll() {
    refreshCalib();
    refreshWeighs();
    refreshStatus();
    refreshHealth();
  }

  $('#calib-form').addEventListener('submit', async function (ev) {
    ev.preventDefault();
    const fd = new FormData(ev.target);
    const hookID = String(fd.get('hookID')).trim().toUpperCase();
    const body = {
      tare: parseInt(fd.get('tare'), 10),
      span: parseFloat(fd.get('span')),
      maxLoadKg: parseFloat(fd.get('maxLoadKg'))
    };
    try {
      await fetchJSON('/v1/calib/' + encodeURIComponent(hookID), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      $('#calib-msg').textContent = '标定已保存: ' + hookID;
      refreshCalib();
    } catch (e) {
      $('#calib-msg').textContent = '错误: ' + e.message;
      $('#calib-msg').style.color = 'var(--danger)';
    }
  });

  $('#raw-form').addEventListener('submit', async function (ev) {
    ev.preventDefault();
    const fd = new FormData(ev.target);
    const body = {
      hookID: String(fd.get('hookID')).trim().toUpperCase(),
      rawCounts: parseInt(fd.get('rawCounts'), 10),
      timestamp: new Date().toISOString()
    };
    try {
      const data = await fetchJSON('/v1/sensors/raw', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      $('#raw-result').textContent = JSON.stringify(data, null, 2);
      refreshWeighs();
      refreshStatus();
    } catch (e) {
      $('#raw-result').textContent = '错误: ' + e.message;
    }
  });

  $$('.tab').forEach(function (btn) {
    btn.addEventListener('click', function () {
      $$('.tab').forEach(function (t) { t.classList.remove('active'); });
      $$('.panel').forEach(function (p) { p.classList.remove('active'); });
      btn.classList.add('active');
      $('#panel-' + btn.dataset.panel).classList.add('active');
    });
  });

  refreshAll();
  setInterval(refreshAll, POLL_MS);
})();
