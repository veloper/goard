# WebSocket

Real-time change broadcasting. Connect and receive events when data changes.

**Endpoint:** `ws://<host>:<port>/api/ws?pat=pat_admin`

The PAT is passed as a query parameter because the browser WebSocket API cannot set custom headers.

## Connection

```javascript
const ws = new WebSocket('ws://localhost:8300/api/ws?pat=pat_admin');
ws.onmessage = (msg) => {
  const event = JSON.parse(msg.data);
  console.log(event.type, event.payload);
};
```

The server sends pings every 54 seconds. If the connection drops, reconnect with a short delay (3 seconds is standard).

## Events

All events are JSON with `type` and `payload` fields:

```typescript
{
  type: "issue_updated";
  payload: {
    id: number;
    changed: Record<string, { before: any; after: any }>;
  };
}
```

### Event types

| Type | Payload |
|------|---------|
| `project_created` | Full project object |
| `project_updated` | `{"id": ..., "changed": {...}}` |
| `project_deleted` | `{"id": ...}` |
| `issue_created` | Full issue object |
| `issue_updated` | `{"id": ..., "changed": {...}}` |
| `issue_deleted` | `{"id": ..., "project_id": ...}` |
| `comment_created` | Full comment object |

### Update format

Update events only include the fields that changed, not the full entity:

```json
{
  "type": "issue_updated",
  "payload": {
    "id": 1,
    "changed": {
      "state": {"before": "backlog", "after": "review"},
      "assignee": {"before": null, "after": 2}
    }
  }
}
```

Nullable fields (`assignee`, `parent_id`) show `null` when unset rather than `0` or `""`.

## Self-event suppression

Events caused by your own actions are not sent to your WebSocket connection. If you update an issue via the API, your WebSocket won't receive the `issue_updated` event — other connected clients will. This prevents double-handling (you already know what you did).
