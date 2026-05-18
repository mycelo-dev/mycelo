"use client";

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  getStoredApiKey,
  getJson,
  getStoredAccount,
  getStoredActiveTeamId,
  postJson,
  query,
  setStoredAccount,
  setStoredActiveTeamId,
  setStoredApiKey,
} from "../lib/api";
import type {
  AccountContext,
  ApiKeyResponse,
  DeadLetterEvent,
  DeliveryFailureEvent,
  Destination,
  EventTopic,
  EventsResponse,
  Mapping,
  OutboundMetrics,
  ReplayResult,
  SignUpResponse,
  Team,
  TopicHead,
  Topic,
} from "../lib/types";

type View = "overview" | "delivery" | "dlq" | "observability" | "topics" | "destinations" | "mappings" | "events" | "api-keys";
type Theme = "light" | "dark";

type ViewMeta = {
  id: View;
  label: string;
  kicker: string;
  summary: string;
  short: string;
};

const views: ViewMeta[] = [
  { id: "overview", label: "Overview", kicker: "Control plane", summary: "Health, setup, and recent delivery signals.", short: "OV" },
  { id: "delivery", label: "Delivery state", kicker: "Operations", summary: "Mapping health, cursors, backoff, and endpoint status.", short: "DS" },
  { id: "dlq", label: "Dead letters", kicker: "Recovery", summary: "Replay failed events by destination, topic, or event.", short: "DL" },
  { id: "observability", label: "Observability", kicker: "Telemetry", summary: "Delivery totals, latency, lag, and circuit activity.", short: "OB" },
  { id: "events", label: "Event log", kicker: "Streams", summary: "Live and historical events for each topic.", short: "EV" },
  { id: "topics", label: "Topics", kicker: "Configure", summary: "Create and inspect event topics.", short: "TP" },
  { id: "destinations", label: "Destinations", kicker: "Configure", summary: "Manage webhook destinations and signing secrets.", short: "DN" },
  { id: "mappings", label: "Mappings", kicker: "Configure", summary: "Attach topics to destinations and tune retry policy.", short: "MP" },
  { id: "api-keys", label: "Account", kicker: "Access", summary: "Tenant, teams, and request credentials.", short: "AC" },
];

const navSections: Array<{ label: string; items: View[] }> = [
  { label: "Monitor", items: ["overview", "delivery", "dlq", "observability", "events"] },
  { label: "Configure", items: ["topics", "destinations", "mappings", "api-keys"] },
];

const THEME_STORAGE_KEY = "mycelo_theme";
const AUTO_REFRESH_STORAGE_KEY = "mycelo_metrics_auto_refresh";
const AUTO_REFRESH_INTERVAL_MS = 3000;
const metricViews: View[] = ["overview", "delivery", "dlq", "observability", "events"];

const emptyMetrics: OutboundMetrics = {
  delivery_success_total: 0,
  delivery_success_last_at: 0,
  delivery_failure_total: {},
  dead_letter_write_total: 0,
  dead_letter_replay_total: 0,
  circuit_opened_total: {},
  circuit_blocked_total: {},
  delivery_lag_ms: { count: 0, total: 0, max: 0, last: 0, average: 0 },
  delivery_attempt_duration_ms: { count: 0, total: 0, max: 0, last: 0, average: 0 },
};

export function OperatorConsole() {
  const [view, setView] = useState<View>("api-keys");
  const [theme, setTheme] = useState<Theme>("light");
  const [autoRefreshMetrics, setAutoRefreshMetrics] = useState(true);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [destinations, setDestinations] = useState<Destination[]>([]);
  const [mappings, setMappings] = useState<Mapping[]>([]);
  const [topicHeads, setTopicHeads] = useState<Record<string, number>>({});
  const [dlq, setDlq] = useState<DeadLetterEvent[]>([]);
  const [deliveryFailures, setDeliveryFailures] = useState<DeliveryFailureEvent[]>([]);
  const [eventTopics, setEventTopics] = useState<string[]>([]);
  const [metrics, setMetrics] = useState<OutboundMetrics>(emptyMetrics);
  const [events, setEvents] = useState<EventsResponse>({ events: [], count: 0, cursor: 0, has_more: false });
  const [selectedTopic, setSelectedTopic] = useState("");
  const [isViewingLatestEvents, setIsViewingLatestEvents] = useState(true);
  const [dlqDestinationFilter, setDlqDestinationFilter] = useState("");
  const [dlqTopicFilter, setDlqTopicFilter] = useState("");
  const [toast, setToast] = useState("");
  const [loading, setLoading] = useState(false);
  const [apiKeyLoaded, setApiKeyLoaded] = useState(false);
  const [account, setAccount] = useState<AccountContext | null>(null);
  const [activeTeamId, setActiveTeamId] = useState("");
  const autoRefreshInFlight = useRef(false);

  useEffect(() => {
    const storedTheme = getStoredTheme();
    setTheme(storedTheme);
    applyTheme(storedTheme);
    setAutoRefreshMetrics(getStoredAutoRefresh());
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme((current) => {
      const nextTheme = current === "dark" ? "light" : "dark";
      applyTheme(nextTheme);
      return nextTheme;
    });
  }, []);

  const applyAutoRefresh = useCallback((enabled: boolean) => {
    setAutoRefreshMetrics(enabled);
    setStoredAutoRefresh(enabled);
  }, []);

  const clearScopedData = useCallback(() => {
    setTopics([]);
    setDestinations([]);
    setMappings([]);
    setTopicHeads({});
    setDlq([]);
    setDeliveryFailures([]);
    setEventTopics([]);
    setMetrics(emptyMetrics);
    setEvents({ events: [], count: 0, cursor: 0, has_more: false });
    setSelectedTopic("");
  }, []);

  useEffect(() => {
    const storedAccount = getStoredAccount();
    const storedTeamId = getStoredActiveTeamId() || storedAccount?.team_public_id || "";
    if (storedTeamId) {
      setStoredActiveTeamId(storedTeamId);
    }
    setAccount(storedAccount);
    setActiveTeamId(storedTeamId);
    setApiKeyLoaded(true);
    if (!storedAccount) {
      setView("api-keys");
      setToast("Sign up to continue");
    } else if (!storedTeamId) {
      setView("api-keys");
      setToast("Select or create a team");
    } else {
      setView("overview");
    }
  }, []);

  const applyActiveTeam = useCallback((teamPublicId: string) => {
    const trimmed = teamPublicId.trim();
    setStoredActiveTeamId(trimmed);
    setActiveTeamId(trimmed);
    if (!trimmed) {
      clearScopedData();
      setView("api-keys");
    }
  }, [clearScopedData]);

  const applyAccount = useCallback((nextAccount: AccountContext | null) => {
    setStoredAccount(nextAccount);
    setAccount(nextAccount);
    if (!nextAccount) {
      setStoredApiKey("");
      setActiveTeamId("");
      clearScopedData();
      setView("api-keys");
      return;
    }

    const nextTeamId = nextAccount.team_public_id || getStoredActiveTeamId();
    if (nextTeamId) {
      applyActiveTeam(nextTeamId);
      setView("overview");
      return;
    }

    setView("api-keys");
  }, [applyActiveTeam, clearScopedData]);

  const applyIssuedApiKey = useCallback((apiKey: string) => {
    const trimmed = apiKey.trim();
    setStoredApiKey(trimmed);
  }, []);

  const unhealthyMappings = useMemo(
    () =>
      mappings
        .filter((mapping) => mapping.last_error || mapping.consecutive_failure_count > 0 || mapping.next_attempt_at > Date.now())
        .sort((a, b) => b.consecutive_failure_count - a.consecutive_failure_count),
    [mappings],
  );

  const currentView = useMemo(() => views.find((item) => item.id === view) ?? views[0], [view]);
  const isMetricView = metricViews.includes(view);
  const toastTone = toast && ["failed", "error", "invalid", "unauthorized", "denied"].some((token) => toast.toLowerCase().includes(token)) ? "bad" : "good";

  const readTopicHeads = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      return {};
    }

    const rows = await getJson<TopicHead[]>("/console/topic_heads");
    return Object.fromEntries((rows ?? []).map((topic) => [topic.topic_id, topic.latest_event_id]));
  }, []);

  const refreshCore = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      clearScopedData();
      return;
    }

    const [topicRows, destinationRows, mappingRows, outboundMetrics, eventTopicRows, deliveryFailureRows, topicHeadRows] = await Promise.all([
      getJson<Topic[]>("/topics"),
      getJson<Destination[]>("/destinations"),
      getJson<Mapping[]>("/destination_topic_mappings"),
      getJson<OutboundMetrics>("/observability/outbound"),
      getJson<EventTopic[]>("/console/event_topics"),
      getJson<DeliveryFailureEvent[]>("/delivery_failures?limit=100"),
      readTopicHeads(),
    ]);
    setTopics(topicRows ?? []);
    setDestinations(destinationRows ?? []);
    setMappings(mappingRows ?? []);
    setTopicHeads(topicHeadRows);
    setMetrics(outboundMetrics ?? emptyMetrics);
    setDeliveryFailures(deliveryFailureRows ?? []);
    const nextEventTopics = uniqueStrings([
      ...(topicRows ?? []).map((topic) => topic.topic_name),
      ...(eventTopicRows ?? []).map((topic) => topic.topic_name),
    ]);
    setEventTopics(nextEventTopics);
    setSelectedTopic((current) => current || nextEventTopics[0] || "");
  }, [clearScopedData, readTopicHeads]);

  const refreshMappings = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      setMappings([]);
      return;
    }

    const [mappingRows, topicHeadRows] = await Promise.all([
      getJson<Mapping[]>("/destination_topic_mappings"),
      readTopicHeads(),
    ]);
    setMappings(mappingRows ?? []);
    setTopicHeads(topicHeadRows);
  }, [readTopicHeads]);

  const refreshOutboundMetrics = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      setMetrics(emptyMetrics);
      return;
    }

    const outboundMetrics = await getJson<OutboundMetrics>("/observability/outbound");
    setMetrics(outboundMetrics ?? emptyMetrics);
  }, []);

  const refreshDlq = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      setDlq([]);
      return;
    }

    const rows = await getJson<DeadLetterEvent[]>(
      `/dead_letter_events${query({
        destination_id: dlqDestinationFilter,
        topic_id: dlqTopicFilter,
        limit: 100,
      })}`,
    );
    setDlq(rows ?? []);
  }, [dlqDestinationFilter, dlqTopicFilter]);

  const refreshDeliveryFailures = useCallback(async () => {
    if (!getStoredActiveTeamId()) {
      setDeliveryFailures([]);
      return;
    }

    const rows = await getJson<DeliveryFailureEvent[]>("/delivery_failures?limit=100");
    setDeliveryFailures(rows ?? []);
  }, []);

  const refreshEvents = useCallback(async (nextCursor = 0) => {
    if (!selectedTopic || !activeTeamId) {
      setEvents({ events: [], count: 0, cursor: 0, has_more: false });
      return;
    }
    const response = await getJson<EventsResponse>(
      `/console/events${query({ topic: selectedTopic, offset: nextCursor, limit: 50, order: "desc" })}`,
    );
    setEvents(response);
  }, [activeTeamId, selectedTopic]);

  const refreshCurrentMetrics = useCallback(async () => {
    if (!activeTeamId) {
      return;
    }

    if (view === "overview") {
      await Promise.all([refreshMappings(), refreshOutboundMetrics(), refreshDlq(), refreshDeliveryFailures()]);
      return;
    }

    if (view === "delivery") {
      await Promise.all([refreshMappings(), refreshDeliveryFailures()]);
      return;
    }

    if (view === "dlq") {
      await refreshDlq();
      return;
    }

    if (view === "observability") {
      await refreshOutboundMetrics();
      return;
    }

    if (view === "events") {
      setIsViewingLatestEvents(true);
      await refreshEvents(0);
      return;
    }

    await refreshCore();
  }, [activeTeamId, refreshCore, refreshDeliveryFailures, refreshDlq, refreshEvents, refreshMappings, refreshOutboundMetrics, view]);

  const run = useCallback(async (action: () => Promise<void | false>, success: string) => {
    setLoading(true);
    setToast("");
    try {
      const result = await action();
      if (result !== false) {
        setToast(success);
      }
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!apiKeyLoaded) {
      return;
    }

    if (!activeTeamId) {
      clearScopedData();
      setToast(account ? "Select or create a team" : "Sign up to continue");
      return;
    }

    run(refreshCore, "Console refreshed");
  }, [account, activeTeamId, apiKeyLoaded, clearScopedData, refreshCore, run]);

  useEffect(() => {
    if (!apiKeyLoaded || !activeTeamId) {
      return;
    }

    refreshDlq().catch((error) => setToast(error instanceof Error ? error.message : "Failed to load DLQ"));
    refreshDeliveryFailures().catch((error) => setToast(error instanceof Error ? error.message : "Failed to load delivery failures"));
  }, [activeTeamId, apiKeyLoaded, refreshDeliveryFailures, refreshDlq]);

  useEffect(() => {
    setIsViewingLatestEvents(true);
    refreshEvents(0).catch(() => undefined);
  }, [refreshEvents]);

  useEffect(() => {
    if (view !== "events") {
      return;
    }

    setIsViewingLatestEvents(true);
    refreshEvents(0).catch((error) => setToast(error instanceof Error ? error.message : "Failed to load events"));
  }, [view, refreshEvents]);

  useEffect(() => {
    if (view !== "events" || !selectedTopic || !isViewingLatestEvents || !autoRefreshMetrics) {
      return;
    }

    const interval = window.setInterval(() => {
      if (autoRefreshInFlight.current) {
        return;
      }

      autoRefreshInFlight.current = true;
      refreshEvents(0)
        .catch((error) => setToast(error instanceof Error ? error.message : "Failed to refresh events"))
        .finally(() => {
          autoRefreshInFlight.current = false;
        });
    }, AUTO_REFRESH_INTERVAL_MS);

    return () => window.clearInterval(interval);
  }, [autoRefreshMetrics, isViewingLatestEvents, refreshEvents, selectedTopic, view]);

  useEffect(() => {
    if (!apiKeyLoaded || !activeTeamId || !autoRefreshMetrics || !isMetricView || view === "events") {
      return;
    }

    const interval = window.setInterval(() => {
      if (autoRefreshInFlight.current) {
        return;
      }

      autoRefreshInFlight.current = true;
      refreshCurrentMetrics()
        .catch((error) => setToast(error instanceof Error ? error.message : "Failed to auto refresh"))
        .finally(() => {
          autoRefreshInFlight.current = false;
        });
    }, AUTO_REFRESH_INTERVAL_MS);

    return () => window.clearInterval(interval);
  }, [activeTeamId, apiKeyLoaded, autoRefreshMetrics, isMetricView, refreshCurrentMetrics, view]);

  if (!apiKeyLoaded) {
    return (
      <main className="auth-shell">
        <div className="auth-card">
          <div className="auth-card-header">
            <BrandLockup />
            <ThemeToggle theme={theme} onToggle={toggleTheme} />
          </div>
          <EmptyState title="Loading console" detail="Checking your stored session." />
        </div>
      </main>
    );
  }

  if (!account) {
    return <AuthView message={toast} onAccountApplied={applyAccount} onThemeToggle={toggleTheme} theme={theme} />;
  }

  return (
    <main className="console-shell">
      <aside className="sidebar">
        <BrandLockup />

        <nav className="nav-groups" aria-label="Console sections">
          {navSections.map((section) => (
            <div className="nav-group" key={section.label}>
              <p>{section.label}</p>
              {section.items.map((viewId) => {
                const item = views.find((candidate) => candidate.id === viewId);
                if (!item) {
                  return null;
                }

                return (
                  <button
                    aria-current={view === item.id ? "page" : undefined}
                    className={view === item.id ? "nav-item active" : "nav-item"}
                    key={item.id}
                    onClick={() => setView(item.id)}
                    type="button"
                  >
                    <span className="nav-icon">{item.short}</span>
                    <span>{item.label}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </nav>

        <div className="connection-card">
          <div>
            <p className="side-label">Workspace</p>
            <strong>{account?.tenant_name}</strong>
            <span>{account?.user_name}</span>
          </div>
          <StatusPill label={activeTeamId ? "Team scoped" : "Team needed"} tone={activeTeamId ? "good" : "idle"} />
        </div>
      </aside>

      <section className="workspace" aria-busy={loading}>
        <header className="topbar">
          <div className="page-title">
            <p className="eyebrow">{currentView.kicker}</p>
            <h2>{currentView.label}</h2>
            <span>{currentView.summary}</span>
          </div>
          <div className="topbar-actions">
            <StatusPill label={toast || "Live"} tone={toastTone} />
            <ThemeToggle theme={theme} onToggle={toggleTheme} />
            {isMetricView && (
              <label className="auto-refresh-control">
                <input checked={autoRefreshMetrics} onChange={(event) => applyAutoRefresh(event.target.checked)} type="checkbox" />
                <span>Auto refresh</span>
              </label>
            )}
            <span className="time-window">Last 15 minutes</span>
            <span className="avatar" aria-label={account.user_name || "Account"}>
              {account.user_name?.slice(0, 2).toUpperCase() || "AD"}
            </span>
            <button className="secondary compact-action" disabled={loading} onClick={() => run(refreshCurrentMetrics, "Metrics refreshed")} type="button">
              Refresh
            </button>
          </div>
        </header>

        {view === "overview" && (
          <OverviewView
            destinations={destinations}
            dlq={dlq}
            mappings={mappings}
            metrics={metrics}
            onNavigate={setView}
            topics={topics}
            unhealthyMappings={unhealthyMappings}
          />
        )}
        {view === "delivery" && (
          <DeliveryState deliveryFailures={deliveryFailures} mappings={mappings} topicHeads={topicHeads} unhealthyMappings={unhealthyMappings} />
        )}
        {view === "dlq" && (
          <DlqView
            destinations={destinations}
            dlq={dlq}
            mappings={mappings}
            onReplay={(payload) =>
              run(async () => {
                const isBulkReplay = !payload.dead_letter_event_id;
                if (isBulkReplay && !window.confirm("Replay up to 25 matching dead-letter events now?")) {
                  setToast("Replay cancelled");
                  return false;
                }
                const result = await postJson<ReplayResult>("/dead_letter_events", payload);
                await Promise.all([refreshDlq(), refreshCore()]);
                setToast(`Replayed ${result.replayed_count} DLQ event(s)`);
              }, "Replay complete")
            }
            onRefresh={() => run(refreshDlq, "DLQ refreshed")}
            setDestinationFilter={setDlqDestinationFilter}
            setTopicFilter={setDlqTopicFilter}
            topicFilter={dlqTopicFilter}
            destinationFilter={dlqDestinationFilter}
          />
        )}
        {view === "observability" && <Observability metrics={metrics} />}
        {view === "topics" && <TopicsView topics={topics} onDone={() => run(refreshCore, "Topics refreshed")} />}
        {view === "destinations" && <DestinationsView destinations={destinations} onDone={() => run(refreshCore, "Destinations refreshed")} />}
        {view === "mappings" && (
          <MappingsView
            destinations={destinations}
            mappings={mappings}
            onDone={() => run(refreshCore, "Mappings refreshed")}
            onNavigate={setView}
            topics={topics}
          />
        )}
        {view === "events" && (
          <EventsView
            autoRefresh={autoRefreshMetrics}
            events={events}
            isLive={isViewingLatestEvents}
            onNext={() => {
              setIsViewingLatestEvents(false);
              run(() => refreshEvents(events.cursor), "Events loaded");
            }}
            onRefreshLatest={() => {
              setIsViewingLatestEvents(true);
              run(() => refreshEvents(0), "Latest events loaded");
            }}
            onReset={() => {
              setIsViewingLatestEvents(true);
              run(() => refreshEvents(0), "Events reset");
            }}
            selectedTopic={selectedTopic}
            setSelectedTopic={setSelectedTopic}
            topics={eventTopics}
          />
        )}
        {view === "api-keys" && (
          <ApiKeysView
            account={account}
            activeTeamId={activeTeamId}
            onActiveTeamApplied={applyActiveTeam}
            onAccountApplied={applyAccount}
            onApiKeyApplied={applyIssuedApiKey}
            onDone={() => run(refreshCore, "Console refreshed")}
          />
        )}
      </section>
    </main>
  );
}

function BrandLockup() {
  return (
    <div className="brand-lockup">
      <span className="brand-mark">M</span>
      <div>
        <p className="eyebrow">Operator Console</p>
        <h1>Mycelo</h1>
      </div>
    </div>
  );
}

function ThemeToggle({ theme, onToggle }: { theme: Theme; onToggle: () => void }) {
  const isDark = theme === "dark";

  return (
    <button
      aria-label={`Switch to ${isDark ? "light" : "dark"} theme`}
      aria-pressed={isDark}
      className="theme-toggle"
      onClick={onToggle}
      type="button"
    >
      <span className="theme-toggle-track">
        <span className="theme-toggle-thumb" />
      </span>
      <span>{isDark ? "Dark" : "Light"}</span>
    </button>
  );
}

function AuthView({
  message,
  onThemeToggle,
  onAccountApplied,
  theme,
}: {
  message: string;
  onThemeToggle: () => void;
  onAccountApplied: (account: AccountContext) => void;
  theme: Theme;
}) {
  const [mode, setMode] = useState<"login" | "signup">("login");
  const [tenantName, setTenantName] = useState("");
  const [userName, setUserName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loginEmail, setLoginEmail] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [formMessage, setFormMessage] = useState(message);

  useEffect(() => {
    setFormMessage(message);
  }, [message]);

  async function signUp(event: FormEvent) {
    event.preventDefault();
    setFormMessage("");
    try {
      const response = await postJson<SignUpResponse>("/signup", {
        tenant_name: tenantName,
        user_name: userName,
        email,
        password,
      });
      onAccountApplied(response);
    } catch (error) {
      setFormMessage(error instanceof Error ? error.message : "Signup failed");
    }
  }

  async function login(event: FormEvent) {
    event.preventDefault();
    setFormMessage("");
    try {
      const response = await postJson<SignUpResponse>("/login", {
        email: loginEmail,
        password: loginPassword,
      });
      onAccountApplied(response);
    } catch (error) {
      setFormMessage(error instanceof Error ? error.message : "Login failed");
    }
  }

  return (
    <main className="auth-shell">
      <section className="auth-card">
        <div className="auth-card-header">
          <BrandLockup />
          <ThemeToggle theme={theme} onToggle={onThemeToggle} />
        </div>
        <div className="auth-heading">
          <p className="eyebrow">Access</p>
          <h2>{mode === "login" ? "Log in" : "Create account"}</h2>
          <span>{mode === "login" ? "Use your session to enter the operator console." : "Create a tenant and owner session."}</span>
        </div>
        <div className="segmented" role="tablist" aria-label="Authentication mode">
          <button className={mode === "login" ? "active" : ""} onClick={() => setMode("login")} type="button">
            Log in
          </button>
          <button className={mode === "signup" ? "active" : ""} onClick={() => setMode("signup")} type="button">
            Sign up
          </button>
        </div>
        {mode === "login" ? (
          <form className="form-panel" onSubmit={login}>
            <label>
              Email
              <input required type="email" value={loginEmail} onChange={(event) => setLoginEmail(event.target.value)} />
            </label>
            <label>
              Password
              <input required type="password" value={loginPassword} onChange={(event) => setLoginPassword(event.target.value)} />
            </label>
            <button className="primary" type="submit">Log in</button>
          </form>
        ) : (
          <form className="form-panel" onSubmit={signUp}>
            <label>
              Tenant name
              <input required value={tenantName} onChange={(event) => setTenantName(event.target.value)} />
            </label>
            <label>
              Your name
              <input required value={userName} onChange={(event) => setUserName(event.target.value)} />
            </label>
            <label>
              Email
              <input required type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
            </label>
            <label>
              Password
              <input required minLength={8} type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
            </label>
            <button className="primary" type="submit">Create account</button>
          </form>
        )}
        {formMessage && <div className="notice">{formMessage}</div>}
      </section>
    </main>
  );
}

function getStoredTheme(): Theme {
  if (typeof window === "undefined") {
    return "light";
  }

  const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
  if (stored === "dark" || stored === "light") {
    return stored;
  }

  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function applyTheme(theme: Theme) {
  if (typeof document !== "undefined") {
    document.documentElement.dataset.theme = theme;
  }

  if (typeof window !== "undefined") {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  }
}

function getStoredAutoRefresh() {
  if (typeof window === "undefined") {
    return true;
  }

  const stored = window.localStorage.getItem(AUTO_REFRESH_STORAGE_KEY);
  return stored === null ? true : stored === "true";
}

function setStoredAutoRefresh(enabled: boolean) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(AUTO_REFRESH_STORAGE_KEY, String(enabled));
}

function OverviewView(props: {
  topics: Topic[];
  destinations: Destination[];
  mappings: Mapping[];
  unhealthyMappings: Mapping[];
  dlq: DeadLetterEvent[];
  metrics: OutboundMetrics;
  onNavigate: (view: View) => void;
}) {
  const failureTotal = Object.values(props.metrics.delivery_failure_total ?? {}).reduce((sum, value) => sum + value, 0);
  const disabledCount = props.mappings.filter((mapping) => !mapping.delivery_flag).length;
  const dueNow = props.mappings.filter((mapping) => mapping.next_attempt_at > 0 && mapping.next_attempt_at <= Date.now()).length;
  const healthRows = props.unhealthyMappings.length ? props.unhealthyMappings : props.mappings;
  const setupItems: Array<{ label: string; value: number; view: View }> = [
    { label: "Topics", value: props.topics.length, view: "topics" },
    { label: "Destinations", value: props.destinations.length, view: "destinations" },
    { label: "Mappings", value: props.mappings.length, view: "mappings" },
  ];

  return (
    <div className="stack">
      <div className="metric-grid overview-metrics">
        <Metric label="Successful deliveries" value={props.metrics.delivery_success_total} tone="good" />
        <Metric label="Failures" value={failureTotal} tone={failureTotal ? "bad" : "idle"} />
        <Metric label="Needs attention" value={props.unhealthyMappings.length} tone={props.unhealthyMappings.length ? "bad" : "idle"} />
        <Metric label="DLQ records" value={props.dlq.length} />
      </div>

      <div className="overview-grid">
        <div className="panel panel-large">
          <div className="panel-header">
            <div>
              <h3>Delivery health</h3>
              <span>{dueNow} retry due now</span>
            </div>
            <button className="secondary" onClick={() => props.onNavigate("delivery")} type="button">
              Open delivery
            </button>
          </div>
          <MappingTable mappings={healthRows} limit={6} />
        </div>

        <div className="overview-side">
          <div className="panel control-plane-panel">
            <div className="panel-header">
              <div>
                <h3>Control plane</h3>
                <span>{disabledCount} disabled routes</span>
              </div>
            </div>
            <div className="setup-list">
              {setupItems.map((item) => (
                <button className="setup-row" key={item.label} onClick={() => props.onNavigate(item.view)} type="button">
                  <span>
                    <strong>{item.label}</strong>
                    <small>{item.value} configured</small>
                  </span>
                  <span className="setup-count">{item.value}</span>
                </button>
              ))}
            </div>
          </div>
          <div className="panel platform-panel">
            <div className="panel-header">
              <div>
                <h3>Platform status</h3>
                <span>Outbound runtime</span>
              </div>
            </div>
            <StatusPill label={failureTotal ? "degraded" : "all systems operational"} tone={failureTotal ? "bad" : "good"} />
            <small>{props.metrics.delivery_success_last_at ? `Updated ${formatTime(props.metrics.delivery_success_last_at)}` : "Waiting for first delivery signal."}</small>
          </div>
        </div>
      </div>

      <div className="split">
        <KeyValuePanel title="Failure categories" values={props.metrics.delivery_failure_total} />
        <div className="panel">
          <div className="panel-header">
            <div>
              <h3>Recent dead letters</h3>
              <span>{props.dlq.length} loaded</span>
            </div>
            <button className="secondary" onClick={() => props.onNavigate("dlq")} type="button">
              Open DLQ
            </button>
          </div>
          <div className="activity-list">
            {props.dlq.length ? (
              props.dlq.slice(0, 5).map((record) => (
                <p key={record.dead_letter_event_id}>
                  <span>
                    <strong>{record.destination_name}</strong>
                    <small>{record.topic_name}</small>
                  </span>
                  <StatusPill label={record.failure_category || "unknown"} tone="bad" />
                </p>
              ))
            ) : (
              <EmptyState title="No dead letters" detail="Failures replayed or no dead-letter records loaded." />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function DeliveryState({
  deliveryFailures,
  mappings,
  topicHeads,
  unhealthyMappings,
}: {
  deliveryFailures: DeliveryFailureEvent[];
  mappings: Mapping[];
  topicHeads: Record<string, number>;
  unhealthyMappings: Mapping[];
}) {
  const blockedCount = unhealthyMappings.length;
  const disabledCount = mappings.filter((mapping) => !mapping.delivery_flag).length;
  const dueNow = mappings.filter((mapping) => mapping.next_attempt_at > 0 && mapping.next_attempt_at <= Date.now()).length;
  const topicCounters = useMemo(() => topicLatestCounters(mappings, topicHeads), [mappings, topicHeads]);
  const topicRows = useMemo(() => uniqueTopics(mappings), [mappings]);
  const totalBacklog = mappings.reduce((sum, mapping) => sum + mappingBacklog(mapping, topicHeads), 0);
  const unorderedMappings = mappings.filter((mapping) => mapping.delivery_mode === "unordered");
  const parallelCap = unorderedMappings
    .filter((mapping) => mapping.delivery_flag)
    .reduce((sum, mapping) => sum + Math.max(0, mapping.unordered_max_in_flight || 0), 0);
  const activeInFlight = unorderedMappings.reduce((sum, mapping) => sum + (mapping.unordered_in_flight_count || 0), 0);
  const [rates, setRates] = useState<{
    sampleReady: boolean;
    topicPublishRate: number;
    mappingDeliveryRate: number;
    topicRates: Map<string, number>;
    mappingRates: Map<string, number>;
  }>({
    sampleReady: false,
    topicPublishRate: 0,
    mappingDeliveryRate: 0,
    topicRates: new Map(),
    mappingRates: new Map(),
  });
  const previousSample = useRef<{
    at: number;
    topics: Map<string, number>;
    mappings: Map<string, number>;
  } | null>(null);

  useEffect(() => {
    const now = Date.now();
    const topics = topicLatestCounters(mappings, topicHeads);
    const mappingRows = new Map(mappings.map((mapping) => [mappingKey(mapping), mappingDeliveryCounter(mapping)]));
    const previous = previousSample.current;

    if (previous && now > previous.at) {
      const elapsedSeconds = (now - previous.at) / 1000;
      let topicPublishRate = 0;
      let mappingDeliveryRate = 0;
      const nextTopicRates = new Map<string, number>();
      const nextMappingRates = new Map<string, number>();

      topics.forEach((currentLatestEventId, topicId) => {
        const previousLatestEventId = previous.topics.get(topicId);
        if (previousLatestEventId === undefined) {
          return;
        }
        const topicRate = Math.max(0, currentLatestEventId - previousLatestEventId) / elapsedSeconds;
        nextTopicRates.set(topicId, topicRate);
        topicPublishRate += topicRate;
      });

      mappingRows.forEach((currentDeliveryCounter, key) => {
        const previousDeliveryCounter = previous.mappings.get(key);
        if (previousDeliveryCounter === undefined) {
          return;
        }
        const mappingRate = Math.max(0, currentDeliveryCounter - previousDeliveryCounter) / elapsedSeconds;
        nextMappingRates.set(key, mappingRate);
        mappingDeliveryRate += mappingRate;
      });

      setRates({
        sampleReady: true,
        topicPublishRate,
        mappingDeliveryRate,
        topicRates: nextTopicRates,
        mappingRates: nextMappingRates,
      });
    }

    previousSample.current = { at: now, topics, mappings: mappingRows };
  }, [mappings, topicHeads]);

  return (
    <div className="stack">
      <div className="metric-grid">
        <Metric label="Mappings" value={mappings.length} />
        <Metric label="Total backlog" value={totalBacklog} tone={totalBacklog ? "bad" : "good"} />
        <Metric label="Topics" value={topicRows.length} />
        <Metric label="Topic pub / sec" value={rates.sampleReady ? formatRate(rates.topicPublishRate) : "sampling"} />
        <Metric label="Mapping del / sec" value={rates.sampleReady ? formatRate(rates.mappingDeliveryRate) : "sampling"} />
        <Metric label="Parallel cap" value={parallelCap || "ordered"} />
        <Metric label="In flight" value={activeInFlight} />
        <Metric label="Needs attention" value={blockedCount} tone={blockedCount ? "bad" : "good"} />
        <Metric label="Retry due now" value={dueNow} />
        <Metric label="Disabled" value={disabledCount} />
      </div>
      <div className="split">
        <div className="panel">
          <div className="panel-header">
            <div>
              <h3>Topic publish rates</h3>
              <span>One row per topic, not per mapping</span>
            </div>
          </div>
          <div className="activity-list">
            {topicRows.length ? (
              topicRows.map((topic) => (
                <p key={topic.topic_id}>
                  <span>
                    <strong>{topic.topic_name}</strong>
                    <small>latest event {topicCounters.get(topic.topic_id) || 0}</small>
                  </span>
                  <StatusPill label={rates.sampleReady ? formatRate(rates.topicRates.get(topic.topic_id) || 0) : "sampling"} tone="idle" />
                </p>
              ))
            ) : (
              <EmptyState title="No topic rates" detail="Create a topic and mapping to populate publish-rate rows." />
            )}
          </div>
        </div>
        <div className="panel">
          <div className="panel-header">
            <div>
              <h3>Mapping delivery rates</h3>
              <span>One row per destination-topic mapping</span>
            </div>
          </div>
          <div className="activity-list">
            {mappings.length ? (
              mappings.map((mapping) => (
                <p key={mappingKey(mapping)}>
                  <span>
                    <strong>{mapping.destination_name}</strong>
                    <small>
                      {mapping.topic_name} - delivered {mappingDeliveryCounter(mapping)}
                    </small>
                  </span>
                  <StatusPill label={rates.sampleReady ? formatRate(rates.mappingRates.get(mappingKey(mapping)) || 0) : "sampling"} tone="idle" />
                </p>
              ))
            ) : (
              <EmptyState title="No mapping rates" detail="Assign a topic to a destination to populate delivery-rate rows." />
            )}
          </div>
        </div>
      </div>
      <div className="panel">
        <div className="panel-header">
          <div>
            <h3>Failure queue</h3>
            <span>{unhealthyMappings.length || mappings.length} shown</span>
          </div>
          <button className="secondary compact-action" type="button">
            Filters
          </button>
        </div>
        <MappingTable
          mappings={unhealthyMappings.length ? unhealthyMappings : mappings}
          focusState
          mappingRates={rates.mappingRates}
          topicHeads={topicHeads}
          topicRates={rates.topicRates}
        />
      </div>
      <div className="panel">
        <div className="panel-header">
          <div>
            <h3>Recent delivery failures</h3>
            <span>{deliveryFailures.length} captured</span>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Mapping</th>
                <th>Failure</th>
                <th>Source</th>
                <th>Endpoint</th>
              </tr>
            </thead>
            <tbody>
              {deliveryFailures.length ? (
                deliveryFailures.map((record) => (
                  <tr key={record.delivery_failure_id}>
                    <td>#{record.delivery_failure_id}</td>
                    <td>
                      <strong>{record.destination_name}</strong>
                      <small>{record.topic_name}</small>
                    </td>
                    <td>
                      <StatusPill label={record.failure_category || "unknown"} tone="bad" />
                      <small>{record.failure_reason}</small>
                      <small>{record.failure_count} failed attempt(s)</small>
                    </td>
                    <td>
                      <span>event {record.source_event_id}</span>
                      <small>First {formatTime(record.first_failed_at)}</small>
                      <small>Last {formatTime(record.last_failed_at)}</small>
                    </td>
                    <td>
                      <code>{record.endpoint}</code>
                    </td>
                  </tr>
                ))
              ) : (
                <EmptyTableRow colSpan={5} title="No delivery failures" detail="Recent transport and endpoint failures will appear here." />
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function DlqView(props: {
  destinations: Destination[];
  dlq: DeadLetterEvent[];
  mappings: Mapping[];
  destinationFilter: string;
  topicFilter: string;
  setDestinationFilter: (value: string) => void;
  setTopicFilter: (value: string) => void;
  onRefresh: () => void;
  onReplay: (payload: Record<string, string | number>) => void;
}) {
  return (
    <div className="stack">
      <div className="toolbar">
        <select value={props.destinationFilter} onChange={(event) => props.setDestinationFilter(event.target.value)}>
          <option value="">All destinations</option>
          {props.destinations.map((destination) => (
            <option key={destination.destination_id} value={destination.destination_id}>
              {destination.destination_name}
            </option>
          ))}
        </select>
        <select value={props.topicFilter} onChange={(event) => props.setTopicFilter(event.target.value)}>
          <option value="">All topics</option>
          {uniqueTopics(props.mappings).map((topic) => (
            <option key={topic.topic_id} value={topic.topic_id}>
              {topic.topic_name}
            </option>
          ))}
        </select>
        <button className="secondary" onClick={props.onRefresh} type="button">
          Refresh DLQ
        </button>
        <button
          className="primary"
          onClick={() =>
            props.onReplay({
              destination_id: props.destinationFilter,
              topic_id: props.topicFilter,
              limit: 25,
              confirmation: "REPLAY_FILTERED_DLQ",
            })
          }
          type="button"
        >
          Replay filtered
        </button>
      </div>
      <div className="panel">
        <div className="panel-header">
          <h3>Dead-letter events</h3>
          <span>{props.dlq.length} records</span>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Mapping</th>
                <th>Failure</th>
                <th>Source</th>
                <th>Payload</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {props.dlq.length ? (
                props.dlq.map((record) => (
                  <tr key={record.dead_letter_event_id}>
                    <td>#{record.dead_letter_event_id}</td>
                    <td>
                      <strong>{record.destination_name}</strong>
                      <small>{record.topic_name}</small>
                    </td>
                    <td>
                      <StatusPill label={record.failure_category || "unknown"} tone="bad" />
                      <small>{record.failure_reason}</small>
                    </td>
                    <td>
                      <span>event {record.source_event_id}</span>
                      <small>{formatTime(record.dead_lettered_at)}</small>
                    </td>
                    <td>
                      <code>{safeJson(record.event_payload)}</code>
                    </td>
                    <td>
                      <button
                        className="secondary"
                        onClick={() => props.onReplay({ dead_letter_event_id: record.dead_letter_event_id })}
                        type="button"
                      >
                        Replay
                      </button>
                    </td>
                  </tr>
                ))
              ) : (
                <EmptyTableRow colSpan={6} title="No dead-letter events" detail="Try another filter or refresh after a failed delivery." />
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function Observability({ metrics }: { metrics: OutboundMetrics }) {
  const failureTotal = Object.values(metrics.delivery_failure_total ?? {}).reduce((sum, value) => sum + value, 0);

  return (
    <div className="stack">
      <div className="metric-grid">
        <Metric label="Successes" value={metrics.delivery_success_total} tone="good" />
        <Metric label="Failures" value={failureTotal} tone={failureTotal ? "bad" : "good"} />
        <Metric label="DLQ writes" value={metrics.dead_letter_write_total} />
        <Metric label="DLQ replays" value={metrics.dead_letter_replay_total} />
        <Metric label="Lag last" value={`${metrics.delivery_lag_ms.last}ms`} />
        <Metric label="Attempt last" value={`${metrics.delivery_attempt_duration_ms.last}ms`} />
      </div>
      <div className="split">
        <KeyValuePanel title="Failure categories" values={metrics.delivery_failure_total} />
        <KeyValuePanel title="Circuit opened" values={metrics.circuit_opened_total} />
        <KeyValuePanel title="Circuit blocked" values={metrics.circuit_blocked_total} />
        <div className="panel">
          <h3>Freshness</h3>
          <p className="large-time">{formatTime(metrics.delivery_success_last_at)}</p>
          <small>Most recent successful durable outbound delivery.</small>
        </div>
      </div>
    </div>
  );
}

function TopicsView({ topics, onDone }: { topics: Topic[]; onDone: () => void }) {
  const [topicName, setTopicName] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    await postJson<void>("/create_topic", { topic_name: topicName });
    setTopicName("");
    onDone();
  }

  return (
    <div className="split">
      <form className="panel form-panel" onSubmit={submit}>
        <h3>Create topic</h3>
        <label>
          Topic name
          <input required value={topicName} onChange={(event) => setTopicName(event.target.value)} />
        </label>
        <button className="primary" type="submit">
          Create
        </button>
      </form>
      <SimpleList title="Topics" rows={topics.map((topic) => [topic.topic_name, topic.topic_id])} />
    </div>
  );
}

function DestinationsView({ destinations, onDone }: { destinations: Destination[]; onDone: () => void }) {
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [editing, setEditing] = useState<Destination | null>(null);
  const [secret, setSecret] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (editing) {
      await postJson<void>("/update_destination", {
        id: editing.destination_id,
        destination_name: name,
        destination_address: address,
        webhook_signing_secret: secret || undefined,
      });
    } else {
      await postJson<void>("/create_destination", {
        destination_name: name,
        destination_address: address,
      });
    }
    setName("");
    setAddress("");
    setSecret("");
    setEditing(null);
    onDone();
  }

  return (
    <div className="stack">
      <form className="panel form-grid" onSubmit={submit}>
        <h3>{editing ? "Edit destination" : "Create destination"}</h3>
        <label>
          Name
          <input required value={name} onChange={(event) => setName(event.target.value)} />
        </label>
        <label>
          Endpoint URL
          <input required type="url" value={address} onChange={(event) => setAddress(event.target.value)} />
        </label>
        <label>
          Signing secret
          <input value={secret} onChange={(event) => setSecret(event.target.value)} placeholder="Leave blank unless changing" />
        </label>
        <button className="primary" type="submit">
          {editing ? "Save" : "Create"}
        </button>
      </form>
      <div className="panel table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Endpoint</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {destinations.length ? (
              destinations.map((destination) => (
                <tr key={destination.destination_id}>
                  <td>{destination.destination_name}</td>
                  <td>{destination.destination_address}</td>
                  <td><StatusPill label={destination.delivery_flag ? "enabled" : "disabled"} tone={destination.delivery_flag ? "good" : "idle"} /></td>
                  <td>
                    <button
                      className="secondary"
                      onClick={() => {
                        setEditing(destination);
                        setName(destination.destination_name);
                        setAddress(destination.destination_address);
                      }}
                      type="button"
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))
            ) : (
              <EmptyTableRow colSpan={4} title="No destinations" detail="Create a webhook endpoint before assigning topics." />
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function MappingsView(props: { destinations: Destination[]; topics: Topic[]; mappings: Mapping[]; onDone: () => void; onNavigate: (view: View) => void }) {
  const [destinationId, setDestinationId] = useState("");
  const [topicId, setTopicId] = useState("");
  const [deliveryMode, setDeliveryMode] = useState<"ordered" | "unordered">("ordered");
  const [unorderedMaxInFlight, setUnorderedMaxInFlight] = useState(32);

  async function assign(event: FormEvent) {
    event.preventDefault();
    await postJson<void>("/assign_topic_to_destination", {
      destination_id: destinationId,
      topic_id: topicId,
      delivery_mode: deliveryMode,
      unordered_max_in_flight: unorderedMaxInFlight,
    });
    props.onDone();
  }

  return (
    <div className="configure-screen">
      <div className="config-tabs" aria-label="Configure sections">
        <button className="active" type="button">Mappings</button>
        <button onClick={() => props.onNavigate("topics")} type="button">Topics</button>
        <button onClick={() => props.onNavigate("destinations")} type="button">Destinations</button>
      </div>
      <div className="mapping-workbench">
        <form className="panel form-panel mapping-form" onSubmit={assign}>
          <h3>Create / assign mapping</h3>
          <label>
            Destination
            <select required value={destinationId} onChange={(event) => setDestinationId(event.target.value)}>
              <option value="">Destination</option>
              {props.destinations.map((destination) => (
                <option key={destination.destination_id} value={destination.destination_id}>
                  {destination.destination_name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Topic
            <select required value={topicId} onChange={(event) => setTopicId(event.target.value)}>
              <option value="">Topic</option>
              {props.topics.map((topic) => (
                <option key={topic.topic_id} value={topic.topic_id}>
                  {topic.topic_name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Delivery mode
            <select value={deliveryMode} onChange={(event) => setDeliveryMode(event.target.value as "ordered" | "unordered")}>
              <option value="ordered">Ordered</option>
              <option value="unordered">Unordered parallel</option>
            </select>
          </label>
          <label>
            Parallel limit
            <input
              disabled={deliveryMode === "ordered"}
              min={1}
              max={256}
              type="number"
              value={unorderedMaxInFlight}
              onChange={(event) => setUnorderedMaxInFlight(Number(event.target.value))}
            />
          </label>
          <button className="primary full" type="submit">
            Assign mapping
          </button>
        </form>
        <MappingPolicyTable mappings={props.mappings} onDone={props.onDone} />
      </div>
    </div>
  );
}

function EventsView(props: {
  autoRefresh: boolean;
  topics: string[];
  selectedTopic: string;
  setSelectedTopic: (value: string) => void;
  events: EventsResponse;
  isLive: boolean;
  onNext: () => void;
  onRefreshLatest: () => void;
  onReset: () => void;
}) {
  return (
    <div className="stack">
      <div className="toolbar">
        <select value={props.selectedTopic} onChange={(event) => props.setSelectedTopic(event.target.value)}>
          <option value="">Select topic</option>
          {props.topics.map((topic) => (
            <option key={topic} value={topic}>
              {topic}
            </option>
          ))}
        </select>
        <button className="secondary" onClick={props.onReset} type="button">
          Reset cursor
        </button>
        <button className="secondary" onClick={props.onRefreshLatest} type="button">
          {props.isLive ? "Refresh latest" : "Resume live"}
        </button>
        <button className="primary" disabled={!props.events.has_more} onClick={props.onNext} type="button">
          Next page
        </button>
        <StatusPill
          label={props.autoRefresh ? (props.isLive ? "auto refresh on" : "history paused") : "auto refresh off"}
          tone={props.autoRefresh && props.isLive ? "good" : "idle"}
        />
      </div>
      <div className="panel table-wrap">
        <table>
          <thead>
            <tr>
              <th>Created</th>
              <th>Topic</th>
              <th>Payload</th>
            </tr>
          </thead>
          <tbody>
            {props.events.events.length ? (
              props.events.events.map((event, index) => (
                <tr key={`${event.created_at}-${index}`}>
                  <td>{formatTime(event.created_at)}</td>
                  <td>{event.topic}</td>
                  <td><code>{safeJson(event.event_data)}</code></td>
                </tr>
              ))
            ) : (
              <EmptyTableRow colSpan={3} title="No events loaded" detail="Choose a topic or publish an event to populate the log." />
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ApiKeysView(props: {
  account: AccountContext;
  activeTeamId: string;
  onActiveTeamApplied: (teamPublicId: string) => void;
  onAccountApplied: (account: AccountContext | null) => void;
  onApiKeyApplied: (apiKey: string) => void;
  onDone: () => void;
}) {
  const { account, activeTeamId, onActiveTeamApplied, onAccountApplied, onApiKeyApplied, onDone } = props;
  const [apiKey, setApiKey] = useState(() => getStoredApiKey());
  const [teamName, setTeamName] = useState("");
  const [teams, setTeams] = useState<Team[]>([]);
  const [message, setMessage] = useState("");

  const refreshTeams = useCallback(async () => {
    const rows = await getJson<Team[]>("/teams");
    setTeams(rows ?? []);
    if (!activeTeamId && rows?.[0]?.team_public_id) {
      onActiveTeamApplied(rows[0].team_public_id);
    }
  }, [activeTeamId, onActiveTeamApplied]);

  useEffect(() => {
    refreshTeams().catch((error) => setMessage(error instanceof Error ? error.message : "Failed to load teams"));
  }, [refreshTeams]);

  function applyIssuedKey(nextApiKey: string) {
    setApiKey(nextApiKey);
    onApiKeyApplied(nextApiKey);
  }

  async function createTeam(event: FormEvent) {
    event.preventDefault();
    setMessage("");
    try {
      const team = await postJson<Team>("/create_team", {
        team_name: teamName,
      });
      setTeamName("");
      setTeams((current) => [...current, team].sort((a, b) => a.team_name.localeCompare(b.team_name)));
      onActiveTeamApplied(team.team_public_id);
      setMessage("Team created");
      onDone();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Team creation failed");
    }
  }

  async function selectTeam(team: Team) {
    onActiveTeamApplied(team.team_public_id);
    setMessage(`${team.team_name} selected`);
    onDone();
  }

  async function createTeamApiKey(team: Team) {
    setMessage("");
    try {
      const response = await postJson<ApiKeyResponse>("/create_api_key", {
        team_public_id: team.team_public_id,
      });
      onActiveTeamApplied(team.team_public_id);
      applyIssuedKey(response.api_key);
      setMessage(`${team.team_name} request key created and stored for API requests`);
      onDone();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Request key creation failed");
    }
  }

  async function revokeKey() {
    setMessage("");
    try {
      await postJson<void>("/revoke_api_key");
      applyIssuedKey("");
      setMessage("Request key revoked");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Revocation failed");
    }
  }

  async function signOut() {
    try {
      await postJson<void>("/logout");
    } finally {
      onAccountApplied(null);
    }
  }

  return (
    <div className="split">
      <form className="panel form-panel" onSubmit={createTeam}>
        <h3>Create team</h3>
        <small>{account.tenant_name} / {account.user_name}</small>
        <label>
          Team name
          <input required value={teamName} onChange={(event) => setTeamName(event.target.value)} />
        </label>
        <button className="primary" type="submit">Create team</button>
        {apiKey && (
          <div className="secret-block">
            <small>Current browser request key</small>
            <code className="secret">{apiKey}</code>
          </div>
        )}
        {message && <small>{message}</small>}
      </form>
      <div className="panel">
        <div className="panel-header">
          <h3>Teams</h3>
          <span>{teams.length} teams</span>
        </div>
        <div className="team-list">
          {teams.length ? teams.map((team) => (
            <div className="team-row" key={team.team_public_id}>
              <span>
                <strong>{team.team_name}</strong>
                {activeTeamId === team.team_public_id && <small>Active team</small>}
              </span>
              <button className="secondary" onClick={() => selectTeam(team)} type="button">Use team</button>
              <button className="primary" onClick={() => createTeamApiKey(team)} type="button">Issue request key</button>
            </div>
          )) : <small>No teams yet.</small>}
        </div>
        <div className="button-row account-actions">
          <button className="primary danger" disabled={!activeTeamId} onClick={revokeKey} type="button">Revoke request key</button>
          <button className="secondary" onClick={signOut} type="button">Sign out</button>
        </div>
      </div>
    </div>
  );
}

function MappingPolicyTable({ mappings, onDone }: { mappings: Mapping[]; onDone: () => void }) {
  return (
    <div className="panel table-wrap">
      <div className="panel-header">
        <div>
          <h3>Mapping policy</h3>
          <span>{mappings.length} mappings</span>
        </div>
      </div>
      <table>
        <thead>
          <tr>
            <th>Mapping</th>
            <th>Retry</th>
            <th>Skip policy</th>
            <th>Delivery</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {mappings.length ? (
            mappings.map((mapping) => (
              <MappingPolicyRow key={`${mapping.destination_id}-${mapping.topic_id}`} mapping={mapping} onDone={onDone} />
            ))
          ) : (
            <EmptyTableRow colSpan={5} title="No mappings" detail="Assign a topic to a destination to configure policy." />
          )}
        </tbody>
      </table>
    </div>
  );
}

function MappingPolicyRow({ mapping, onDone }: { mapping: Mapping; onDone: () => void }) {
  const [base, setBase] = useState(mapping.retry_base_delay_ms);
  const [max, setMax] = useState(mapping.retry_max_delay_ms);
  const [failures, setFailures] = useState(mapping.max_consecutive_failures_before_skip);
  const [dlq, setDlq] = useState(mapping.dead_letter_queue_enabled);
  const [deliveryMode, setDeliveryMode] = useState<"ordered" | "unordered">(mapping.delivery_mode || "ordered");
  const [unorderedMaxInFlight, setUnorderedMaxInFlight] = useState(mapping.unordered_max_in_flight || 32);
  const modeLocked = mapping.delivery_flag;

  async function save() {
    await postJson<void>("/update_destination_topic_mapping_policy", {
      destination_id: mapping.destination_id,
      topic_id: mapping.topic_id,
      retry_base_delay_ms: base,
      retry_max_delay_ms: max,
      max_consecutive_failures_before_skip: failures,
      dead_letter_queue_enabled: dlq,
      skip_on_endpoint_4xx: mapping.skip_on_endpoint_4xx,
      skip_on_endpoint_5xx: mapping.skip_on_endpoint_5xx,
      skip_on_endpoint_transport_error: mapping.skip_on_endpoint_transport_error,
      skip_on_event_payload_error: mapping.skip_on_event_payload_error,
      delivery_mode: deliveryMode,
      unordered_max_in_flight: unorderedMaxInFlight,
    });
    onDone();
  }

  async function toggleDelivery() {
    await postJson<void>("/update_destination_delivery_flag", {
      destination_id: mapping.destination_id,
      topic_id: mapping.topic_id,
      delivery_flag: !mapping.delivery_flag,
    });
    onDone();
  }

  async function deleteMapping() {
    const confirmed = window.confirm(`Delete mapping from ${mapping.destination_name} to ${mapping.topic_name}?`);
    if (!confirmed) {
      return;
    }

    await postJson<void>("/delete_topic_for_destination", {
      destination_id: mapping.destination_id,
      topic_id: mapping.topic_id,
    });
    onDone();
  }

  return (
    <tr>
      <td>
        <strong>{mapping.destination_name}</strong>
        <small>{mapping.topic_name}</small>
        <small>{mapping.delivery_mode === "unordered" ? "Unordered parallel" : "Ordered"}</small>
      </td>
      <td className="inline-inputs">
        <input aria-label="Retry base milliseconds" type="number" min={1} value={base} onChange={(event) => setBase(Number(event.target.value))} />
        <input aria-label="Retry max milliseconds" type="number" min={1} value={max} onChange={(event) => setMax(Number(event.target.value))} />
      </td>
      <td className="inline-inputs">
        <input aria-label="Failure threshold" type="number" min={0} value={failures} onChange={(event) => setFailures(Number(event.target.value))} />
        <label className="check">
          <input checked={dlq} type="checkbox" onChange={(event) => setDlq(event.target.checked)} />
          DLQ
        </label>
      </td>
      <td>
        <StatusPill label={mapping.delivery_flag ? "enabled" : "disabled"} tone={mapping.delivery_flag ? "good" : "idle"} />
        <div className="mode-controls">
          <select
            aria-label="Delivery mode"
            disabled={modeLocked}
            value={deliveryMode}
            onChange={(event) => setDeliveryMode(event.target.value as "ordered" | "unordered")}
          >
            <option value="ordered">Ordered</option>
            <option value="unordered">Unordered parallel</option>
          </select>
          <input
            aria-label="Parallel delivery limit"
            disabled={modeLocked || deliveryMode === "ordered"}
            max={256}
            min={1}
            type="number"
            value={unorderedMaxInFlight}
            onChange={(event) => setUnorderedMaxInFlight(Number(event.target.value))}
          />
        </div>
        {modeLocked && <small>Pause to change mode</small>}
      </td>
      <td className="button-row">
        <button className="secondary" onClick={toggleDelivery} type="button">{mapping.delivery_flag ? "Disable" : "Enable"}</button>
        <button className="primary" onClick={save} type="button">Save</button>
        <button className="primary danger" onClick={deleteMapping} type="button">Delete</button>
      </td>
    </tr>
  );
}

function MappingTable({
  mappings,
  focusState,
  limit,
  topicRates,
  mappingRates,
  topicHeads,
}: {
  mappings: Mapping[];
  focusState?: boolean;
  limit?: number;
  topicRates?: Map<string, number>;
  mappingRates?: Map<string, number>;
  topicHeads?: Record<string, number>;
}) {
  const visibleMappings = limit ? mappings.slice(0, limit) : mappings;

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Mapping</th>
            <th>Cursor</th>
            <th>Failure</th>
            <th>Backoff</th>
            <th>Last activity</th>
            {focusState && <th>Endpoint</th>}
          </tr>
        </thead>
        <tbody>
          {visibleMappings.length ? (
            visibleMappings.map((mapping) => (
              <tr key={`${mapping.destination_id}-${mapping.topic_id}`}>
                <td>
                  <strong>{mapping.destination_name}</strong>
                  <small>{mapping.topic_name}</small>
                  <small>{mapping.delivery_mode === "unordered" ? `parallel x${mapping.unordered_max_in_flight}` : "ordered"}</small>
                  {topicRates && <small>topic pub {formatRate(topicRates.get(mapping.topic_id) || 0)}</small>}
                  {mappingRates && <small>mapping del {formatRate(mappingRates.get(mappingKey(mapping)) || 0)}</small>}
                </td>
                <td>
                  <span>delivered {mapping.last_delivered_event_id}</span>
                  <small>
                    {mapping.delivery_mode === "unordered"
                      ? `enqueued ${mapping.unordered_last_enqueued_event_id}`
                      : `attempted ${mapping.last_attempted_event_id}`}
                  </small>
                  <small>latest {mappingLatestEventId(mapping, topicHeads)}</small>
                  <small>{mappingBacklog(mapping, topicHeads)} backlog</small>
                </td>
                <td>
                  <StatusPill
                    label={mapping.last_error_category || (mapping.delivery_flag ? "healthy" : "disabled")}
                    tone={mapping.last_error_category ? "bad" : mapping.delivery_flag ? "good" : "idle"}
                  />
                  <small>{mapping.last_error || "No current error"}</small>
                </td>
                <td>
                  <span>{mapping.consecutive_failure_count} failures</span>
                  <small>
                    {mapping.delivery_mode === "unordered"
                      ? `${mapping.unordered_pending_count || 0} pending / ${mapping.unordered_in_flight_count || 0} in flight / ${mapping.unordered_failed_count || 0} failed`
                      : mapping.next_attempt_at
                        ? `next ${formatTime(mapping.next_attempt_at)}`
                        : "no backoff"}
                  </small>
                </td>
                <td>
                  <span>{formatTime(mapping.last_attempted_at)}</span>
                  <small>success {formatTime(mapping.last_succeeded_at)}</small>
                </td>
                {focusState && <td>{mapping.destination_address}</td>}
              </tr>
            ))
          ) : (
            <EmptyTableRow colSpan={focusState ? 6 : 5} title="No delivery mappings" detail="Create a destination-topic mapping to start delivery." />
          )}
        </tbody>
      </table>
    </div>
  );
}

function Metric({ label, value, tone = "idle" }: { label: string; value: number | string; tone?: "good" | "bad" | "idle" }) {
  return (
    <div className={`metric ${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function EmptyState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="empty-state">
      <strong>{title}</strong>
      <small>{detail}</small>
    </div>
  );
}

function EmptyTableRow({ colSpan, title, detail }: { colSpan: number; title: string; detail: string }) {
  return (
    <tr>
      <td colSpan={colSpan}>
        <EmptyState title={title} detail={detail} />
      </td>
    </tr>
  );
}

function StatusPill({ label, tone }: { label: string; tone: "good" | "bad" | "idle" }) {
  return <span className={`pill ${tone}`}>{label}</span>;
}

function KeyValuePanel({ title, values }: { title: string; values: Record<string, number> }) {
  const entries = Object.entries(values ?? {});
  return (
    <div className="panel">
      <h3>{title}</h3>
      <div className="kv-list">
        {entries.length ? entries.map(([key, value]) => (
          <p key={key}>
            <span>{key}</span>
            <strong>{value}</strong>
          </p>
        )) : <small>No values reported yet.</small>}
      </div>
    </div>
  );
}

function SimpleList({ title, rows }: { title: string; rows: string[][] }) {
  return (
    <div className="panel">
      <h3>{title}</h3>
      <div className="list">
        {rows.length ? (
          rows.map(([primary, secondary]) => (
            <p key={secondary}>
              <strong>{primary}</strong>
              <small>{secondary}</small>
            </p>
          ))
        ) : (
          <EmptyState title={`No ${title.toLowerCase()}`} detail="Create the first record to populate this section." />
        )}
      </div>
    </div>
  );
}

function uniqueTopics(mappings: Mapping[]) {
  const byId = new Map<string, { topic_id: string; topic_name: string }>();
  mappings.forEach((mapping) => byId.set(mapping.topic_id, { topic_id: mapping.topic_id, topic_name: mapping.topic_name }));
  return Array.from(byId.values());
}

function uniqueStrings(values: string[]) {
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort((a, b) => a.localeCompare(b));
}

function mappingKey(mapping: Mapping) {
  return `${mapping.destination_id}:${mapping.topic_id}`;
}

function mappingLatestEventId(mapping: Mapping, topicHeads?: Record<string, number>) {
  return Math.max(topicHeads?.[mapping.topic_id] || 0, mapping.latest_event_id || 0, mapping.last_delivered_event_id || 0);
}

function topicLatestCounters(mappings: Mapping[], topicHeads?: Record<string, number>) {
  const topics = new Map<string, number>();

  mappings.forEach((mapping) => {
    const latestEventId = mappingLatestEventId(mapping, topicHeads);
    topics.set(mapping.topic_id, Math.max(topics.get(mapping.topic_id) || 0, latestEventId));
  });

  return topics;
}

function mappingBacklog(mapping: Mapping, topicHeads?: Record<string, number>) {
  return Math.max(0, mappingLatestEventId(mapping, topicHeads) - mapping.last_delivered_event_id);
}

function mappingDeliveryCounter(mapping: Mapping) {
  if (mapping.delivery_mode === "unordered") {
    return mapping.unordered_delivered_count || 0;
  }

  return mapping.last_delivered_event_id || 0;
}

function formatRate(value: number) {
  if (!Number.isFinite(value) || value <= 0) {
    return "0/s";
  }

  const precision = value >= 10 ? 0 : 1;
  return `${value.toFixed(precision)}/s`;
}

function safeJson(value: unknown) {
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function formatTime(value: number) {
  if (!value) {
    return "never";
  }
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));
}
