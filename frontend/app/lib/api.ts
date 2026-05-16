import type { AccountContext } from "./types";

const API_PREFIX = "/api/mycelo";
const ACCOUNT_STORAGE_KEY = "mycelo_account";
const ACTIVE_TEAM_STORAGE_KEY = "mycelo_active_team_public_id";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function getJson<T>(path: string): Promise<T> {
  return requestJson<T>(path, { method: "GET" });
}

export async function postJson<T>(path: string, payload?: unknown): Promise<T> {
	return requestJson<T>(path, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: payload === undefined ? undefined : JSON.stringify(payload),
	});
}

async function requestJson<T>(path: string, init: RequestInit): Promise<T> {
	const headers = new Headers(init.headers);
	const apiKey = isStreamPath(path) ? getStoredApiKey() : "";
	if (apiKey) {
		headers.set("Authorization", `Bearer ${apiKey}`);
	}
	if (!isStreamPath(path)) {
		const activeTeamId = getStoredActiveTeamId();
		if (activeTeamId) {
			headers.set("X-Mycelo-Team", activeTeamId);
		}
	}

	const response = await fetch(`${API_PREFIX}${path}`, {
		...init,
		headers,
		cache: "no-store",
		credentials: "same-origin",
	});

  const text = await response.text();
  if (!response.ok) {
    throw new ApiError(response.status, text.trim() || response.statusText);
  }

  if (!text.trim()) {
    return undefined as T;
  }

	return JSON.parse(text) as T;
}

function isStreamPath(path: string) {
	const pathname = path.split("?")[0];
	return pathname === "/publish" || pathname === "/events";
}

export function getStoredApiKey() {
	if (typeof window === "undefined") {
		return "";
	}

	return window.localStorage.getItem("mycelo_api_key") ?? "";
}

export function setStoredApiKey(apiKey: string) {
	if (typeof window === "undefined") {
		return;
	}

	const trimmed = apiKey.trim();
	if (trimmed) {
		window.localStorage.setItem("mycelo_api_key", trimmed);
	} else {
		window.localStorage.removeItem("mycelo_api_key");
	}
}

export function getStoredActiveTeamId() {
	if (typeof window === "undefined") {
		return "";
	}

	return window.localStorage.getItem(ACTIVE_TEAM_STORAGE_KEY) ?? "";
}

export function setStoredActiveTeamId(teamPublicId: string) {
	if (typeof window === "undefined") {
		return;
	}

	const trimmed = teamPublicId.trim();
	if (trimmed) {
		window.localStorage.setItem(ACTIVE_TEAM_STORAGE_KEY, trimmed);
	} else {
		window.localStorage.removeItem(ACTIVE_TEAM_STORAGE_KEY);
	}
}

export function getStoredAccount(): AccountContext | null {
	if (typeof window === "undefined") {
		return null;
	}

	const raw = window.localStorage.getItem(ACCOUNT_STORAGE_KEY);
	if (!raw) {
		return null;
	}

	try {
		const account = JSON.parse(raw) as AccountContext;
		if (!account.tenant_public_id || !account.user_public_id) {
			window.localStorage.removeItem(ACCOUNT_STORAGE_KEY);
			return null;
		}

		const sanitized: AccountContext = {
			tenant_public_id: account.tenant_public_id,
			user_public_id: account.user_public_id,
			tenant_name: account.tenant_name,
			user_name: account.user_name,
			email: account.email,
			team_public_id: account.team_public_id,
			team_name: account.team_name,
		};
		window.localStorage.setItem(ACCOUNT_STORAGE_KEY, JSON.stringify(sanitized));
		return sanitized;
	} catch {
		window.localStorage.removeItem(ACCOUNT_STORAGE_KEY);
		return null;
	}
}

export function setStoredAccount(account: AccountContext | null) {
	if (typeof window === "undefined") {
		return;
	}

	if (account) {
		if (account.team_public_id) {
			setStoredActiveTeamId(account.team_public_id);
		}
		window.localStorage.setItem(ACCOUNT_STORAGE_KEY, JSON.stringify({
			tenant_public_id: account.tenant_public_id,
			user_public_id: account.user_public_id,
			tenant_name: account.tenant_name,
			user_name: account.user_name,
			email: account.email,
			team_public_id: account.team_public_id,
			team_name: account.team_name,
		}));
	} else {
		window.localStorage.removeItem(ACCOUNT_STORAGE_KEY);
		setStoredActiveTeamId("");
	}
}

export function query(params: Record<string, string | number | boolean | undefined>) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });

  const rendered = search.toString();
  return rendered ? `?${rendered}` : "";
}
