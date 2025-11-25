# Webhooks

Event notification system via webhooks.

## Overview

Webhooks allow you to receive real-time notifications when events occur in the system.

**Status**: 🚧 Planned - Not yet implemented

**When available**, this system will support:
- Event subscriptions
- Webhook endpoints management
- Retry mechanisms
- Signature verification
- Event filtering
- Delivery tracking

## Planned Features

### Supported Events

```typescript
// Future events
type WebhookEvent =
  // Tenant events
  | 'tenant.created'
  | 'tenant.updated'
  | 'tenant.deleted'
  | 'tenant.suspended'
  | 'tenant.activated'
  
  // Member events
  | 'member.invited'
  | 'member.joined'
  | 'member.updated'
  | 'member.removed'
  | 'member.role_changed'
  
  // User events
  | 'user.created'
  | 'user.updated'
  | 'user.deleted'
  
  // RBAC events
  | 'role.created'
  | 'role.updated'
  | 'role.deleted'
  | 'policy.created'
  | 'policy.updated'
  | 'policy.deleted'
  | 'permission.created'
  
  // System User events
  | 'system_user.created'
  | 'system_user.rotated'
  | 'system_user.deactivated'
  | 'system_user.expiring_soon';
```

### Registering Webhooks

```javascript
// Future API
const webhook = await fetch('/api/v1/webhooks', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    url: 'https://your-app.com/webhooks/utm',
    events: ['tenant.created', 'member.joined'],
    description: 'Production webhook',
    secret: 'your-webhook-secret-key', // For signature verification
    active: true
  })
});

const webhookData = await webhook.json();
console.log('Webhook ID:', webhookData.data.id);
```

### Webhook Payload Format

```typescript
interface WebhookPayload {
  id: string;                    // Event ID
  type: WebhookEvent;           // Event type
  timestamp: string;            // ISO 8601 timestamp
  tenant_id?: string;           // Tenant ID (if applicable)
  data: Record<string, any>;   // Event-specific data
  metadata: {
    delivery_attempt: number;   // Delivery attempt count
    webhook_id: string;         // Webhook subscription ID
  };
}

// Example: tenant.created
{
  "id": "evt_01HQXYZ...",
  "type": "tenant.created",
  "timestamp": "2024-11-25T10:30:00Z",
  "tenant_id": "tenant-uuid",
  "data": {
    "tenant": {
      "id": "tenant-uuid",
      "name": "Acme Corp",
      "slug": "acme-corp",
      "created_by": "user-uuid",
      "created_at": "2024-11-25T10:30:00Z"
    }
  },
  "metadata": {
    "delivery_attempt": 1,
    "webhook_id": "webhook-uuid"
  }
}
```

### Receiving Webhooks

```javascript
// Express.js example
const express = require('express');
const crypto = require('crypto');

const app = express();

function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expectedSignature = 'sha256=' + hmac.digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

app.post('/webhooks/utm', express.raw({type: 'application/json'}), (req, res) => {
  const signature = req.headers['x-utm-signature'];
  const secret = process.env.WEBHOOK_SECRET;
  
  // Verify signature
  if (!verifyWebhookSignature(req.body.toString(), signature, secret)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Parse event
  const event = JSON.parse(req.body.toString());
  
  console.log(`Received event: ${event.type}`);
  
  // Handle event
  switch (event.type) {
    case 'tenant.created':
      handleTenantCreated(event.data.tenant);
      break;
    
    case 'member.joined':
      handleMemberJoined(event.data.member);
      break;
    
    // ... handle other events
  }
  
  // Acknowledge receipt
  res.status(200).send('OK');
});

function handleTenantCreated(tenant) {
  console.log('New tenant:', tenant.name);
  // Your business logic
}

function handleMemberJoined(member) {
  console.log('New member:', member.email);
  // Your business logic
}
```

## Current Workaround: Polling

Until webhooks are implemented, use polling:

```javascript
// Poll for new tenants
let lastChecked = new Date();

async function pollForNewTenants() {
  try {
    const response = await fetch(
      `/api/v1/tenants?created_after=${lastChecked.toISOString()}`,
      {credentials: 'include'}
    );
    
    const {data} = await response.json();
    
    if (data.data.length > 0) {
      console.log(`Found ${data.data.length} new tenants`);
      
      for (const tenant of data.data) {
        handleTenantCreated(tenant);
      }
      
      lastChecked = new Date();
    }
  } catch (error) {
    console.error('Polling failed:', error);
  }
}

// Poll every 30 seconds
setInterval(pollForNewTenants, 30000);
```

### Background Jobs Alternative

Use the Asynq job queue to process events:

```go
// internal/jobs/tasks/webhook_notifier.go
package tasks

import (
    "context"
    "encoding/json"
    "github.com/hibiken/asynq"
)

type WebhookPayload struct {
    URL   string                 `json:"url"`
    Event string                 `json:"event"`
    Data  map[string]interface{} `json:"data"`
}

func NewWebhookTask(payload WebhookPayload) (*asynq.Task, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    return asynq.NewTask("webhook:notify", data), nil
}

func HandleWebhookTask(ctx context.Context, t *asynq.Task) error {
    var p WebhookPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    
    // Send webhook
    // Implement retry logic here
    
    return nil
}
```

## Planned Management API

### List Webhooks

```javascript
const response = await fetch('/api/v1/webhooks', {
  credentials: 'include'
});

const webhooks = await response.json();
// Returns list of configured webhooks
```

### Update Webhook

```javascript
await fetch(`/api/v1/webhooks/${webhookId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    active: false,  // Disable webhook
    events: ['tenant.created']  // Update events
  })
});
```

### Webhook Delivery Log

```javascript
// Get delivery history for a webhook
const logs = await fetch(`/api/v1/webhooks/${webhookId}/deliveries`, {
  credentials: 'include'
});

// Example response
{
  "data": [
    {
      "id": "delivery-id",
      "event_id": "evt_01...",
      "event_type": "tenant.created",
      "status": "success",
      "status_code": 200,
      "attempts": 1,
      "delivered_at": "2024-11-25T10:30:05Z"
    },
    {
      "id": "delivery-id-2",
      "event_id": "evt_02...",
      "event_type": "member.joined",
      "status": "failed",
      "status_code": 500,
      "attempts": 3,
      "last_attempt_at": "2024-11-25T10:35:00Z",
      "error": "Connection timeout"
    }
  ]
}
```

### Retry Failed Delivery

```javascript
await fetch(`/api/v1/webhooks/deliveries/${deliveryId}/retry`, {
  method: 'POST',
  credentials: 'include'
});
```

## Security Best Practices

### 1. Verify Signatures

Always verify webhook signatures:

```python
import hmac
import hashlib

def verify_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = 'sha256=' + hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    return hmac.compare_digest(signature, expected)
```

### 2. Use HTTPS

Only accept webhooks over HTTPS in production

### 3. Implement Idempotency

Store processed event IDs to avoid duplicate processing:

```javascript
const processedEvents = new Set();

function handleWebhook(event) {
  if (processedEvents.has(event.id)) {
    console.log('Event already processed:', event.id);
    return;
  }
  
  // Process event
  processEvent(event);
  
  // Mark as processed
  processedEvents.add(event.id);
}
```

### 4. Rate Limiting

Implement rate limiting on your webhook endpoint

### 5. Timeout Handling

Set reasonable timeouts for webhook processing

## Contributing

Interested in implementing webhooks? See:
- [Backend Integration Guide](/guides/backend-integration)
- [Custom Middleware](/advanced/custom-middleware)

## Request for Implementation

If you need this feature, please:
1. Open a GitHub issue with your use case
2. Describe which events you need
3. Share your expected volume

**Implementation Priority**: Based on community demand

## Related Documentation

- [Permission Hooks](/advanced/permission-hooks) - Event hooks
- [Backend Integration](/guides/backend-integration) - Adding features
- [API Overview](/x-api/overview) - API reference
