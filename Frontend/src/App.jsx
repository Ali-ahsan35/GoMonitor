import { useMemo, useState } from "react";
import axios from "axios";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
const PIE_COLORS = ["#1f9d55", "#dc2626"];
const VIEW_MODES = {
  LATENCY: "latency",
  PROFILING: "profiling",
};

function shortFunctionName(name) {
  if (!name) return "unknown";
  const parts = name.split("/");
  return parts[parts.length - 1];
}

function formatSampleValue(value, unit) {
  const safe = Number(value || 0);
  if (unit === "nanoseconds") return `${(safe / 1_000_000).toFixed(2)} ms`;
  if (unit === "bytes") return `${safe.toLocaleString()} bytes`;
  return `${safe.toLocaleString()} ${unit || "samples"}`;
}

function App() {
  const [activeView, setActiveView] = useState(VIEW_MODES.LATENCY);

  const [urls, setUrls] = useState(["https://jsonplaceholder.typicode.com/todos/1"]);
  const [apiKey, setApiKey] = useState("");
  const [requestedWith, setRequestedWith] = useState("XMLHttpRequest");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState(null);

  const [profileType, setProfileType] = useState("cpu");
  const [profileSeconds, setProfileSeconds] = useState(10);
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileError, setProfileError] = useState("");
  const [profileResult, setProfileResult] = useState(null);

  const summaryCards = useMemo(() => {
    if (!result?.summary) return [];

    const { summary } = result;
    return [
      { label: "Total Execution", value: `${summary.total_time} ms` },
      { label: "Success", value: summary.success_count },
      { label: "Failure", value: summary.failure_count },
      { label: "Goroutines", value: summary.goroutines },
      { label: "Memory Alloc", value: `${summary.memory_alloc.toLocaleString()} bytes` },
    ];
  }, [result]);

  const pieData = useMemo(() => {
    if (!result?.summary) return [];
    return [
      { name: "Success", value: result.summary.success_count },
      { name: "Failure", value: result.summary.failure_count },
    ];
  }, [result]);

  const barData = useMemo(() => {
    if (!result?.results) return [];
    return result.results.map((item, index) => ({
      name: `API ${index + 1}`,
      url: item.url,
      time: item.time,
      status: item.status,
      success: item.success,
    }));
  }, [result]);

  const topFunctionData = useMemo(() => {
    if (!profileResult?.top_functions?.length) return [];
    return profileResult.top_functions.map((entry) => ({
      name: shortFunctionName(entry.name),
      fullName: entry.name,
      value: entry.value,
    }));
  }, [profileResult]);

  const goroutineStateData = useMemo(() => {
    if (!profileResult?.goroutine_states?.length) return [];
    return profileResult.goroutine_states.map((entry) => ({
      name: entry.state,
      value: entry.count,
    }));
  }, [profileResult]);

  const updateUrl = (index, value) => {
    setUrls((prev) => prev.map((url, i) => (i === index ? value : url)));
  };

  const addUrl = () => setUrls((prev) => [...prev, ""]);

  const removeUrl = (index) => {
    setUrls((prev) => {
      if (prev.length === 1) return prev;
      return prev.filter((_, i) => i !== index);
    });
  };

  const runTest = async (event) => {
    event.preventDefault();
    setError("");

    const payloadUrls = urls.map((u) => u.trim()).filter(Boolean);
    if (payloadUrls.length === 0) {
      setError("Add at least one valid URL.");
      return;
    }

    const headers = {};
    if (apiKey.trim()) {
      headers["x-api-key"] = apiKey.trim();
    }
    if (requestedWith.trim()) {
      headers["X-Requested-With"] = requestedWith.trim();
    }

    setLoading(true);
    try {
      const { data } = await axios.post(`${API_BASE_URL}/run-test`, {
        urls: payloadUrls,
        headers,
      });
      setResult(data);
    } catch (err) {
      setResult(null);
      setError(err.response?.data?.error || "Failed to run test.");
    } finally {
      setLoading(false);
    }
  };

  const runProfileCapture = async (event) => {
    event.preventDefault();
    setProfileError("");
    setProfileLoading(true);

    try {
      const payload = {
        type: profileType,
      };
      if (profileType === "cpu") {
        payload.seconds = Number(profileSeconds) || 10;
      }

      const { data } = await axios.post(`${API_BASE_URL}/profiles/capture`, payload);
      setProfileResult(data);
    } catch (err) {
      setProfileResult(null);
      setProfileError(err.response?.data?.error || "Failed to capture profile.");
    } finally {
      setProfileLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <div className="aurora" aria-hidden="true" />
      <header className="hero">
        <p className="kicker">API Performance Dashboard</p>
        <h1>GoMonitor</h1>
        <p className="lead">
          Fire concurrent API tests from a Go backend and inspect latency, reliability, memory, and runtime behavior in one view.
        </p>
        <div className="view-toggle">
          <button
            type="button"
            className={activeView === VIEW_MODES.LATENCY ? "active-tab" : "secondary"}
            onClick={() => setActiveView(VIEW_MODES.LATENCY)}
          >
            Latency Dashboard
          </button>
          <button
            type="button"
            className={activeView === VIEW_MODES.PROFILING ? "active-tab" : "secondary"}
            onClick={() => setActiveView(VIEW_MODES.PROFILING)}
          >
            Profiling Dashboard
          </button>
        </div>
      </header>

      {activeView === VIEW_MODES.LATENCY ? <main className="content-grid">
        <section className="panel form-panel">
          <h2>Run a Test</h2>
          <form onSubmit={runTest}>
            {urls.map((url, index) => (
              <div key={index} className="url-row">
                <input
                  type="url"
                  placeholder="https://example.com/api"
                  value={url}
                  onChange={(e) => updateUrl(index, e.target.value)}
                  required
                />
                <button type="button" className="secondary" onClick={() => removeUrl(index)}>
                  Remove
                </button>
              </div>
            ))}
            <div className="header-grid">
              <input
                type="text"
                placeholder="x-api-key"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
              />
              <input
                type="text"
                placeholder="X-Requested-With"
                value={requestedWith}
                onChange={(e) => setRequestedWith(e.target.value)}
              />
            </div>
            <div className="actions">
              <button type="button" className="secondary" onClick={addUrl}>
                Add URL
              </button>
              <button type="submit" disabled={loading}>
                {loading ? "Running..." : "Run Test"}
              </button>
            </div>
          </form>
          {error && <p className="error-text">{error}</p>}
        </section>

        <section className="panel summary-panel">
          <h2>Summary</h2>
          {summaryCards.length > 0 ? (
            <div className="summary-cards">
              {summaryCards.map((card) => (
                <article key={card.label} className="stat-card">
                  <p>{card.label}</p>
                  <strong>{card.value}</strong>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">Run a test to see execution statistics.</p>
          )}
        </section>

        <section className="panel table-panel">
          <h2>Results</h2>
          {result?.results?.length ? (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>URL</th>
                    <th>Time (ms)</th>
                    <th>Status</th>
                    <th>Success</th>
                  </tr>
                </thead>
                <tbody>
                  {result.results.map((item, index) => (
                    <tr key={`${item.url}-${index}`}>
                      <td>{item.url}</td>
                      <td>{item.time}</td>
                      <td>{item.status}</td>
                      <td>{item.success ? "Yes" : "No"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="muted">No results yet.</p>
          )}
        </section>

        <section className="panel charts-panel">
          <h2>Charts</h2>
          {result?.results?.length ? (
            <div className="charts-grid">
              <div className="chart-card">
                <h3>Response Time by API</h3>
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={barData} margin={{ top: 16, right: 16, left: 0, bottom: 32 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#36314a" />
                    <XAxis dataKey="name" angle={-15} textAnchor="end" interval={0} height={50} stroke="#b8adc9" />
                    <YAxis stroke="#b8adc9" />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="time" name="Time (ms)" fill="#f59e0b" radius={[8, 8, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>

              <div className="chart-card">
                <h3>Success vs Failure</h3>
                <ResponsiveContainer width="100%" height={260}>
                  <PieChart>
                    <Pie data={pieData} dataKey="value" nameKey="name" outerRadius={95} label>
                      {pieData.map((entry, index) => (
                        <Cell key={`${entry.name}-${index}`} fill={PIE_COLORS[index % PIE_COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </div>
          ) : (
            <p className="muted">Charts appear after a successful run.</p>
          )}
        </section>
      </main> : <main className="content-grid">
        <section className="panel form-panel">
          <h2>Capture Profile</h2>
          <form onSubmit={runProfileCapture}>
            <div className="header-grid">
              <select value={profileType} onChange={(e) => setProfileType(e.target.value)}>
                <option value="cpu">CPU</option>
                <option value="heap">Heap (In-Use)</option>
                <option value="allocs">Heap (Alloc Space)</option>
                <option value="goroutine">Goroutine</option>
                <option value="mutex">Mutex</option>
                <option value="block">Blocking</option>
                <option value="threadcreate">Thread Create</option>
              </select>
              <input
                type="number"
                min="1"
                max="60"
                value={profileSeconds}
                onChange={(e) => setProfileSeconds(e.target.value)}
                disabled={profileType !== "cpu"}
                placeholder="CPU seconds"
              />
            </div>
            <div className="actions">
              <button type="submit" disabled={profileLoading}>
                {profileLoading ? "Capturing..." : "Capture Profile"}
              </button>
              {profileResult?.download_url && (
                <a
                  className="download-link"
                  href={`${API_BASE_URL}${profileResult.download_url}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  Download Raw .pprof
                </a>
              )}
            </div>
          </form>
          {profileError && <p className="error-text">{profileError}</p>}
        </section>

        <section className="panel summary-panel">
          <h2>Profile Summary</h2>
          {profileResult ? (
            <div className="summary-cards">
              <article className="stat-card">
                <p>Profile Type</p>
                <strong>{profileResult.profile_type}</strong>
              </article>
              <article className="stat-card">
                <p>Sample Type</p>
                <strong>{profileResult.sample_type || "n/a"}</strong>
              </article>
              <article className="stat-card">
                <p>Total Samples</p>
                <strong>{formatSampleValue(profileResult.total_samples, profileResult.sample_unit)}</strong>
              </article>
              <article className="stat-card">
                <p>Capture Window</p>
                <strong>{profileResult.duration_seconds || 0}s</strong>
              </article>
            </div>
          ) : (
            <p className="muted">Capture a profile to inspect runtime hotspots.</p>
          )}
          {profileResult?.notes?.length > 0 && (
            <ul className="note-list">
              {profileResult.notes.map((note, index) => (
                <li key={`${note}-${index}`}>{note}</li>
              ))}
            </ul>
          )}
        </section>

        <section className="panel table-panel">
          <h2>Top Functions</h2>
          {profileResult?.top_functions?.length ? (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Function</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  {profileResult.top_functions.map((item, index) => (
                    <tr key={`${item.name}-${index}`}>
                      <td>{item.name}</td>
                      <td>{formatSampleValue(item.value, profileResult.sample_unit)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="muted">No function samples yet.</p>
          )}
        </section>

        <section className="panel charts-panel">
          <h2>Profile Visuals</h2>
          {profileResult?.top_functions?.length ? (
            <div className="charts-grid">
              <div className="chart-card">
                <h3>Top Hotspots</h3>
                <ResponsiveContainer width="100%" height={320}>
                  <BarChart data={topFunctionData} layout="vertical" margin={{ top: 16, right: 16, left: 80, bottom: 16 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#36314a" />
                    <XAxis type="number" stroke="#b8adc9" />
                    <YAxis dataKey="name" type="category" width={180} stroke="#b8adc9" />
                    <Tooltip formatter={(value) => formatSampleValue(value, profileResult.sample_unit)} />
                    <Bar dataKey="value" name="Profile Value" fill="#34d399" radius={[0, 8, 8, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>

              <div className="chart-card">
                <h3>Goroutine States</h3>
                {goroutineStateData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={320}>
                    <PieChart>
                      <Pie data={goroutineStateData} dataKey="value" nameKey="name" outerRadius={110} label>
                        {goroutineStateData.map((entry, index) => (
                          <Cell key={`${entry.name}-${index}`} fill={PIE_COLORS[index % PIE_COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                ) : (
                  <p className="muted">Goroutine state data appears for goroutine captures.</p>
                )}
              </div>
            </div>
          ) : (
            <p className="muted">Capture a profile to render visual analysis.</p>
          )}
        </section>
      </main>}
    </div>
  );
}

export default App;
