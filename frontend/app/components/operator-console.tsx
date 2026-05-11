"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { getJson, postJson, query } from "../lib/api";
import type {
  ApiKeyResponse,
  DeadLetterEvent,
  Destination,
  EventsResponse,
  Mapping,
  OutboundMetrics,
  ReplayResult,
  Topic,
} from "../lib/types";

type View = "delivery" | "dlq" | "observability" | "topics" | "destinations" | "mappings" | "events" | "api-keys";

const views: Array<{ id: View; label: string }> = [
  { id: "delivery", label: "Delivery state" },
  { id: "dlq", label: "DLQ" },
  { id: "observability", label: "Observability" },
  { id: "topics", label: "Topics" },
  { id: "destinations", label: "Destinations" },
  { id: "mappings", label: "Mappings" },
  { id: "events", label: "Event log" },
  { id: "api-keys", label: "API keys" },
];

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
  const [view, setView] = useState<View>("delivery");
  const [topics, setTopics] = useState<Topic[]>([]);
  const [destinations, setDestinations] = useState<Destination[]>([]);
  const [mappings, setMappings] = useState<Mapping[]>([]);
  const [dlq, setDlq] = useState<DeadLetterEvent[]>([]);
  const [metrics, setMetrics] = useState<OutboundMetrics>(emptyMetrics);
  const [events, setEvents] = useState<EventsResponse>({ events: [], count: 0, cursor: 0, has_more: false });
  const [selectedTopic, setSelectedTopic] = useState("");
  const [isViewingLatestEvents, setIsViewingLatestEvents] = useState(true);
  const [dlqDestinationFilter, setDlqDestinationFilter] = useState("");
  const [dlqTopicFilter, setDlqTopicFilter] = useState("");
  const [toast, setToast] = useState("");
  const [loading, setLoading] = useState(false);

  const unhealthyMappings = useMemo(
    () =>
      mappings
        .filter((mapping) => mapping.last_error || mapping.consecutive_failure_count > 0 || mapping.next_attempt_at > Date.now())
        .sort((a, b) => b.consecutive_failure_count - a.consecutive_failure_count),
    [mappings],
  );

  const refreshCore = useCallback(async () => {
    const [topicRows, destinationRows, mappingRows, outboundMetrics] = await Promise.all([
      getJson<Topic[]>("/topics"),
      getJson<Destination[]>("/destinations"),
      getJson<Mapping[]>("/destination_topic_mappings"),
      getJson<OutboundMetrics>("/observability/outbound"),
    ]);
    setTopics(topicRows ?? []);
    setDestinations(destinationRows ?? []);
    setMappings(mappingRows ?? []);
    setMetrics(outboundMetrics ?? emptyMetrics);
    setSelectedTopic((current) => current || topicRows?.[0]?.topic_name || "");
  }, []);

  const refreshDlq = useCallback(async () => {
    const rows = await getJson<DeadLetterEvent[]>(
      `/dead_letter_events${query({
        destination_id: dlqDestinationFilter,
        topic_id: dlqTopicFilter,
        limit: 100,
      })}`,
    );
    setDlq(rows ?? []);
  }, [dlqDestinationFilter, dlqTopicFilter]);

  const refreshEvents = useCallback(async (nextCursor = 0) => {
    if (!selectedTopic) {
      setEvents({ events: [], count: 0, cursor: 0, has_more: false });
      return;
    }
    const response = await getJson<EventsResponse>(
      `/events${query({ topic: selectedTopic, offset: nextCursor, limit: 50, order: "desc" })}`,
    );
    setEvents(response);
  }, [selectedTopic]);

  const run = useCallback(async (action: () => Promise<void>, success: string) => {
    setLoading(true);
    setToast("");
    try {
      await action();
      setToast(success);
    } catch (error) {
      setToast(error instanceof Error ? error.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    run(refreshCore, "Console refreshed");
  }, [refreshCore, run]);

  useEffect(() => {
    refreshDlq().catch((error) => setToast(error instanceof Error ? error.message : "Failed to load DLQ"));
  }, [refreshDlq]);

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
    if (view !== "events" || !selectedTopic || !isViewingLatestEvents) {
      return;
    }

    const interval = window.setInterval(() => {
      refreshEvents(0).catch((error) => setToast(error instanceof Error ? error.message : "Failed to refresh events"));
    }, 3000);

    return () => window.clearInterval(interval);
  }, [isViewingLatestEvents, refreshEvents, selectedTopic, view]);

  return (
    <main className="console-shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Mycelo v0.0.4</p>
          <h1>Operator console</h1>
        </div>
        <nav className="nav-list" aria-label="Console sections">
          {views.map((item) => (
            <button
              className={view === item.id ? "nav-item active" : "nav-item"}
              key={item.id}
              onClick={() => setView(item.id)}
              type="button"
            >
              {item.label}
            </button>
          ))}
        </nav>
        <button className="primary full" disabled={loading} onClick={() => run(refreshCore, "Console refreshed")} type="button">
          Refresh
        </button>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">Realtime operations</p>
            <h2>{views.find((item) => item.id === view)?.label}</h2>
          </div>
          <StatusPill label={toast || "Ready"} tone={toast.toLowerCase().includes("failed") || toast.includes("error") ? "bad" : "good"} />
        </header>

        {view === "delivery" && <DeliveryState mappings={mappings} unhealthyMappings={unhealthyMappings} />}
        {view === "dlq" && (
          <DlqView
            destinations={destinations}
            dlq={dlq}
            mappings={mappings}
            onReplay={(payload) =>
              run(async () => {
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
          <MappingsView destinations={destinations} topics={topics} mappings={mappings} onDone={() => run(refreshCore, "Mappings refreshed")} />
        )}
        {view === "events" && (
          <EventsView
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
            topics={topics}
          />
        )}
        {view === "api-keys" && <ApiKeysView />}
      </section>
    </main>
  );
}

function DeliveryState({ mappings, unhealthyMappings }: { mappings: Mapping[]; unhealthyMappings: Mapping[] }) {
  const blockedCount = unhealthyMappings.length;
  const disabledCount = mappings.filter((mapping) => !mapping.delivery_flag).length;
  const dueNow = mappings.filter((mapping) => mapping.next_attempt_at > 0 && mapping.next_attempt_at <= Date.now()).length;

  return (
    <div className="stack">
      <div className="metric-grid">
        <Metric label="Mappings" value={mappings.length} />
        <Metric label="Needs attention" value={blockedCount} tone={blockedCount ? "bad" : "good"} />
        <Metric label="Retry due now" value={dueNow} />
        <Metric label="Disabled" value={disabledCount} />
      </div>
      <div className="panel">
        <div className="panel-header">
          <h3>Failure queue</h3>
          <span>{unhealthyMappings.length || mappings.length} shown</span>
        </div>
        <MappingTable mappings={unhealthyMappings.length ? unhealthyMappings : mappings} focusState />
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
              limit: 100,
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
              {props.dlq.map((record) => (
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
              ))}
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
            {destinations.map((destination) => (
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
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function MappingsView(props: { destinations: Destination[]; topics: Topic[]; mappings: Mapping[]; onDone: () => void }) {
  const [destinationId, setDestinationId] = useState("");
  const [topicId, setTopicId] = useState("");

  async function assign(event: FormEvent) {
    event.preventDefault();
    await postJson<void>("/assign_topic_to_destination", {
      destination_id: destinationId,
      topic_id: topicId,
    });
    props.onDone();
  }

  return (
    <div className="stack">
      <form className="panel form-grid compact" onSubmit={assign}>
        <h3>Assign topic</h3>
        <select required value={destinationId} onChange={(event) => setDestinationId(event.target.value)}>
          <option value="">Destination</option>
          {props.destinations.map((destination) => (
            <option key={destination.destination_id} value={destination.destination_id}>
              {destination.destination_name}
            </option>
          ))}
        </select>
        <select required value={topicId} onChange={(event) => setTopicId(event.target.value)}>
          <option value="">Topic</option>
          {props.topics.map((topic) => (
            <option key={topic.topic_id} value={topic.topic_id}>
              {topic.topic_name}
            </option>
          ))}
        </select>
        <button className="primary" type="submit">
          Assign
        </button>
      </form>
      <MappingPolicyTable mappings={props.mappings} onDone={props.onDone} />
    </div>
  );
}

function EventsView(props: {
  topics: Topic[];
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
            <option key={topic.topic_id} value={topic.topic_name}>
              {topic.topic_name}
            </option>
          ))}
        </select>
        <button className="secondary" onClick={props.onReset} type="button">
          Reset cursor
        </button>
        <button className="secondary" onClick={props.onRefreshLatest} type="button">
          Refresh latest
        </button>
        <button className="primary" disabled={!props.events.has_more} onClick={props.onNext} type="button">
          Next page
        </button>
        <StatusPill label={props.isLive ? "live refresh" : "history paused"} tone={props.isLive ? "good" : "idle"} />
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
            {props.events.events.map((event, index) => (
              <tr key={`${event.created_at}-${index}`}>
                <td>{formatTime(event.created_at)}</td>
                <td>{event.topic}</td>
                <td><code>{safeJson(event.event_data)}</code></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ApiKeysView() {
  const [apiKey, setApiKey] = useState("");
  const [tenantPublicId, setTenantPublicId] = useState("");
  const [teamPublicId, setTeamPublicId] = useState("");
  const [message, setMessage] = useState("");

  async function createKey() {
    const response = await postJson<ApiKeyResponse>("/create_api_key");
    setApiKey(response.api_key);
    setMessage("API key created");
  }

  async function rotateKey() {
    const response = await postJson<ApiKeyResponse>("/rotate_api_key");
    setApiKey(response.api_key);
    setMessage("API key rotated");
  }

  async function revokeKey(event: FormEvent) {
    event.preventDefault();
    await postJson<void>("/revoke_api_key", { tenant_public_id: tenantPublicId, team_public_id: teamPublicId });
    setMessage("API key revoked");
  }

  return (
    <div className="split">
      <div className="panel form-panel">
        <h3>Issue or rotate</h3>
        <div className="button-row">
          <button className="primary" onClick={createKey} type="button">Create</button>
          <button className="secondary" onClick={rotateKey} type="button">Rotate</button>
        </div>
        {apiKey && <code className="secret">{apiKey}</code>}
        {message && <small>{message}</small>}
      </div>
      <form className="panel form-panel" onSubmit={revokeKey}>
        <h3>Revoke</h3>
        <label>
          Tenant public id
          <input required value={tenantPublicId} onChange={(event) => setTenantPublicId(event.target.value)} />
        </label>
        <label>
          Team public id
          <input required value={teamPublicId} onChange={(event) => setTeamPublicId(event.target.value)} />
        </label>
        <button className="primary danger" type="submit">Revoke</button>
      </form>
    </div>
  );
}

function MappingPolicyTable({ mappings, onDone }: { mappings: Mapping[]; onDone: () => void }) {
  return (
    <div className="panel table-wrap">
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
          {mappings.map((mapping) => (
            <MappingPolicyRow key={`${mapping.destination_id}-${mapping.topic_id}`} mapping={mapping} onDone={onDone} />
          ))}
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

  return (
    <tr>
      <td>
        <strong>{mapping.destination_name}</strong>
        <small>{mapping.topic_name}</small>
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
      </td>
      <td className="button-row">
        <button className="secondary" onClick={toggleDelivery} type="button">{mapping.delivery_flag ? "Disable" : "Enable"}</button>
        <button className="primary" onClick={save} type="button">Save</button>
      </td>
    </tr>
  );
}

function MappingTable({ mappings, focusState }: { mappings: Mapping[]; focusState?: boolean }) {
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
          {mappings.map((mapping) => (
            <tr key={`${mapping.destination_id}-${mapping.topic_id}`}>
              <td>
                <strong>{mapping.destination_name}</strong>
                <small>{mapping.topic_name}</small>
              </td>
              <td>
                <span>delivered {mapping.last_delivered_event_id}</span>
                <small>attempted {mapping.last_attempted_event_id}</small>
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
                <small>{mapping.next_attempt_at ? `next ${formatTime(mapping.next_attempt_at)}` : "no backoff"}</small>
              </td>
              <td>
                <span>{formatTime(mapping.last_attempted_at)}</span>
                <small>success {formatTime(mapping.last_succeeded_at)}</small>
              </td>
              {focusState && <td>{mapping.destination_address}</td>}
            </tr>
          ))}
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
        {rows.map(([primary, secondary]) => (
          <p key={secondary}>
            <strong>{primary}</strong>
            <small>{secondary}</small>
          </p>
        ))}
      </div>
    </div>
  );
}

function uniqueTopics(mappings: Mapping[]) {
  const byId = new Map<string, { topic_id: string; topic_name: string }>();
  mappings.forEach((mapping) => byId.set(mapping.topic_id, { topic_id: mapping.topic_id, topic_name: mapping.topic_name }));
  return Array.from(byId.values());
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
