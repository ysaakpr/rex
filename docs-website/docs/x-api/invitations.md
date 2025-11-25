# Invitations API

API endpoints for managing user invitations to tenants.

## Overview

The Invitations system allows you to invite users to join your tenant. Users receive an email with a unique link to accept the invitation.

## Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/tenants/:id/invitations` | Create invitation | Tenant Member |
| GET | `/tenants/:id/invitations` | List invitations | Tenant Member |
| GET | `/invitations/:token` | Get invitation details | Public |
| POST | `/invitations/:token/accept` | Accept invitation | Authenticated |
| POST | `/invitations/check-pending` | Auto-accept pending | Authenticated |
| DELETE | `/invitations/:id` | Cancel invitation | Tenant Member |

## Create Invitation

Invite a user to join the tenant.

### Request

```http
POST /api/v1/tenants/:id/invitations
Content-Type: application/json
Authorization: Bearer <token>
```

**Body**:
```json
{
  "email": "newuser@example.com",
  "role_id": "writer-role-uuid"
}
```

### Response

**Status**: `201 Created`

```json
{
  "success": true,
  "message": "Invitation sent successfully",
  "data": {
    "id": "invitation-uuid",
    "tenant_id": "tenant-uuid",
    "email": "newuser@example.com",
    "role_id": "writer-role-uuid",
    "token": "unique-token-uuid",
    "status": "pending",
    "expires_at": "2024-11-23T10:00:00Z",
    "invited_by": "inviter-user-id",
    "created_at": "2024-11-20T10:00:00Z"
  }
}
```

### Behavior

1. Email sent to invited user
2. Token valid for 72 hours (default)
3. User can accept even if they don't have an account yet
4. Duplicate invitations prevented (same email + tenant)

## List Invitations

Get all invitations for a tenant.

### Request

```http
GET /api/v1/tenants/:id/invitations?page=1&page_size=20
Authorization: Bearer <token>
```

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "uuid-1",
        "email": "user1@example.com",
        "role_id": "writer-role-id",
        "status": "pending",
        "expires_at": "2024-11-23T10:00:00Z",
        "invited_by": "inviter-id",
        "created_at": "2024-11-20T10:00:00Z"
      },
      {
        "id": "uuid-2",
        "email": "user2@example.com",
        "role_id": "viewer-role-id",
        "status": "accepted",
        "accepted_at": "2024-11-21T10:00:00Z",
        "created_at": "2024-11-20T10:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total_count": 2,
    "total_pages": 1
  }
}
```

## Get Invitation (Public)

Get invitation details by token. **No authentication required**.

### Request

```http
GET /api/v1/invitations/:token
```

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "invitation-uuid",
    "tenant_id": "tenant-uuid",
    "tenant_name": "My Company",
    "email": "newuser@example.com",
    "role_id": "writer-role-uuid",
    "role_name": "Writer",
    "status": "pending",
    "expires_at": "2024-11-23T10:00:00Z",
    "invited_by": "inviter-user-id",
    "inviter_name": "John Doe"
  }
}
```

### Error Cases

**Already Accepted**:
```json
{
  "success": false,
  "error": "Invitation already accepted"
}
```

**Expired**:
```json
{
  "success": false,
  "error": "Invitation has expired"
}
```

**Cancelled**:
```json
{
  "success": false,
  "error": "Invitation has been cancelled"
}
```

## Accept Invitation

Accept an invitation and become a tenant member.

### Request

```http
POST /api/v1/invitations/:token/accept
Authorization: Bearer <token>
```

**User must be authenticated** (signed up or logged in).

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "member-uuid",
    "tenant_id": "tenant-uuid",
    "user_id": "user-id",
    "role_id": "writer-role-id",
    "status": "active",
    "joined_at": "2024-11-20T10:00:00Z"
  }
}
```

### Flow

```
1. User clicks invitation link
   ↓
2. Frontend checks if user logged in
   ↓
3. If not logged in → Sign up/Sign in
   ↓
4. POST /invitations/:token/accept
   ↓
5. User becomes tenant member
   ↓
6. Redirect to tenant dashboard
```

## Check Pending Invitations

Automatically accept any pending invitations for the current user's email.

### Request

```http
POST /api/v1/invitations/check-pending
Authorization: Bearer <token>
```

**Use Case**: Call this after user signs up to automatically accept any invitations sent before they created an account.

### Response

**Status**: `200 OK`

```json
{
  "success": true,
  "data": {
    "accepted_count": 2,
    "memberships": [
      {
        "id": "member-uuid-1",
        "tenant_id": "tenant-uuid-1",
        "user_id": "user-id",
        "role_id": "writer-role-id",
        "status": "active"
      },
      {
        "id": "member-uuid-2",
        "tenant_id": "tenant-uuid-2",
        "user_id": "user-id",
        "role_id": "admin-role-id",
        "status": "active"
      }
    ]
  }
}
```

## Cancel Invitation

Cancel a pending invitation.

### Request

```http
DELETE /api/v1/invitations/:id
Authorization: Bearer <token>
```

**Path Parameters**:
- `id`: Invitation UUID (not token!)

### Response

**Status**: `204 No Content`

### Behavior

- Only pending invitations can be cancelled
- User notified via email (optional)
- Token becomes invalid

## Examples

### Complete Invitation Flow (Frontend)

```javascript
// 1. Create invitation
async function inviteUser(tenantId, email, roleId) {
  const response = await fetch(`/api/v1/tenants/${tenantId}/invitations`, {
    method: 'POST',
    credentials: 'include',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email, role_id: roleId})
  });
  return await response.json();
}

// 2. User receives email, clicks link
// URL: https://yourdomain.com/invitations/:token/accept

// 3. Show invitation details (public endpoint)
async function getInvitationDetails(token) {
  const response = await fetch(`/api/v1/invitations/${token}`);
  return await response.json();
}

// 4. User signs up or logs in
// (handled by SuperTokens)

// 5. Accept invitation
async function acceptInvitation(token) {
  const response = await fetch(`/api/v1/invitations/${token}/accept`, {
    method: 'POST',
    credentials: 'include'
  });
  return await response.json();
}

// 6. Check for other pending invitations
async function checkPendingInvitations() {
  const response = await fetch('/api/v1/invitations/check-pending', {
    method: 'POST',
    credentials: 'include'
  });
  return await response.json();
}
```

## Configuration

### Invitation Expiry

Configure in `.env`:

```bash
INVITATION_EXPIRY_HOURS=72  # Default: 72 hours (3 days)
```

### Invitation URL

Configure base URL:

```bash
INVITATION_BASE_URL=https://yourdomain.com/invitations
```

Email will contain:
```
https://yourdomain.com/invitations/:token/accept
```

## Email Template

Default invitation email:

```
Subject: You've been invited to join [Tenant Name]

Hi,

[Inviter Name] has invited you to join [Tenant Name] as a [Role Name].

Click here to accept: [Invitation Link]

This invitation expires in 72 hours.
```

## Validation Rules

**Email**:
- Required
- Valid email format
- Case-insensitive
- Duplicate check per tenant

**Role ID**:
- Required
- Must be valid role UUID
- Role must exist in system

**Token**:
- UUID v4 format
- Unique per invitation
- Single-use (consumed on accept)

## Status Lifecycle

```
pending → accepted
   ↓
cancelled
   ↓
expired
```

- **pending**: Waiting for user to accept
- **accepted**: User accepted and became member
- **cancelled**: Admin cancelled invitation
- **expired**: Passed expiry time

## Next Steps

- **[Frontend: Invitation Flow](/frontend/invitation-flow)** - Implement in UI
- **[Member Management Guide](/guides/member-management)** - Managing team
- **[Members API](/x-api/members)** - Member endpoints
