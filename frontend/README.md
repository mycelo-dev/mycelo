# Mycelo Frontend

Operator console for Mycelo v0.0.4.

## Local development

```powershell
npm install
npm run dev
```

Set the backend base URL for the built-in proxy:

```powershell
$env:MYCELO_API_BASE_URL="http://localhost:3000"
```

The browser talks to `/api/mycelo/*`; the Next server proxies those requests to `MYCELO_API_BASE_URL`.
